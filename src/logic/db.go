package logic

import (
	"context"
	"errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"go.mongodb.org/mongo-driver/x/bsonx"
	"time"
)

var listCollection *mongo.Collection
var accessCollection *mongo.Collection

func InitDB(url, dbName, accessCollectionName, listCollectionName string) error {
	client, err := mongo.NewClient(options.Client().ApplyURI(url))
	if err != nil {
		return err
	}
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	err = client.Connect(ctx)
	if err != nil {
		return err
	}
	err = client.Ping(ctx, readpref.Primary())
	if err != nil {
		return err
	}
	listCollection = client.Database(dbName).Collection(listCollectionName)
	_, err = listCollection.Indexes().CreateOne(context.TODO(),
		mongo.IndexModel{
			Keys:    bsonx.Doc{{"id", bsonx.Int32(1)}},
			Options: options.Index().SetUnique(true),
		})
	if err != nil {
		return err
	}
	accessCollection = client.Database(dbName).Collection(accessCollectionName)
	_, err = listCollection.Indexes().CreateOne(context.TODO(),
		mongo.IndexModel{
			Keys:    bsonx.Doc{{"id", bsonx.Int32(1)}},
			Options: options.Index().SetUnique(true),
		})
	if err != nil {
		return err
	}
	return nil
}

func getAccessByUsername(username string) (*accessRecord, error) {
	res := accessCollection.FindOne(context.TODO(), bson.D{{"username", username}})
	if res.Err() != nil {
		return nil, res.Err()
	}
	var record accessRecord
	err := res.Decode(&record)
	if err != nil {
		return nil, err
	}
	return &record, nil
}

func getListById(id string) (*list, error) {
	res := listCollection.FindOne(context.TODO(), bson.D{{"id", id}})
	if res.Err() != nil {
		return nil, res.Err()
	}
	var listRec list
	err := res.Decode(&listRec)
	if err != nil {
		return nil, err
	}
	return &listRec, nil
}

func updateList(id, content string) error {
	res, err := listCollection.UpdateOne(context.TODO(), bson.D{{"id", id}}, bson.D{{"$set", bson.D{{"last_changed", time.Now()}, {"content", content}}}})
	if err != nil {
		return err
	}
	if res.ModifiedCount != 1 {
		return errors.New("couldn't modify selected list")
	}
	return nil
}

func insertListInDB(listRec list) error {
	_, err := listCollection.InsertOne(context.TODO(), listRec)
	if err != nil {
		return err
	}
	return nil
}

func insertAccessRecord(record accessRecord) error {
	_, err := accessCollection.InsertOne(context.TODO(), record)
	if err != nil {
		return err
	}
	return nil
}

func updateAccessRecord(record accessRecord) error {
	_, err := accessCollection.UpdateOne(context.TODO(), bson.D{{"username", record.Username}}, bson.D{{"$set", record}})
	if err != nil {
		return err
	}
	return nil
}

/*
May have race condition. Not too critical, but would be better to fix it.
*/
func addToAccessLists(username, id string) error {
	res := accessCollection.FindOne(context.TODO(), bson.D{{"username", username}})
	if res.Err() != nil {
		return res.Err()
	}
	var record accessRecord
	err := res.Decode(&record)
	if err != nil {
		return err
	}
	ls := listCollection.FindOne(context.TODO(), bson.D{{"id", id}})
	if ls.Err() != nil {
		return ls.Err()
	}
	var listRec list
	err = ls.Decode(&listRec)
	if err != nil {
		return err
	}
	record.AvailableLists = append(record.AvailableLists, listLink{
		Id:          id,
		DisplayName: listRec.OriginalName,
	})
	err = updateAccessRecord(record)
	if err != nil {
		return err
	}
	return nil
}

/*
May have race condition. Not too critical, but would be better to fix it.
*/
func removeFromAccessLists(username, id string) error {
	res := accessCollection.FindOne(context.TODO(), bson.D{{"username", username}})
	if res.Err() != nil {
		return res.Err()
	}
	var record accessRecord
	err := res.Decode(&record)
	if err != nil {
		return err
	}
	var newLists []listLink = nil
	for i, lst := range record.AvailableLists {
		if lst.Id == id {
			last := len(record.AvailableLists) - 1
			record.AvailableLists[i] = record.AvailableLists[last]
			newLists = record.AvailableLists[:last]
			break
		}
	}
	if newLists == nil {
		return errors.New("no such list")
	}
	record.AvailableLists = newLists
	err = updateAccessRecord(record)
	if err != nil {
		return err
	}
	return nil
}
