package collections

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"lab1/database"
	"time"
)

type UserToken struct {
	ID        primitive.ObjectID `bson:"_id" json:"id"`
	Token     string             `bson:"token" json:"token"`
	ExpiredAt time.Time          `bson:"expired_at" json:"expired_at"`
	UserID    primitive.ObjectID `bson:"user_id" json:"user_id"`

	CreatedAt  time.Time  `bson:"created_at" json:"created_at"`
	ModifiedAt time.Time  `bson:"modified_at" json:"modified_at"`
	DeletedAt  *time.Time `bson:"deleted_at" json:"deleted_at"`
}

func (u *UserToken) collectionName() string {
	return "user_tokens"
}

func (u *UserToken) Create(DB *mongo.Database) error {
	ctx, cancel := context.WithTimeout(context.Background(), database.CTimeOut)
	defer cancel()
	u.ID = primitive.NewObjectID()
	u.CreatedAt = time.Now()
	u.ModifiedAt = time.Now()
	if _, err := DB.Collection(u.collectionName()).InsertOne(ctx, u); err != nil {
		return err
	}
	return nil
}

func (u *UserToken) FindByToken(DB *mongo.Database, token string) error {
	filter := bson.M{
		"token": token,
	}
	ctx, cancel := context.WithTimeout(context.Background(), database.CTimeOut)
	defer cancel()
	if result := DB.Collection(u.collectionName()).FindOne(ctx, filter); result.Err() != nil {
		return result.Err()
	} else {
		return result.Decode(&u)
	}
}
