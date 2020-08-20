package logic

import (
	"encoding/json"
	"github.com/dgrijalva/jwt-go"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
	"shoppinglist-server/src/utils"
)

func getUsername(r *http.Request) string {
	context := r.Context()
	claims := context.Value("user").(*jwt.Token).Claims.(jwt.MapClaims)
	return claims["username"].(string)
}

func internalError(w http.ResponseWriter, err error) {
	log.Errorln(err)
	w.WriteHeader(http.StatusInternalServerError)
	_, _ = w.Write(utils.WrapError(err))
}

func accessDenied(w http.ResponseWriter, username, id string) {
	log.Warningln("User", username, "tried to access list", id, "without permission")
	w.WriteHeader(http.StatusUnauthorized)
	_, _ = w.Write(utils.NewWrappedError("access denied"))
}

func HandleGetList(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	username := getUsername(r)
	authorized, err := hasAccessToList(username, id)
	if err != nil {
		internalError(w, err)
		return
	}
	if !authorized {
		_, _ = w.Write(utils.NewWrappedError("access denied"))
		return
	}
	listRec, err := getListById(id)
	if err != nil {
		internalError(w, err)
		return
	}
	result, err := json.Marshal(listRec)
	if err != nil {
		internalError(w, err)
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
		internalError(w, err)
		return
	}
	if !authorized {
		_, _ = w.Write(utils.NewWrappedError("access denied"))
		return
	}
	var reqContent requestListContent
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		internalError(w, err)
		return
	}
	err = json.Unmarshal(body, &reqContent)
	if err != nil {
		log.Warnln(err)
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write(utils.NewWrappedError("invalid request body"))
		return
	}
	err = updateList(id, reqContent.Content)
	if err != nil {
		internalError(w, err)
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
		internalError(w, err)
		return
	}
	err = json.Unmarshal(body, &reqNList)
	if err != nil {
		log.Warnln(err)
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write(utils.NewWrappedError("invalid request body"))
		return
	}
	id, err := createList(username, reqNList.Name, reqNList.Content)
	if err != nil {
		internalError(w, err)
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
		internalError(w, err)
		return
	}
	if !authorized {
		accessDenied(w, username, id)
		return
	}

	err = unlinkList(username, id)
	if err != nil {
		internalError(w, err)
		return
	}
	w.WriteHeader(http.StatusOK)
	return
}

func HandleGetSharedLists(w http.ResponseWriter, r *http.Request) {
	username := getUsername(r)
	lists, err := listSharedLists(username)
	if err != nil {
		internalError(w, err)
		return
	}
	result, err := json.Marshal(lists)
	if err != nil {
		internalError(w, err)
		return
	}
	_, _ = w.Write(result)
}

func HandleGetOwnedLists(w http.ResponseWriter, r *http.Request) {
	username := getUsername(r)
	lists, err := listOwnedLists(username)
	if err != nil {
		internalError(w, err)
		return
	}
	result, err := json.Marshal(lists)
	if err != nil {
		internalError(w, err)
		return
	}
	_, _ = w.Write(result)
}

type shareReq struct {
	Id    string `json:"id"`
	Guest string `json:"guest"`
}

func HandleShareList(w http.ResponseWriter, r *http.Request) {
	username := getUsername(r)
	var request shareReq
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		internalError(w, err)
		return
	}
	err = json.Unmarshal(body, &request)
	if err != nil {
		log.Warnln(err)
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write(utils.NewWrappedError("invalid request body"))
		return
	}
	err = addGuest(username, request.Guest, request.Id)
	if err != nil {
		internalError(w, err)
		return
	}
}
