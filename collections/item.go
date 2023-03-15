package collections

import (
	"context"
	"time"

	"lab1/database"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type Item struct {
	ID         primitive.ObjectID `bson:"_id" json:"id"`
	Title      string             `bson:"title" json:"title"`
	Status     bool               `bson:"status" json:"status"`
	CreatedAt  time.Time          `bson:"created_at" json:"created_at"`
	ModifiedAt time.Time          `bson:"modified_at" json:"modified_at"`
}

func (u *Item) CollectionName() string {
	return "items"
}

func (u *Item) Create(DB *mongo.Database) error {
	ctx, cancel := context.WithTimeout(context.Background(), database.CTimeOut)
	defer cancel()

	u.ID = primitive.NewObjectID()
	u.CreatedAt = time.Now()
	u.ModifiedAt = time.Now()

	if _, err := DB.Collection(u.CollectionName()).InsertOne(ctx, u); err != nil {
		return err
	} else {
		return nil
	}
}
