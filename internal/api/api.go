package api

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/shynn2/cmd-gram/internal/models"
	"github.com/shynn2/cmd-gram/internal/storage"
	"github.com/shynn2/cmd-gram/pkg/client/postgresql"
	"github.com/shynn2/cmd-gram/pkg/utils"
	"github.com/sirupsen/logrus"
)

type api struct {
	config      *Config
	logger      *logrus.Logger
	r           *mux.Router
	db          storage.Storage
	chatClients map[int]map[*websocket.Conn]bool //map with int define room where clients are, map with conn define who is present
}

var upgreader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func New(router *mux.Router, config *Config) *api {
	return &api{
		r:           router,
		config:      config,
		logger:      logrus.New(),
		chatClients: make(map[int]map[*websocket.Conn]bool),
	}
}

func (api *api) Start() error {
	if err := api.configureLogger(); err != nil {
		return err
	}

	api.Handle()

	if err := api.configureStore(); err != nil {
		return err
	}
	api.logger.Info("store was connected successfully")

	api.logger.Info("starting api server")
	return http.ListenAndServe(api.config.BindAddr, api.r)
}

func (api *api) configureLogger() error {
	level, err := logrus.ParseLevel(api.config.LogLevel)
	if err != nil {
		return err
	}
	api.logger.SetLevel(level)

	return nil
}

func (api *api) configureStore() error {
	db, err := postgresql.NewClient(api.config.Store)
	if err != nil {
		return err
	}

	api.db = storage.New(db)

	return nil
}

func (api *api) Handle() {
	api.r.HandleFunc("/api/hello", api.handleHello)
	api.r.HandleFunc("/api/sign-in", api.SingIn).Methods(http.MethodPost)
	api.r.HandleFunc("/api/login", api.LogIn).Methods(http.MethodPost)
	api.r.HandleFunc("/api/new-chat", api.NewChat).Methods(http.MethodPost)
	api.r.HandleFunc("/api/{user_id}/chats", api.GetAllChats).Methods(http.MethodGet)
	api.r.HandleFunc("/api/{user_id}/chats/{chat_id}", api.OpenChat)
}

func (api *api) handleHello(w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, "Hello")
}

func (api *api) SingIn(w http.ResponseWriter, r *http.Request) {
	var u = &models.UserDTO{}
	var status int
	var res any

	utils.ParseBody(r, u)

	id, err := api.db.CreateUser(context.Background(), u)
	if err != nil {
		api.logger.Errorf("cannot create user due to error: %v", err)
		status = http.StatusBadRequest
		res = models.Error{Text: fmt.Sprintf("cannot create user due to error: %v", err)}
		writeResponse(w, status, res)
		return
	}

	status = http.StatusCreated
	res = models.User{ID: id, Email: u.Email, EncryptedPassword: u.EncryptedPassword}
	writeResponse(w, status, res)
}

func (api *api) LogIn(w http.ResponseWriter, r *http.Request) {
	var gotU = &models.UserDTO{}
	utils.ParseBody(r, gotU)
	u, err := api.db.FindByEmail(context.Background(), gotU.Email)
	if err != nil {
		api.logger.Errorf("cannot find user due to error: %v", err)
		status := http.StatusBadRequest
		res := models.Error{Text: fmt.Sprintf("There is no user with email %s", gotU.Email)}
		writeResponse(w, status, res)
		return
	}
	if u.EncryptedPassword != gotU.EncryptedPassword {
		status := http.StatusBadRequest
		res := &models.Error{Text: "Wrong password"}
		writeResponse(w, status, res)
		return
	}

	status := http.StatusOK
	res := u
	writeResponse(w, status, res)
}

func (api *api) NewChat(w http.ResponseWriter, r *http.Request) {
	var udto []*models.UserDTO
	var us []*models.User

	utils.ParseBody(r, &udto)
	for _, v := range udto {
		u, err := api.db.FindByEmail(context.Background(), v.Email)
		if err != nil {
			status := http.StatusBadRequest
			res := &models.Error{Text: fmt.Sprintf("cant find user: %s", v.Email)}
			writeResponse(w, status, res)
			return
		}
		us = append(us, u)
	}
	id, err := api.db.CreateChat(context.Background(), us)
	if err != nil {
		status := http.StatusBadRequest
		res := &models.Error{Text: "cant create chat"}
		writeResponse(w, status, res)
		return
	}

	status := http.StatusCreated
	res := &models.Chat{ID: id, Name: fmt.Sprintf("%s-%s", us[0].Email, us[1].Email)}
	writeResponse(w, status, res)
}

func (api *api) GetAllChats(w http.ResponseWriter, r *http.Request) {
	uid, err := strconv.Atoi(mux.Vars(r)["user_id"])
	if err != nil {
		status := http.StatusBadRequest
		res := &models.Error{Text: "invalid argument"}
		writeResponse(w, status, res)
		return
	}

	chats, err := api.db.GetAllChats(context.Background(), uid)
	if err != nil {
		api.logger.Error(err)
		status := http.StatusBadGateway
		res := &models.Error{Text: fmt.Sprintf("cant get chats due to err: %v", err)}
		writeResponse(w, status, res)
		return
	}
	status := http.StatusOK
	writeResponse(w, status, chats)

}

func (api *api) OpenChat(w http.ResponseWriter, r *http.Request) {
	ws, err := upgreader.Upgrade(w, r, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer ws.Close()

	cid, err := strconv.Atoi(mux.Vars(r)["chat_id"])
	if err != nil {
		api.logger.Error(err)
		msg := &models.Error{Text: "invalid argument"}
		ws.WriteJSON(msg)
		return
	}

	var broadcast = make(chan *models.MessageDTO)

	if api.chatClients[cid] == nil {
		api.chatClients[cid] = make(map[*websocket.Conn]bool)
	}
	api.chatClients[cid][ws] = true //users connected to the room

	msgs, err := api.db.GetAllMessages(context.Background(), cid)
	if err != nil {
		api.logger.Error(err)
		msg := &models.Error{Text: fmt.Sprintf("cant get messages due to error: %v", err)}
		ws.WriteJSON(msg)
		return
	}

	if err := ws.WriteJSON(msgs); err != nil {
		api.logger.Panic(err)
	}
	go func() {
		for {
			msg := <-broadcast
			for client := range api.chatClients[cid] {
				if ws != client {
					err := client.WriteJSON(*msg)
					if err != nil {
						fmt.Println(err)
						client.Close()
						delete(api.chatClients[cid], client)
					}
				}
			}
		}
	}()
	for {
		var msg *models.MessageDTO
		err := ws.ReadJSON(&msg)
		if err != nil {
			fmt.Println(err)
			delete(api.chatClients[cid], ws)
			return
		}

		_, err = api.db.CreateMessage(context.Background(), msg)
		if err != nil {
			log.Println(err)
			res := &models.Error{Text: fmt.Sprintf("cant create messages due to error: %v", err)}
			ws.WriteJSON(res)
			return
		}

		broadcast <- msg
	}

}

// 	ch := make(chan *models.MessageDTO)
// 	if _, ok := api.chatChan[cid]; !ok {
// 		api.chatChan[cid] = make([]chan *models.MessageDTO, 2)
// 		api.chatChan[cid][0] = ch
// 	} else {
// 		api.chatChan[cid][1] = ch
// 	}

// 	go func() {
// 		for {
// 			for {
// 				nt := time.NewTimer(2 * time.Minute)
// 				go func() {
// 					<-newtimer.C

// 					fmt.Println("No one connected")
// 				}()
// 				if api.chatChan[cid][1] == nil {
// 					break
// 				}
// 			}
// 			msg := <-ch
// 			if msg.Body == "/exit" {
// 				close(ch)
// 				return
// 			}
// 			res, _ := json.Marshal(msg)
// 			ws.WriteJSON(res)
// 		}
// 	}()
// 	for {
// 		if err := ws.ReadJSON(&msg); err != nil {
// 			log.Println(err)
// 			return
// 		}
//

// 		go func(chan *models.MessageDTO) {
// 			ch <- msg
// 		}(ch)
// 	}
// }

func writeResponse(w http.ResponseWriter, status int, res any) error {
	log.Println(res, status)
	w.WriteHeader(status)
	w.Header().Set("Content-Type", "application/json")
	out, err := json.Marshal(res)
	if err != nil {
		return err
	}
	w.Write(out)
	return nil
}
