package main

import (
	"lab1/database"
	"lab1/router"

	"github.com/gin-gonic/gin"
)

func main() {
	//Init mongo
	if err := database.InitMongo(); err != nil {
		return
	}
	defer database.CloseMongoDB()

	// Router
	r := gin.New()
	router.Router(r)
	r.Run("0.0.0.0:8080")
}
