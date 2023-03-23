package collections

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
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
	DeletedAt  *time.Time         `bson:"deleted_at" json:"deleted_at"`
}
type Items []Item

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
	}
	return nil
}

func (u *Item) First(DB *mongo.Database, filter bson.M) error {
	ctx, cancel := context.WithTimeout(context.Background(), database.CTimeOut)
	defer cancel()

	if result := DB.Collection(u.CollectionName()).FindOne(ctx, filter); result.Err() != nil {
		return result.Err()
	} else {
		err := result.Decode(&u)
		return err
	}
}

func (u *Item) Update(DB *mongo.Database) error {
	ctx, cancel := context.WithTimeout(context.Background(), database.CTimeOut)
	defer cancel()
	u.ModifiedAt = time.Now()
	if _, err := DB.Collection(u.CollectionName()).UpdateOne(ctx, bson.M{"_id": u.ID}, bson.M{
		"$set": u,
	}, options.Update()); err != nil {
		return err
	}
	return nil
}

func (u *Item) Find(DB *mongo.Database, filter bson.M, opts ...*options.FindOptions) (Items, error) {
	ctx, cancel := context.WithTimeout(context.Background(), database.CTimeOut)
	defer cancel()
	data := make(Items, 0)

	/* Lấy danh sách bản ghi */
	if cursor, err := DB.Collection(u.CollectionName()).Find(ctx, filter, opts...); err == nil {
		for cursor.Next(ctx) {
			var elem Item
			if err = cursor.Decode(&elem); err != nil {
				return data, err
			}
			data = append(data, elem)
		}
		if err = cursor.Err(); err != nil {
			return data, err
		}
		return data, cursor.Close(ctx)
	} else {
		return data, err
	}
}

func (u *Item) Delete(DB *mongo.Database) error {
	ctx, cancel := context.WithTimeout(context.Background(), database.CTimeOut)
	defer cancel()
	now := time.Now()
	u.DeletedAt = &now
	if _, err := DB.Collection(u.CollectionName()).UpdateOne(ctx, bson.M{"_id": u.ID}, bson.M{"$set": u}, nil); err != nil {
		return err
	}
	return nil
}

func (u *Item) Count(DB *mongo.Database, filter bson.M) (int64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), database.CTimeOut)
	defer cancel()
	if total, err := DB.Collection(u.CollectionName()).CountDocuments(ctx, filter, options.Count()); err != nil {
		return 0, err
	} else {
		return total, nil
	}
}
