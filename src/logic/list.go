package logic

import "time"

type list struct {
	Id          string    `bson:"id" json:"id"`
	LastChanged time.Time `bson:"last_changed" json:"last_changed"`
	Content     string    `bson:"content" json:"content"`
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
