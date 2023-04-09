package router

import (
	"lab1/controllers"

	"github.com/gin-gonic/gin"
)

func WebRouter(router *gin.RouterGroup) {
	common := router.Group("/")

	// Items
	common.GET("/items", controllers.ListItems)
	common.POST("/item", controllers.CreateItem)
	common.PUT("/item/:id", controllers.UpdateItem)
	common.POST("/change-status-items", controllers.ChangeStatusItems)
	common.POST("/delete-items", controllers.DeleteItems)
	common.GET("/export-items", controllers.ExportListItems)

	//ExportPDF
	common.POST("/export-pdf", controllers.ExportPDF)
}
