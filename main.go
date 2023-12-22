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
	docs "gnosql/docs"
	pb "gnosql/proto"
	"gnosql/src/grpc_handler"
	"gnosql/src/in_memory_database"
	"gnosql/src/router"
	"gnosql/src/utils"
	"log"
	"net"
	"os"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"google.golang.org/grpc"
)

const (
	// Port for gRPC server to listen to
	PORT = ":5455"
)

// @BasePath /api/v1
func main() {
	ginRouter := gin.Default()

	port := os.Getenv("PORT")

	if port == "" {
		port = "5454"
	}

	utils.CreateDatabaseFolder()
	var gnoSQL *in_memory_database.GnoSQL = in_memory_database.CreateGnoSQL()
	gnoSQL.LoadAllDBs()

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

	go func() {
		lis, err := net.Listen("tcp", PORT)

		if err != nil {
			log.Fatalf("failed connection: %v", err)
		}

		s := grpc.NewServer()

		pb.RegisterGnoSQLServiceServer(s, &grpc_handler.GnoSQLServer{GnoSQL: gnoSQL})

		log.Printf("server listening at %v", lis.Addr())

		if err := s.Serve(lis); err != nil {
			log.Fatalf("failed to server: %v", err)
		}
	}()

	// Ensure the main goroutine doesn't exit immediately
	select {}

}
