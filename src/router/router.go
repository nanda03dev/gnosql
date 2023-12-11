package router

import (
	"github.com/gin-gonic/gin"
	_ "gnosql/docs"
	"gnosql/src/handler"
	"gnosql/src/in_memory_database"
	"gnosql/src/seed"
	"net/http"
)

func RouterInit(router *gin.Engine, gnoSQL *in_memory_database.GnoSQL) {
	SeedRoute(router, gnoSQL)
	DatabaseRoutes(router, gnoSQL)
	CollectionRoutes(router, gnoSQL)
	DocumentRoutes(router, gnoSQL)
}

// @Summary      generate seed database
// @Description  This will create generate seed database.
// @Tags         generate-seed-data
// @Produce      json
// @Success      200
// @Router       /generate-seed-data [get]
func SeedRoute(router *gin.Engine, gnoSQL *in_memory_database.GnoSQL) {
	router.GET("/generate-seed-data", func(c *gin.Context) {
		var database *in_memory_database.Database = seed.SeedData(router, gnoSQL)
		if database == nil {
			c.JSON(http.StatusBadRequest, gin.H{"status": "Seed database and routes exists already"})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"status": "Seed database and routes created"})
	})
}

func DatabaseRoutes(router *gin.Engine, gnoSQL *in_memory_database.GnoSQL) {
	router.POST("/add-database", func(c *gin.Context) {
		handler.CreateDatabase(c, router, gnoSQL)
	})

	router.POST("/delete-database", func(c *gin.Context) {
		handler.DeleteDatabase(c, gnoSQL)
	})

	router.GET("/get-all-database", func(c *gin.Context) {
		handler.GetAllDatabases(c, gnoSQL)
	})
}

func CollectionRoutes(router *gin.Engine, gnoSQL *in_memory_database.GnoSQL) {

	path := "/:DatabaseName"

	router.POST(path+"/add-collections", func(c *gin.Context) {
		DatabaseName := c.Param("DatabaseName")
		var db *in_memory_database.Database = gnoSQL.GetDatabase(DatabaseName)

		handler.CreateCollection(c, router, db)
	})

	router.DELETE(path+"/delete-collections", func(c *gin.Context) {
		DatabaseName := c.Param("DatabaseName")
		var db *in_memory_database.Database = gnoSQL.GetDatabase(DatabaseName)
		handler.DeleteCollection(c, db)
	})

	router.GET(path+"/get-all-collection", func(c *gin.Context) {
		DatabaseName := c.Param("DatabaseName")
		var db *in_memory_database.Database = gnoSQL.GetDatabase(DatabaseName)
		handler.GetAllCollections(c, db)
	})

}

func DocumentRoutes(router *gin.Engine, gnoSQL *in_memory_database.GnoSQL) {
	path := "/:DatabaseName/:CollectionName"

	// Create
	router.POST(path, func(c *gin.Context) {
		db, collection := gnoSQL.GetDatabaseAndCollection(c.Param("DatabaseName"), c.Param("CollectionName"))
		handler.CreateDocument(c, db, collection)
	})

	// Read
	router.GET(path+"/:id", func(c *gin.Context) {
		db, collection := gnoSQL.GetDatabaseAndCollection(c.Param("DatabaseName"), c.Param("CollectionName"))
		handler.ReadDocument(c, db, collection)
	})

	// Read by index
	router.POST(path+"/filter", func(c *gin.Context) {
		db, collection := gnoSQL.GetDatabaseAndCollection(c.Param("DatabaseName"), c.Param("CollectionName"))
		handler.FilterDocument(c, db, collection)
	})

	// Update
	router.PUT(path+"/:id", func(c *gin.Context) {
		db, collection := gnoSQL.GetDatabaseAndCollection(c.Param("DatabaseName"), c.Param("CollectionName"))
		handler.UpdateDocument(c, db, collection)
	})

	// Delete
	router.DELETE(path+"/:id", func(c *gin.Context) {
		db, collection := gnoSQL.GetDatabaseAndCollection(c.Param("DatabaseName"), c.Param("CollectionName"))
		handler.DeleteDocument(c, db, collection)
	})

	// Get all data
	router.GET(path+"/all-data", func(c *gin.Context) {
		db, collection := gnoSQL.GetDatabaseAndCollection(c.Param("DatabaseName"), c.Param("CollectionName"))
		handler.ReadAllDocument(c, db, collection)
	})

	// Get collection stats
	router.GET(path+"/stats", func(c *gin.Context) {
		db, collection := gnoSQL.GetDatabaseAndCollection(c.Param("DatabaseName"), c.Param("CollectionName"))
		handler.CollectionStats(c, db, collection)
	})
}
