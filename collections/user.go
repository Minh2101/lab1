package collections

import (
	"context"
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
	"lab1/database"
	"time"
)

type User struct {
	ID              primitive.ObjectID `bson:"_id" json:"id"`
	Name            string             `bson:"name" json:"name"`
	Phone           string             `bson:"phone" json:"phone"`
	Email           string             `bson:"email" json:"email"`
	UserName        string             `bson:"user_name" json:"user_name"`
	PasswordHash    string             `bson:"password_hash" json:"password_hash"`
	Password        string             `bson:"-" json:"password"`
	PasswordConfirm string             `bson:"-" json:"password_confirm"`

	CreatedAt  time.Time  `bson:"created_at" json:"created_at"`
	ModifiedAt time.Time  `bson:"modified_at" json:"modified_at"`
	DeletedAt  *time.Time `bson:"deleted_at" json:"deleted_at"`
}

func (u *User) collectionName() string {
	return "users"
}

type ItemInterface interface {
	Create(db interface{}) error
}

func (u *User) Create(DB *mongo.Database) error {
	ctx, cancel := context.WithTimeout(context.Background(), database.CTimeOut)
	defer cancel()

	u.ID = primitive.NewObjectID()
	PassHash, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
	if err != nil {
		return errors.WithStack(err)
	}
	u.PasswordHash = string(PassHash)
	u.CreatedAt = time.Now()
	u.ModifiedAt = time.Now()

	if _, err = DB.Collection(u.collectionName()).InsertOne(ctx, u); err != nil {
		return err
	}
	return nil
}

func (u *User) First(DB *mongo.Database, filter bson.M) error {
	ctx, cancel := context.WithTimeout(context.Background(), database.CTimeOut)
	defer cancel()
	if result := DB.Collection(u.collectionName()).FindOne(ctx, filter); result.Err() != nil {
		return result.Err()
	} else {
		err := result.Decode(&u)
		return err
	}
}

func (u *User) FindByID(DB *mongo.Database, ID primitive.ObjectID) error {
	ctx, cancel := context.WithTimeout(context.Background(), database.CTimeOut)
	defer cancel()
	if result := DB.Collection(u.collectionName()).FindOne(ctx, bson.M{"_id": ID}); result.Err() != nil {
		return result.Err()
	} else {
		return result.Decode(&u)
	}
}
