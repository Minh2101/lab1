package middlewares

import (
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"lab1/controllers"
	"net/http"
	"time"
)

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Lấy JWT token từ header
		tokenString := c.Request.Header.Get("Authorization")
		if tokenString == "" {
			controllers.ResponseError(c, http.StatusUnauthorized, "Missing Authorization Header", nil)
			return
		}

		// Parse JWT token
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return []byte("aCSnbH6B1ATyRIDkOS3pB9xXMwOza9m7XrPnceNNVXxwvkbqjXwqgTuFgD1j6GsA"), nil
		})
		if err != nil {
			controllers.ResponseError(c, http.StatusUnauthorized, "Invalid Token", err)
			return
		}
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok || !token.Valid {
			controllers.ResponseError(c, http.StatusUnauthorized, "Invalid Token", err)
			return
		}
		// Kiểm tra thời gian hết hạn của token
		expiredAt, _ := claims["exp"].(float64)
		if int64(expiredAt) < time.Now().Unix() {
			controllers.ResponseError(c, http.StatusUnauthorized, "Token expired", err)
			return
		}

		// Lưu thông tin user vào context
		c.Set("user_id", claims["ID"])
		c.Set("expired_at", claims["exp"])
		c.Next()
	}
}
