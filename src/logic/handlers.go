package logic

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"shoppinglist-server/src/utils"
)

func handleGetList(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	context := r.Context()
	username := context.Value("username").(string)
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

func handleUpdateList(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	context := r.Context()
	username := context.Value("username").(string)
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
