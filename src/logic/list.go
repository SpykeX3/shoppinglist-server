package logic

import (
	"errors"
	"github.com/segmentio/ksuid"
	log "github.com/sirupsen/logrus"
	"time"
)

type list struct {
	Id           string    `bson:"id" json:"id"`
	Owner        string    `bson:"owner" json:"owner"`
	Guests       []string  `bson:"guests" json:"guests"`
	OriginalName string    `bson:"name"`
	LastChanged  time.Time `bson:"last_changed" json:"last_changed"`
	Content      string    `bson:"content" json:"content"`
}

type listLink struct {
	Id          string `bson:"id" json:"id"`
	DisplayName string `bson:"display" json:"display_name"`
}

type accessRecord struct {
	Username    string     `bson:"username" json:"username"`
	OwnedLists  []listLink `bson:"owned" json:"owned_lists"`
	SharedLists []listLink `bson:"shared" json:"shared_lists"`
}

func hasAccessToList(username, id string) (bool, error) {
	accessRec, err := getAccessByUsername(username)
	if err != nil {
		return false, err
	}
	for _, listLn := range accessRec.OwnedLists {
		if listLn.Id == id {
			return true, nil
		}
	}
	for _, listLn := range accessRec.SharedLists {
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
		Owner:        username,
		Guests:       make([]string, 0),
		OriginalName: name,
		LastChanged:  time.Now(),
		Content:      content,
	}
	err := insertListInDB(newList)
	if err != nil {
		return "", err
	}
	err = addToAccessListsOwned(username, listLink{
		Id:          id,
		DisplayName: name,
	})
	if err != nil {
		return "", err
	}
	return id, nil
}

func unlinkList(username, id string) error {
	accessRec, err := getAccessByUsername(username)
	if err != nil {
		return err
	}
	for _, listLn := range accessRec.OwnedLists {
		if listLn.Id == id {
			err = deleteList(id)
			if err != nil {
				log.Error("Owner \"", username, "\" has unlinked list ", id, "but it is not removed:", err)
			}
			err := removeFromAccessListsOwned(username, id)
			if err != nil {
				return err
			}
			return nil
		}
	}
	for _, listLn := range accessRec.SharedLists {
		if listLn.Id == id {
			err := removeFromAccessListsShared(username, id)
			return err
		}
	}
	return errors.New("")
}

func deleteList(id string) error {
	list, err := getListById(id)
	if err != nil {
		return err
	}
	err = removeListFromDB(id)
	if err != nil {
		return err
	}
	for _, guest := range list.Guests {
		if err = removeFromAccessListsShared(guest, id); err != nil {
			log.Error("Failed to remove ", id, " from shared list of ", guest, ": ", err)
		}
	}
	return nil
}

func listOwnedLists(username string) ([]listLink, error) {
	acc, err := getAccessByUsername(username)
	if err != nil {
		return nil, err
	}
	return acc.OwnedLists, nil
}

func listSharedLists(username string) ([]listLink, error) {
	acc, err := getAccessByUsername(username)
	if err != nil {
		return nil, err
	}
	return acc.OwnedLists, nil
}

func InitNewUser(username string) error {
	return insertAccessRecord(accessRecord{
		Username:    username,
		OwnedLists:  make([]listLink, 0),
		SharedLists: make([]listLink, 0),
	})
}
