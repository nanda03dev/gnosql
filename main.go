// @title GnoSQL Database
// @version 1.0
// @description     No sql database in Go using Gin framework.

// @contact.name   Nanda Kumar

// @contact.url    https://twitter.com/nanda0311

// @contact.email  nanda23311@gmail.com

// @license.name  Apache 2.0

// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host      localhost:5454
// @BasePath  /api

package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	docs "gnosql/docs"
	"gnosql/src/in_memory_database"
	"gnosql/src/router"
	"gnosql/src/utils"
	"os"
	"os/signal"
	"syscall"
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

	gnoSQL.LoadAllDatabases()

	router.RouterInit(ginRouter, gnoSQL)

	docs.SwaggerInfo.BasePath = "/"
	docs.SwaggerInfo.Host = "localhost:5454"

	// Swagger handler
	ginRouter.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Start the server in a separate goroutine
	go func() {
		if err := ginRouter.Run(":" + port); err != nil {
			fmt.Printf("Error starting server: %v\n", err)
		}
	}()

	// Handle signals for cleanup before application stops
	handleSignals()

	// Ensure the main goroutine doesn't exit immediately
	select {}
}

func handleSignals() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	go func() {
		// Wait for a signal
		sig := <-c
		fmt.Printf("\nReceived signal: %v\n", sig)

		// Perform cleanup or specific logic before the application stops
		cleanup()

		// Exit the application
		os.Exit(1)
	}()
}

func cleanup() {
	// Add your cleanup logic here
	// This function will be executed when the application receives a signal
	fmt.Println("Performing cleanup before application stops...")
	// Example: Close database connections, save state, etc.
}
