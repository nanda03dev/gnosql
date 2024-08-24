// @title GnoSQL Database
// @version 1.0
// @description     No sql database in Go using Gin framework.

// @contact.name   Nanda Kumar

// @contact.url    https://twitter.com/nanda0311

// @contact.email  nanda03dev@gmail.com

// @license.name  Apache 2.0

// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host      localhost:5454
// @BasePath  /api

package main

import (
	"fmt"
	docs "gnosql/docs"
	pb "gnosql/proto"
	"gnosql/src/common"
	"gnosql/src/grpc_handler"
	"gnosql/src/in_memory_database"
	"gnosql/src/router"
	"html/template"
	"log"
	"net"
	"os"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"google.golang.org/grpc"
)

var (
	// Port for gRPC server to listen to
	GRPC_PORT = "5455"
	GIN_PORT  = "5454"
)

// @BasePath /api/v1
func main() {
	ginRouter := gin.Default()
	ginRouter.SetHTMLTemplate(template.Must(template.ParseGlob("./src/templates/*")))

	if port := os.Getenv("GIN_PORT"); port != "" {
		GIN_PORT = port
	}
	if port := os.Getenv("GRPC_PORT"); port != "" {
		GRPC_PORT = port
	}

	fmt.Printf("\n GIN_PORT: %v", GIN_PORT)
	fmt.Printf("\n GRPC_PORT: %v", GRPC_PORT)

	// Creating gnosql/db folder
	common.CreateDatabaseFolder()

	// Creating Gnosql
	var gnoSQL *in_memory_database.GnoSQL = in_memory_database.CreateGnoSQL()

	// Load existing database
	gnoSQL.LoadAllDBs()

	router.RouterInit(ginRouter, gnoSQL)

	docs.SwaggerInfo.BasePath = "/"
	docs.SwaggerInfo.Host = "localhost:" + GIN_PORT

	// Swagger handler
	ginRouter.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	// EX: localhost:5454/swagger/index.html

	// Start the server in a separate goroutine
	go func() {
		if err := ginRouter.Run(":" + GIN_PORT); err != nil {
			fmt.Printf("Error starting server: %v\n", err)
		}
	}()

	go func() {
		lis, err := net.Listen("tcp", ":"+GRPC_PORT)

		if err != nil {
			log.Fatalf("failed connection: %v", err)
		}

		s := grpc.NewServer()

		pb.RegisterGnoSQLServiceServer(s, &grpc_handler.GnoSQLServer{GnoSQL: gnoSQL})

		fmt.Println("GRPC server started successfully ", lis.Addr())

		if err := s.Serve(lis); err != nil {
			log.Fatalf("failed to server: %v", err)
		}
	}()

	// Ensure the main goroutine doesn't exit immediately
	select {}

}
