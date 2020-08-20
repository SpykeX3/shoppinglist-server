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
	/*
		->
		GET example.com/v1/list/get?id=1gMzFPoiPWNywuRwYYrilF6RP2D

		<-
		{"error":"something went wrong"}
		or
		{"id":"1gMzFPoiPWNywuRwYYrilF6RP2D","owner":"katya","guests":["vasya"],"OriginalName":"Katya kishechka","last_changed":"2020-08-20T15:59:04.82Z","content":"100 proc"}
	*/
	authenticatedRouter.Path("/v1/list/get").Methods("GET").HandlerFunc(logic.HandleGetList)
	// Create new list
	/*
		->
		POST example.com/v1/list/create

		{"name":"New list name","content":"list goes here"}
		<-
		{"error":"something went wrong"}
		or
		{"id":"1gMzFPoiPWNywuRwYYrilF6RP2D"}
	*/
	authenticatedRouter.Path("/v1/list/create").Methods("POST").HandlerFunc(logic.HandleCreateList)
	// Delete a list
	/*
		->
		POST example.com/v1/list/delete?id=1gMzFPoiPWNywuRwYYrilF6RP2D

		<-
		{"error":"something went wrong"}
		or
		Status 200 and empty response
	*/
	authenticatedRouter.Path("/v1/list/delete").Methods("POST").HandlerFunc(logic.HandleDeleteList)
	// Update list contents
	/*
		->
		POST example.com/v1/list/update?id=1gMzFPoiPWNywuRwYYrilF6RP2D

		{"content":"updated content here"}
		<-
		{"error":"something went wrong"}
		or
		Status 200 and empty response
	*/
	authenticatedRouter.Path("/v1/list/update").Methods("POST").HandlerFunc(logic.HandleUpdateList)
	// Share a list with another user
	// Update list contents
	/*
		->
		POST example.com/v1/list/share

		{"id":"1gMzFPoiPWNywuRwYYrilF6RP2D","guest":"username of receiver"}
		<-
		{"error":"something went wrong"}
		or
		Status 200 and empty response
	*/
	authenticatedRouter.Path("/v1/list/share").Methods("POST").HandlerFunc(logic.HandleShareList)
	// Get all shared lists
	/*
		->
		GET example.com/v1/list/shared

		<-
		{"error":"something went wrong"}
		or
		[{"id":"1gMwLXlw92AZMcvAwyidItzOR29","display_name":"List1"},{"id":"Jn7wLXlw92A36cvAwyidItzOH65","display_name":"List2"}]
	*/
	authenticatedRouter.Path("/v1/lists/shared").Methods("GET").HandlerFunc(logic.HandleGetSharedLists)
	// Get all owned lists
	/*
		->
		GET example.com/v1/list/owned

		<-
		{"error":"something went wrong"}
		or
		[{"id":"1gMwLXlw92AZMcvAwyidItzOR29","display_name":"List1"},{"id":"Jn7wLXlw92A36cvAwyidItzOH65","display_name":"List2"}]
	*/
	authenticatedRouter.Path("/v1/lists/owned").Methods("GET").HandlerFunc(logic.HandleGetOwnedLists)
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
