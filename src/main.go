package main

import (
	"github.com/auth0/go-jwt-middleware"
	"github.com/dgrijalva/jwt-go"
	"github.com/gorilla/mux"
	"github.com/urfave/negroni"
	"log"
	"net/http"
	"net/url"
	"os"
	"shoppinglist-server/src/auth"
	"shoppinglist-server/src/credentials"
	"shoppinglist-server/src/logic"
	"shoppinglist-server/src/utils"
	"strings"
)

func handlerPlaceholder(w http.ResponseWriter, _ *http.Request) {
	_, _ = w.Write([]byte("Placeholder"))
}

func readEnv() string {
	port := os.Getenv("PORT")
	if port == "" {
		port = ":8080"
	}
	if !strings.HasPrefix(port, ":") {
		port = ":" + port
	}
	return port
}

func main() {
	port := readEnv()
	secretKey := []byte("SECRET KEY WILL BE HERE")

	credChecker, err := credentials.NewMongoDBCredentials("mongodb://localhost:27017", "shoppinglist", "users")
	if err != nil {
		log.Panicln(err)
	}
	log.Println("Connected to the database")
	err = logic.InitDB("mongodb://localhost:27017", "shoppinglist", "access", "lists")
	if err != nil {
		log.Panicln(err)
	}

	unauthenticatedRouter := mux.NewRouter()
	// Sign in
	unauthenticatedRouter.Handle("/v1/user/login", auth.NewLoginHandler(credChecker, secretKey))
	// Create new account
	unauthenticatedRouter.Handle("/v1/user/register", auth.NewRegistrationHandler(credChecker, secretKey))

	authenticatedRouter := mux.NewRouter()
	// Get list contents
	authenticatedRouter.Path("/v1/list/get").Methods("GET").HandlerFunc(logic.HandleGetList)
	// Create new list
	authenticatedRouter.Path("/v1/list/create").Methods("POST").HandlerFunc(logic.HandleCreateList)
	// Delete a list
	authenticatedRouter.Path("/v1/list/delete").Methods("POST").HandlerFunc(logic.HandleDeleteList)
	// Update list contents
	authenticatedRouter.Path("/v1/list/update").Methods("POST").HandlerFunc(logic.HandleUpdateList)
	// Share a list with another user
	authenticatedRouter.Path("/v1/list/share").Methods("POST").HandlerFunc(handlerPlaceholder)
	// Get all shared lists
	authenticatedRouter.Path("/v1/lists/shared").Methods("GET").HandlerFunc(handlerPlaceholder)
	// Get all available lists
	authenticatedRouter.Path("/v1/lists/available").Methods("GET").HandlerFunc(handlerPlaceholder)
	// Get all notifications
	authenticatedRouter.Path("/v1/notifications/get").Methods("GET").HandlerFunc(handlerPlaceholder)
	// Delete notification
	authenticatedRouter.Path("/v1/notifications/delete").Methods("POST").HandlerFunc(handlerPlaceholder)
	// Get all requests
	authenticatedRouter.Path("/v1/requests/get").Methods("GET").HandlerFunc(handlerPlaceholder)
	// Accept request to share a list
	authenticatedRouter.Path("/v1/requests/accept").Methods("POST").HandlerFunc(handlerPlaceholder)
	// Decline request to share a list
	authenticatedRouter.Path("/v1/requests/decline").Methods("POST").HandlerFunc(handlerPlaceholder)

	authMW := negroni.New()
	authMW.UseFunc(jwtmiddleware.New(jwtmiddleware.Options{
		ValidationKeyGetter: func(_ *jwt.Token) (interface{}, error) {
			return secretKey, nil
		},
		Extractor: func(r *http.Request) (string, error) {
			cookie, err := r.Cookie("jwt")
			if err != nil {
				return "", err
			}
			token, err := url.QueryUnescape(cookie.Value)
			if err != nil {
				return "", err
			}
			return token, nil
		},
		SigningMethod: jwt.SigningMethodHS256,
		ErrorHandler: func(w http.ResponseWriter, r *http.Request, err string) {
			w.WriteHeader(http.StatusUnauthorized)
			_, _ = w.Write(utils.NewWrappedError(err))
		},
	}).HandlerWithNext)
	authMW.UseHandler(authenticatedRouter)

	outerRouter := mux.NewRouter()
	outerRouter.PathPrefix("/v1/user/").Handler(unauthenticatedRouter)
	outerRouter.PathPrefix("/v1/").Handler(authMW)

	mainChain := negroni.New()
	mainChain.Use(negroni.NewLogger())
	mainChain.UseHandler(outerRouter)

	log.Println("Using port", port)

	err = http.ListenAndServe(port, mainChain)
	log.Println(err)

}
