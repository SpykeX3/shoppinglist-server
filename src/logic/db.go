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

func initDB(url, dbName, accessCollectionName, listCollectionName string) error {
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
