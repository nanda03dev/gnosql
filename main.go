package main

import (
	"gnosql/src/in_memory_database"
	"gnosql/src/router"
	"gnosql/src/seed"
	"gnosql/src/utils"
	"os"

	"github.com/gin-gonic/gin"
)

func main() {

	ginRouter := gin.Default()

	port := os.Getenv("PORT")

	if port == "" {
		port = "8080"
	}

	utils.CreateDatabaseFolder()

	var gnoSQL *in_memory_database.GnoSQL = in_memory_database.CreateGnoSQL()

	router.LoadGnoSQLAndRoutes(ginRouter, gnoSQL)

	router.GenerateDatabaseRoutes(ginRouter, gnoSQL)

	seed.SeedData(ginRouter, gnoSQL)

	ginRouter.Run(":" + port)

	defer gnoSQL.WriteAllDatabases()
}
