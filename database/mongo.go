package database

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var mongoDB *mongo.Database

const CTimeOut = 10 * time.Second

func InitMongo(URI, DBName string) error {
	ctx, cancel := context.WithTimeout(context.Background(), CTimeOut)
	defer cancel()

	clientOptions := options.Client().ApplyURI(URI)
	if client, err := mongo.Connect(ctx, clientOptions); err != nil {
		return err
	} else {
		err = client.Ping(ctx, readpref.Primary())
		if err != nil {
			fmt.Println("Ping không thành công")
			os.Exit(1)
		}
		mongoDB = client.Database(DBName)
		return nil
	}
}

func GetMongoDB() *mongo.Database {
	return mongoDB
}

func CloseMongoDB() {
	ctx, cancel := context.WithTimeout(context.Background(), CTimeOut)
	defer cancel()
	_ = mongoDB.Client().Disconnect(ctx)
}
