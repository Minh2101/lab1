package router

import (
	"lab1/controllers"

	"github.com/gin-gonic/gin"
)

func WebRouter(router *gin.RouterGroup) {
	common := router.Group("/")
	common.POST("/item", controllers.CreateItem)
	common.PUT("/item/:id", controllers.UpdateItem)
	common.POST("/change-status-items", controllers.ChangeStatusItems)
	common.POST("/delete-items", controllers.DeleteItems)
}
