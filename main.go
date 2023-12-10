// @title GnoSQL Database
// @version 1.0
// @description     No sql database in Go using Gin framework.

// @contact.name   Nanda Kumar

// @contact.url    https://twitter.com/nanda0311

// @contact.email  nanda23311@gmail.com

// @license.name  Apache 2.0

// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host      localhost:8080
// @BasePath  /api

package main

import (
	docs "gnosql/docs"
	"gnosql/src/in_memory_database"
	"gnosql/src/router"
	"gnosql/src/utils"
	"os"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// @BasePath /api/v1
func main() {
	ginRouter := gin.Default()

	port := os.Getenv("PORT")

	if port == "" {
		port = "8080"
	}

	utils.CreateDatabaseFolder()

	var gnoSQL *in_memory_database.GnoSQL = in_memory_database.CreateGnoSQL()

	router.GenerateSeedRoute(ginRouter, gnoSQL)

	router.LoadDatabasesAndRoutes(ginRouter, gnoSQL)

	router.GenerateDatabaseRoutes(ginRouter, gnoSQL)

	docs.SwaggerInfo.BasePath = "/"

	// Swagger handler
	ginRouter.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	ginRouter.Run(":" + port)

	defer gnoSQL.WriteAllDatabases()
}
