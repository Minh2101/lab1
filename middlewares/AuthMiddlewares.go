package middlewares

import (
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
	"lab1/collections"
	"lab1/controllers"
	"lab1/database"
	"time"
)

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		var (
			token     = c.Request.Header.Get("Authorization")
			user      collections.User
			db        = database.GetMongoDB()
			userToken collections.UserToken
		)
		if token == "" {
			controllers.ResponseError(c, 401, "Hết phiên đăng nhập", nil)
			return
		}
		if err := userToken.FindByToken(db, token); err == nil {
			// Check token expired
			if userToken.ExpiredAt.Unix() < time.Now().Unix() {
				controllers.ResponseError(c, 401, "Hết phiên đăng nhập", err.Error())
				return
			}
			// FindByID
			if err = user.FindByID(db, userToken.UserID); err == nil {
				c.Set("user", user)
				c.Set("token", token)
				c.Next()
			} else if err == mongo.ErrNoDocuments {
				controllers.ResponseError(c, 404, "Chưa đăng nhập", err.Error())
				return
			} else {
				controllers.ResponseError(c, 500, "Server Error", err.Error())
				return
			}
		} else if err == mongo.ErrNoDocuments {
			controllers.ResponseError(c, 404, "Chưa đăng nhập", err.Error())
			return
		} else {
			controllers.ResponseError(c, 500, "Server Error", err.Error())
			return
		}
	}
}
