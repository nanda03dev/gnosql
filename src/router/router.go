package router

import (
	"fmt"
	"github.com/gin-gonic/gin"
	_ "gnosql/docs"
	"gnosql/src/handler"
	"gnosql/src/in_memory_database"
	"gnosql/src/seed"
	"net/http"
)

// @Summary      generate seed database
// @Description  This will create generate seed database.
// @Tags         generate-seed-data
// @Produce      json
// @Success      200
// @Router       /generate-seed-data [get]
func GenerateSeedRoute(router *gin.Engine, gnoSQL *in_memory_database.GnoSQL) {
	router.GET("/generate-seed-data", func(c *gin.Context) {
		var database *in_memory_database.Database = seed.SeedData(router, gnoSQL)
		if database == nil {
			c.JSON(http.StatusBadRequest, gin.H{"status": "Seed database and routes exists already"})
			return
		}

		GenerateCollectionRoutes(router, database)
		GenerateEntityRoutes(router, database)
		c.JSON(http.StatusBadRequest, gin.H{"status": "Seed database and routes created"})

	})
}

func LoadDatabasesAndRoutes(router *gin.Engine, gnoSQL *in_memory_database.GnoSQL) {
	var databases []*in_memory_database.Database = gnoSQL.LoadAllDatabases()

	for _, database := range databases {
		GenerateCollectionRoutes(router, database)
		GenerateEntityRoutes(router, database)
	}

}

func GenerateDatabaseRoutes(router *gin.Engine, gnoSQL *in_memory_database.GnoSQL) {
	router.POST("/add-database", func(c *gin.Context) {
		handler.CreateDatabase(c, router, gnoSQL, GenerateCollectionRoutes)
	})

	router.POST("/delete-database", func(c *gin.Context) {
		handler.DeleteDatabase(c, gnoSQL)
	})

	router.GET("/get-all-database", func(c *gin.Context) {
		handler.GetAllDatabases(c, gnoSQL)
	})
}

func GenerateCollectionRoutes(router *gin.Engine, db *in_memory_database.Database) {

	path := fmt.Sprintf("/%s", db.DatabaseName)

	router.POST(path+"/add-collections", func(c *gin.Context) {
		handler.CreateCollection(c, router, db, GenerateEntityRoutes)
	})

	router.DELETE(path+"/delete-collections", func(c *gin.Context) {
		handler.DeleteCollection(c, db)
	})

	router.GET(path+"/get-all-collection", func(c *gin.Context) {
		handler.GetAllCollections(c, db)
	})

}

func GenerateEntityRoutes(router *gin.Engine, db *in_memory_database.Database) {
	for _, collection := range db.Collections {
		generateRoutes(router, db, collection)
	}
}

func generateRoutes(router *gin.Engine, db *in_memory_database.Database, collection *in_memory_database.Collection) {
	path := fmt.Sprintf("/%s/%s", db.DatabaseName, collection.CollectionName)

	// Create
	router.POST(path, func(c *gin.Context) {
		handler.CreateDocument(c, db, collection)
	})

	// Read
	router.GET(path+"/:id", func(c *gin.Context) {
		handler.ReadDocument(c, db, collection)
	})

	// Read by index
	router.POST(path+"/filter", func(c *gin.Context) {
		handler.FilterDocument(c, db, collection)
	})

	// Update
	router.PUT(path+"/:id", func(c *gin.Context) {
		handler.UpdateDocument(c, db, collection)
	})

	// Delete
	router.DELETE(path+"/:id", func(c *gin.Context) {
		handler.DeleteDocument(c, db, collection)
	})

	// Get all data
	router.GET(path+"/all-data", func(c *gin.Context) {
		handler.ReadAllDocument(c, db, collection)
	})

	// Get collection stats
	router.GET(path+"/stats", func(c *gin.Context) {
		handler.CollectionStats(c, db, collection)
	})
}
