package main

import (
	"github.com/gin-gonic/gin"
	"gnosql/src/in_memory_database"
	"gnosql/src/router"
	"gnosql/src/seed"
	"os"
)

func main() {
	ginRouter := gin.Default()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	db := in_memory_database.CreateDatabase()

	seed.SeedData(ginRouter, db)

	router.GenerateCollectionRoutes(ginRouter, db)

	ginRouter.Run(":" + port)
}
