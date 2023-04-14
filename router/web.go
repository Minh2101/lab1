package router

import (
	"lab1/controllers"
	"lab1/exportPDF"
	"lab1/middlewares"

	"github.com/gin-gonic/gin"
)

func WebRouter(app *gin.RouterGroup) {
	app.Use(middlewares.AuthMiddleware())

	// Items
	app.GET("/items", controllers.ListItems)
	app.POST("/item", controllers.CreateItem)
	app.PUT("/item/:id", controllers.UpdateItem)
	app.POST("/change-status-items", controllers.ChangeStatusItems)
	app.POST("/delete-items", controllers.DeleteItems)
	app.GET("/export-items", controllers.ExportListItems)

	//ExportPDF
	app.POST("/export-pdf", exportPDF.ExportPDF)
}
