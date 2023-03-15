package router

import (
	"lab1/controllers"

	"github.com/gin-gonic/gin"
)

func WebRouter(router *gin.RouterGroup) {
	common := router.Group("/")
	common.POST("/item", controllers.CreateItem)
}
