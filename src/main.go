package main

import (
	"github.com/auth0/go-jwt-middleware"
	"github.com/dgrijalva/jwt-go"
	"github.com/gorilla/mux"
	"github.com/urfave/negroni"
	"log"
	"net/http"
	"os"
	"strings"
)

func getKey(_ *jwt.Token) (interface{}, error) {
	return []byte("SECRET_KEY"), nil
}

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

	unauthenticatedRouter := mux.NewRouter()
	// Sign in
	unauthenticatedRouter.HandleFunc("/v1/user/login", handlerPlaceholder)
	// Create new account
	unauthenticatedRouter.HandleFunc("/v1/user/register", handlerPlaceholder)

	authenticatedRouter := mux.NewRouter()
	// Get list contents
	authenticatedRouter.Path("/v1/list/id").Methods("GET").HandlerFunc(handlerPlaceholder)
	// Update list contents
	authenticatedRouter.Path("/v1/list/id").Methods("POST").HandlerFunc(handlerPlaceholder)
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
		ValidationKeyGetter: getKey,
		Extractor: func(r *http.Request) (string, error) {
			cookie, err := r.Cookie("jwt")
			if err != nil {
				return "", err
			}
			return cookie.Value, nil
		},
		SigningMethod: jwt.SigningMethodHS256,
	}).HandlerWithNext)
	authMW.UseHandler(authenticatedRouter)

	outerRouter := mux.NewRouter()
	outerRouter.PathPrefix("/v1/user/").Handler(unauthenticatedRouter)
	outerRouter.PathPrefix("/v1/").Handler(authMW)

	mainChain := negroni.New()
	mainChain.Use(negroni.NewLogger())
	mainChain.UseHandler(outerRouter)

	log.Println("Using port", port)

	err := http.ListenAndServe(port, mainChain)
	log.Println(err)

}
