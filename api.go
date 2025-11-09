package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/go-playground/validator/v10"
	jwt "github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/mux"
	"golang.org/x/crypto/bcrypt"
)

var secret string = "mysecretkey"

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

	router.HandleFunc("/login", makeHTTPHanleFunc(s.handleAccountLogin)).Methods("POST")
	router.HandleFunc("/account", makeHTTPHanleFunc(s.handleAccount))
	router.HandleFunc("/account/{id:[0-9]+}", withJWTAuth(makeHTTPHanleFunc(s.handleAccountByID), s.store))
	router.HandleFunc("/transfer", makeHTTPHanleFunc(s.handleTransfer))

	log.Println("JSON API server running on port", s.listenAddr)

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

	account := NewAccount(updateAccountReq.FirstName, updateAccountReq.LastName, nil)
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
	defer r.Body.Close()

	validate := validator.New()
	if err := validate.Struct(createAccountReq); err != nil {
		return err
	}

	account := NewAccount(createAccountReq.FirstName, createAccountReq.LastName, &createAccountReq.Password)
	id, err := s.store.CreateAccount(account)
	if err != nil {
		return err
	}
	account.ID = id

	token, err := createJWT(account)
	if err != nil {
		return err
	}

	res := struct {
		Account Account `json:"account"`
		Token   string  `json:"token"`
	}{
		Account: *account,
		Token:   token,
	}

	return WriteJSON(w, http.StatusCreated, res)
}

func (s *APIServer) handleAccountLogin(w http.ResponseWriter, r *http.Request) error {
	loginReq := &LoginRequest{}
	if err := json.NewDecoder(r.Body).Decode(loginReq); err != nil {
		return err
	}

	account, accountErr := s.store.GetAccountByNumber(int(loginReq.Number))
	if accountErr != nil {
		return accountErr
	}

	brcyptErr := bcrypt.CompareHashAndPassword([]byte(account.Password), []byte(loginReq.Password))

	if brcyptErr != nil {
		return fmt.Errorf("invalid login credentials")
	}

	token, tokenErr := createJWT(account)
	if tokenErr != nil {
		return tokenErr
	}

	res := struct {
		Account Account `json:"account"`
		Token   string  `json:"token"`
	}{
		Account: *account,
		Token:   token,
	}

	return WriteJSON(w, http.StatusOK, res)
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
	transferReq := &TransferRequest{}
	if err := json.NewDecoder(r.Body).Decode(transferReq); err != nil {
		return err
	}
	defer r.Body.Close()

	return WriteJSON(w, http.StatusOK, transferReq)
}

func (s *APIServer) handleGetAllAccount(w http.ResponseWriter, r *http.Request) error {
	accouts, err := s.store.GetAllAccount()
	if err != nil {
		return err
	}

	return WriteJSON(w, http.StatusOK, accouts)
}

func withJWTAuth(handlerFunc http.HandlerFunc, s Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("Called JWT Auth...")

		tokenStr := r.Header.Get("authorization")
		decodedToken, err := validateJWT(tokenStr)
		if err != nil {
			WriteJSON(w, http.StatusUnauthorized, APIError{Error: "User not authorized"})
			return
		}

		if !decodedToken.Valid {
			WriteJSON(w, http.StatusUnauthorized, APIError{Error: "User not authorized"})
			return
		}

		claims := decodedToken.Claims.(jwt.MapClaims)
		userID := int(claims["userID"].(float64))

		account, err := s.GetAccountByID(userID)
		if err != nil || account == nil {
			WriteJSON(w, http.StatusUnauthorized, APIError{Error: "User not authorized"})
			return
		}

		handlerFunc(w, r)
	}
}

func validateJWT(tokenStr string) (*jwt.Token, error) {
	return jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		return []byte(secret), nil
	})
}

func createJWT(acc *Account) (string, error) {
	claims := &jwt.MapClaims{
		"expiresAt":     15000,
		"accountNumber": acc.Number,
		"userID":        acc.ID,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

type APIError struct {
	Error string `json:"error"`
}

func WriteJSON(w http.ResponseWriter, status int, v any) error {
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(status)

	return json.NewEncoder(w).Encode(v)
}

func formatValidationErrors(err error) map[string]string {
	errors := make(map[string]string)

	for _, e := range err.(validator.ValidationErrors) {
		var msg string

		switch e.Tag() {
		case "required":
			msg = fmt.Sprintf("%s is required", e.Field())
		case "alpha":
			msg = fmt.Sprintf("%s must contain only letters", e.Field())
		case "strongpwd":
			msg = "Password must be at least 8 characters long and include an uppercase letter, number, and special symbol"
		default:
			msg = fmt.Sprintf("%s is invalid", e.Field())
		}

		errors[e.Field()] = msg
	}

	return errors
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
