package main

import (
	"fmt"
	"github.com/joho/godotenv"
	"lab1/database"
	"lab1/router"
	"os"

	"github.com/gin-gonic/gin"
)

func main() {
	//Init mongo
	if err := database.InitMongo(EnvMongoURI(), EnvMongoDBName()); err != nil {
		return
	}
	defer database.CloseMongoDB()

	// Router
	r := gin.New()
	router.Router(r)
	r.Run("0.0.0.0:8080")
}
func EnvMongoURI() string {
	err := godotenv.Load()
	if err != nil {
		fmt.Println("Error loading .env file")
	}
	return os.Getenv("MONGO_URI")
}
func EnvMongoDBName() string {
	err := godotenv.Load()
	if err != nil {
		fmt.Println("Error loading .env file")
	}
	return os.Getenv("DB_NAME")
}
