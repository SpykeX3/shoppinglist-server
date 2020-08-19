package logic

import (
	"github.com/segmentio/ksuid"
	"time"
)

type list struct {
	Id           string    `bson:"id" json:"id"`
	OriginalName string    `bson:"name"`
	LastChanged  time.Time `bson:"last_changed" json:"last_changed"`
	Content      string    `bson:"content" json:"content"`
}

type listLink struct {
	Id          string `bson:"id" json:"id"`
	DisplayName string `bson:"display" json:"display_name"`
}

type accessRecord struct {
	Username       string     `bson:"username" json:"username"`
	AvailableLists []listLink `bson:"lists" json:"available_lists"`
}

func hasAccessToList(username, id string) (bool, error) {
	accessRec, err := getAccessByUsername(username)
	if err != nil {
		return false, err
	}
	for _, listLn := range accessRec.AvailableLists {
		if listLn.Id == id {
			return true, nil
		}
	}
	return false, nil
}

func createList(username, name, content string) (string, error) {
	id := ksuid.New().String()
	newList := list{
		Id:           id,
		OriginalName: name,
		LastChanged:  time.Now(),
		Content:      content,
	}
	err := insertListInDB(newList)
	if err != nil {
		return "", err
	}
	err = addToAccessLists(username, id)
	if err != nil {
		return "", err
	}
	return id, nil
}

func deleteList(username, id string) error {
	return removeFromAccessLists(username, id)
}

func InitNewUser(username string) error {
	return insertAccessRecord(accessRecord{
		Username:       username,
		AvailableLists: make([]listLink, 0),
	})
}
