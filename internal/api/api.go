package api

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/shynn2/cmd-gram/internal/models"
	"github.com/shynn2/cmd-gram/internal/storage"
	"github.com/shynn2/cmd-gram/pkg/client/postgresql"
	"github.com/shynn2/cmd-gram/pkg/utils"
	"github.com/sirupsen/logrus"
)

type api struct {
	config *Config
	logger *logrus.Logger
	r      *mux.Router
	db     storage.Storage
}

func New(router *mux.Router, config *Config) *api {
	return &api{
		r:      router,
		config: config,
		logger: logrus.New(),
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
	api.r.HandleFunc("/api/log-in", api.LogIn).Methods(http.MethodPost)
	api.r.HandleFunc("/api/new-chat", api.NewChat).Methods(http.MethodPost)
	api.r.HandleFunc("/api/all-chats/{user_id}", api.GetAllChats).Methods(http.MethodGet)
	api.r.HandleFunc("/api/open-chat/{chat_id}", api.OpenChat)
	api.r.HandleFunc("/api/send-message/{chat_id}", api.SendMassage)
}

func (api *api) handleHello(w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, "Hello")
}

func (api *api) SingIn(w http.ResponseWriter, r *http.Request) {
	var u = &models.UserDTO{}

	utils.ParseBody(r, u)

	id, err := api.db.CreateUser(context.Background(), u)
	if err != nil {
		api.logger.Errorf("cannot create user due to error: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Somthing went wrong"))
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(fmt.Sprintf("Successfully created user with id %d", id)))
}

func (api *api) LogIn(w http.ResponseWriter, r *http.Request) {
	var gotU = &models.UserDTO{}

	utils.ParseBody(r, gotU)

	u, err := api.db.FindByEmail(context.Background(), gotU.Email)
	if err != nil {
		api.logger.Errorf("cannot find user due to error: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(fmt.Sprintf("There is no user with email %s", u.Email)))
		return
	}

	if u.EncryptedPassword != gotU.EncryptedPassword {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Wrong password"))
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Sprintf("Hello %s", u.Email)))
}

func (api *api) NewChat(w http.ResponseWriter, r *http.Request) {
	var u []*models.User
	utils.ParseBody(r, &u)
	id, err := api.db.CreateChat(context.Background(), u)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("cant create chat"))
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(fmt.Sprintf("you created chat id=%d", id)))
}

func (api *api) GetAllChats(w http.ResponseWriter, r *http.Request) {
	uid, err := strconv.Atoi(mux.Vars(r)["chat_id"])
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("invalid argument"))
		return
	}

	chats, err := api.db.GetAllChats(context.Background(), uid)
	if err != nil {
		api.logger.Error(err)
		w.WriteHeader(http.StatusBadGateway)
		return
	}
	jsonChats, err := json.Marshal(chats)
	if err != nil {
		api.logger.Error(err)
		w.WriteHeader(http.StatusBadGateway)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(jsonChats)
}

func (api *api) OpenChat(w http.ResponseWriter, r *http.Request) {
	cid, err := strconv.Atoi(mux.Vars(r)["chat_id"])
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("invalid argument"))
		return
	}

	msgs, err := api.db.GetAllMessages(context.Background(), cid)
	if err != nil {
		api.logger.Error(err)
		return
	}

	jsonMsgs, err := json.Marshal(msgs)
	if err != nil {
		api.logger.Error(err)
		w.WriteHeader(http.StatusBadGateway)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(jsonMsgs)
}

func (api *api) SendMassage(w http.ResponseWriter, r *http.Request) {
	var msg *models.MessageDTO
	utils.ParseBody(r, &msg)
	_, err := api.db.CreateMessage(context.Background(), msg)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("cant send message"))
		return
	}

	w.WriteHeader(http.StatusCreated)
}
