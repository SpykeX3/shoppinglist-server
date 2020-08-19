package credentials

import (
	"context"
	"crypto/md5"
	"errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"go.mongodb.org/mongo-driver/x/bsonx"
	"time"
)

type CredController interface {
	Login(username, password string) error
	Register(username, password string) error
}

type mongoController struct {
	cancel     context.CancelFunc
	collection *mongo.Collection
}

func NewMongoDBCredentials(url, databaseName, collectionName string) (CredController, error) {
	client, err := mongo.NewClient(options.Client().ApplyURI(url))
	if err != nil {
		return nil, err
	}
	ctx, contextCancel := context.WithTimeout(context.Background(), 10*time.Second)
	err = client.Connect(ctx)
	if err != nil {
		return nil, err
	}
	err = client.Ping(ctx, readpref.Primary())
	if err != nil {
		return nil, err
	}
	collection := client.Database(databaseName).Collection(collectionName)
	_, err = collection.Indexes().CreateOne(context.TODO(),
		mongo.IndexModel{
			Keys:    bsonx.Doc{{"username", bsonx.Int32(1)}},
			Options: options.Index().SetUnique(true),
		})
	if err != nil {
		return nil, err
	}
	return mongoController{
		cancel:     contextCancel,
		collection: collection,
	}, nil
}

func (mc mongoController) Login(username, password string) error {
	res := mc.collection.FindOne(context.TODO(), bson.D{{"username", username}, {"password", hashPassword(password)}})
	if res.Err() != nil {
		return errors.New("invalid credentials")
	}
	return nil
}

func (mc mongoController) Register(username, password string) error {

	_, err := mc.collection.InsertOne(context.TODO(), bson.D{{"username", username}, {"password", hashPassword(password)}})
	if err != nil {
		return errors.New("invalid credentials")
	}
	return nil
}

func hashPassword(password string) [16]byte {
	return md5.Sum([]byte(password))
}
