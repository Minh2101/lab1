package router

import (
	"github.com/gin-gonic/gin"
	"lab1/controllers"
)

func AuthRouter(auth *gin.RouterGroup) {
	auth.POST("/register", controllers.Register)
	auth.POST("/login", controllers.Login)
}
