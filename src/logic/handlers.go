package logic

import (
	"encoding/json"
	"github.com/dgrijalva/jwt-go"
	"io/ioutil"
	"net/http"
	"shoppinglist-server/src/utils"
)

func getUsername(r *http.Request) string {
	context := r.Context()
	claims := context.Value("user").(*jwt.Token).Claims.(jwt.MapClaims)
	return claims["username"].(string)
}

func HandleGetList(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	username := getUsername(r)
	authorized, err := hasAccessToList(username, id)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write(utils.NewWrappedError("internal server error"))
		return
	}
	if !authorized {
		_, _ = w.Write(utils.NewWrappedError("access denied"))
		return
	}
	listRec, err := getListById(id)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write(utils.NewWrappedError("internal server error"))
		return
	}
	result, err := json.Marshal(listRec)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write(utils.NewWrappedError("internal server error"))
		return
	}
	_, _ = w.Write(result)
}

type requestListContent struct {
	Content string `json:"content"`
}

func HandleUpdateList(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	username := getUsername(r)
	authorized, err := hasAccessToList(username, id)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write(utils.NewWrappedError("internal server error"))
		return
	}
	if !authorized {
		_, _ = w.Write(utils.NewWrappedError("access denied"))
		return
	}
	var reqContent requestListContent
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write(utils.NewWrappedError("internal server error"))
		return
	}
	err = json.Unmarshal(body, &reqContent)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write(utils.NewWrappedError("invalid request body"))
		return
	}
	err = updateList(id, reqContent.Content)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write(utils.NewWrappedError("internal server error"))
		return
	}
}

type requestNamedList struct {
	Name    string `json:"name"`
	Content string `json:"content"`
}
type idResp struct {
	Id string `json:"id"`
}

func HandleCreateList(w http.ResponseWriter, r *http.Request) {
	username := getUsername(r)
	var reqNList requestNamedList
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write(utils.NewWrappedError("internal server error"))
		return
	}
	err = json.Unmarshal(body, &reqNList)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write(utils.NewWrappedError("invalid request body"))
		return
	}
	id, err := createList(username, reqNList.Name, reqNList.Content)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write(utils.NewWrappedError("internal server error"))
		return
	}
	resp, _ := json.Marshal(idResp{Id: id})
	_, _ = w.Write(resp)
}

func HandleDeleteList(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	username := getUsername(r)
	authorized, err := hasAccessToList(username, id)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write(utils.NewWrappedError("internal server error"))
		return
	}
	if !authorized {
		_, _ = w.Write(utils.NewWrappedError("access denied"))
		return
	}

	err = deleteList(username, id)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write(utils.NewWrappedError("internal server error"))
		return
	}
	w.WriteHeader(http.StatusOK)
	return
}
