package auth

import (
	"encoding/json"
	"github.com/dgrijalva/jwt-go"
	"net/http"
	"shoppinglist-server/src/credentials"
	slUtils "shoppinglist-server/src/utils"
)

type credentialInfo struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type loginHandler struct {
	cred   credentials.CredController
	secret []byte
}
type registrationHandler struct {
	cred   credentials.CredController
	secret []byte
}

func setJWTCookie(writer http.ResponseWriter, username string, secret []byte) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"username": username,
	})
	tokenString, _ := token.SignedString(secret)
	http.SetCookie(writer, &http.Cookie{
		Name:     "jwt",
		Value:    tokenString,
		Path:     "/",
		HttpOnly: true,
		SameSite: 1,
	})
}
func (r registrationHandler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	var creds credentialInfo
	err := json.NewDecoder(request.Body).Decode(&creds)
	if err != nil {
		writer.WriteHeader(http.StatusBadRequest)
		return
	}
	err = r.cred.Register(creds.Username, creds.Password)
	if err != nil {
		writer.WriteHeader(http.StatusForbidden)
		_, _ = writer.Write(slUtils.NewWrappedError("username is already taken"))
		return
	}
	// User has provided correct credentials and needs JWT to be set
	setJWTCookie(writer, creds.Username, r.secret)
}

func (l loginHandler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	var creds credentialInfo
	err := json.NewDecoder(request.Body).Decode(&creds)
	if err != nil {
		writer.WriteHeader(http.StatusBadRequest)
		return
	}
	err = l.cred.Login(creds.Username, creds.Password)
	if err != nil {
		writer.WriteHeader(http.StatusForbidden)
		_, _ = writer.Write(slUtils.NewWrappedError("invalid credentials"))
		return
	}
	// User has provided correct credentials and needs JWT to be set
	setJWTCookie(writer, creds.Username, l.secret)
}

func NewLoginHandler(cred credentials.CredController, secret []byte) http.Handler {
	return loginHandler{
		cred:   cred,
		secret: secret,
	}
}
func NewRegistrationHandler(cred credentials.CredController, secret []byte) http.Handler {
	return registrationHandler{
		cred:   cred,
		secret: secret,
	}
}
