package database

import (
	"context"
	"fmt"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var mongoDB *mongo.Database

const CTimeOut = 10 * time.Second

func EnvMongoURI() string {
	err := godotenv.Load()
	if err != nil {
		fmt.Println("Error loading .env file")
	}

	return os.Getenv("MONGO_URI")
}

func InitMongo() error {
	ctx, cancel := context.WithTimeout(context.Background(), CTimeOut)
	defer cancel()

	clientOptions := options.Client().ApplyURI(EnvMongoURI())
	if client, err := mongo.Connect(ctx, clientOptions); err != nil {
		return err
	} else {
		fmt.Println("Mongo: Kết nối thành công")
		err = client.Ping(ctx, readpref.Primary())
		if err != nil {
			fmt.Println("Ping không thành công")
			os.Exit(1)
		}
		mongoDB = client.Database("myDB")
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
