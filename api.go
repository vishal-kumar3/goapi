package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

type APIServer struct {
	listenAddr string
	store      Storage
}

func NewAPIServer(listenAddr string, store Storage) *APIServer {
	return &APIServer{
		listenAddr: listenAddr,
		store:      store,
	}
}

func (s *APIServer) Run() {
	router := mux.NewRouter()

	router.HandleFunc("/account", makeHTTPHanleFunc(s.handleAccount))
	router.HandleFunc("/account/{id:[0-9]+}", makeHTTPHanleFunc(s.handleAccountByID))

	log.Println("JSON API server running on port: ", s.listenAddr)

	http.ListenAndServe(s.listenAddr, router)
}

func (s *APIServer) handleAccountByID(w http.ResponseWriter, r *http.Request) error {
	if r.Method == "GET" {
		return s.handleGetAccountByID(w, r)
	}
	if r.Method == "DELETE" {
		return s.handleDeleteAccount(w, r)
	}
	if r.Method == "PATCH" {
		return s.handleUpdateAccount(w, r)
	}

	return fmt.Errorf("method not allowed %s", r.Method)
}

func (s *APIServer) handleAccount(w http.ResponseWriter, r *http.Request) error {
	if r.Method == "GET" {
		return s.handleGetAllAccount(w, r)
	}
	if r.Method == "POST" {
		return s.handleCreateAccount(w, r)
	}

	return fmt.Errorf("method not allowed %s", r.Method)
}

func (s *APIServer) handleGetAccountByID(w http.ResponseWriter, r *http.Request) error {
	id, err := getID(r)
	if err != nil {
		return fmt.Errorf("invalid user id")
	}

	account, err := s.store.GetAccountByID(id)
	if err != nil {
		return err
	}
	return WriteJSON(w, http.StatusOK, account)
}

func (s *APIServer) handleUpdateAccount(w http.ResponseWriter, r *http.Request) error {
	id, err := getID(r)
	if err != nil {
		return err
	}

	updateAccountReq := &CreateAccountRequest{}
	if err := json.NewDecoder(r.Body).Decode(updateAccountReq); err != nil {
		return nil
	}

	account := NewAccount(updateAccountReq.FirstName, updateAccountReq.LastName)
	account.ID = id

	storeErr := s.store.UpdateAccount(account)
	if storeErr != nil {
		return storeErr
	}

	return WriteJSON(w, http.StatusNoContent, "")
}

func (s *APIServer) handleCreateAccount(w http.ResponseWriter, r *http.Request) error {
	createAccountReq := &CreateAccountRequest{}
	if err := json.NewDecoder(r.Body).Decode(createAccountReq); err != nil {
		return err
	}

	account := NewAccount(createAccountReq.FirstName, createAccountReq.LastName)
	err := s.store.CreateAccount(account)
	if err != nil {
		return err
	}

	return WriteJSON(w, http.StatusCreated, "User created successfully")
}

func (s *APIServer) handleDeleteAccount(w http.ResponseWriter, r *http.Request) error {
	log.Println("This is delete api")
	id, err := getID(r)
	if err != nil {
		return err
	}

	storeErr := s.store.DeleteAccount(id)
	if storeErr != nil {
		return fmt.Errorf("user not found")
	}

	return WriteJSON(w, http.StatusNoContent, "")
}

func (s *APIServer) handleTransfer(w http.ResponseWriter, r *http.Request) error {
	return nil
}

func (s *APIServer) handleGetAllAccount(w http.ResponseWriter, r *http.Request) error {
	accouts, err := s.store.GetAllAccount()
	if err != nil {
		return err
	}

	return WriteJSON(w, http.StatusOK, accouts)
}

type APIError struct {
	Error string `json:"error"`
}

func WriteJSON(w http.ResponseWriter, status int, v any) error {
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(status)

	return json.NewEncoder(w).Encode(v)
}

type apiFunc func(http.ResponseWriter, *http.Request) error

func makeHTTPHanleFunc(f apiFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := f(w, r); err != nil {
			WriteJSON(w, http.StatusBadRequest, APIError{Error: err.Error()})
		}
	}
}

func getID(r *http.Request) (int, error) {
	idStr := mux.Vars(r)["id"]

	id, err := strconv.Atoi(idStr)
	if err != nil {
		return 0, err
	}
	return id, nil
}
