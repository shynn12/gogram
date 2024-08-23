package api

import (
	"context"
	"fmt"
	"io"
	"net/http"

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

func (api *api) Handle() {
	api.r.HandleFunc("/api/hello", api.handleHello)
	api.r.HandleFunc("/api/sign-in", api.SingIn).Methods(http.MethodPost)
	// api.r.HandleFunc("/api/LogIn", api.LogIn)
	// api.r.HandleFunc("/api/NewChat", api.NewChat)
	// api.r.HandleFunc("/api/OpenChat", api.Open)
	// api.r.HandleFunc("/api/SendMessage", api.SendMassage)
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
