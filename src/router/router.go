package router

import (
	_ "gnosql/docs"
	"gnosql/src/handler"
	"gnosql/src/in_memory_database"
	"gnosql/src/seed"
	"net/http"

	"github.com/gin-gonic/gin"
)

func RouterInit(router *gin.Engine, gnoSQL *in_memory_database.GnoSQL) {
	// // Middleware to capture DatabaseName and store it in the context
	// router.Use(func(c *gin.Context) {
	// 	DatabaseName := c.Param("DatabaseName")
	// 	CollectionName := c.Param("CollectionName")

	// 	if len(DatabaseName) > 0 {
	// 		var db *in_memory_database.Database = gnoSQL.GetDB(DatabaseName)

	// 		if db == nil {
	// 			c.JSON(http.StatusBadRequest, gin.H{"message": "database not found"})
	// 			return
	// 		}

	// 		if len(CollectionName) > 0 {
	// 			var collection *in_memory_database.Collection = db.GetColl(CollectionName)
	// 			if collection == nil {
	// 				c.JSON(http.StatusBadRequest, gin.H{"message": "collection not found"})
	// 				return
	// 			}

	// 		}
	// 	}
	// 	c.Next()
	// })

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
		var database *in_memory_database.Database = seed.SeedData(gnoSQL)
		if database == nil {
			c.JSON(http.StatusBadRequest, gin.H{"status": "Seed database and routes exists already"})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"status": "Seed database and routes created"})
	})
}

func DatabaseRoutes(router *gin.Engine, gnoSQL *in_memory_database.GnoSQL) {
	path := "/database"

	DatabaseRoutesGroup := router.Group(path)
	{
		DatabaseRoutesGroup.POST("/add", func(c *gin.Context) {
			handler.CreateDatabase(c, gnoSQL)
		})

		DatabaseRoutesGroup.POST("/delete", func(c *gin.Context) {
			handler.DeleteDatabase(c, gnoSQL)
		})

		DatabaseRoutesGroup.GET("/get-all", func(c *gin.Context) {
			handler.GetAllDatabases(c, gnoSQL)
		})

		DatabaseRoutesGroup.GET("/load-to-disk", func(c *gin.Context) {
			handler.LoadDatabaseToDisk(c, gnoSQL)
		})

	}
}

func CollectionRoutes(router *gin.Engine, gnoSQL *in_memory_database.GnoSQL) {

	path := "/collection/:DatabaseName"

	CollectionRoutesGroup := router.Group(path)
	{
		CollectionRoutesGroup.POST("/add", func(c *gin.Context) {
			handler.CreateCollection(c, gnoSQL)
		})

		CollectionRoutesGroup.DELETE("/delete", func(c *gin.Context) {
			handler.DeleteCollection(c, gnoSQL)
		})

		CollectionRoutesGroup.GET("/get-all", func(c *gin.Context) {
			handler.GetAllCollections(c, gnoSQL)
		})

		// Get collection stats
		CollectionRoutesGroup.GET("/:CollectionName/stats", func(c *gin.Context) {
			handler.CollectionStats(c, gnoSQL)
		})
	}

}

func DocumentRoutes(router *gin.Engine, gnoSQL *in_memory_database.GnoSQL) {
	path := "/document/:DatabaseName/:CollectionName"

	DocumentRoutesGroup := router.Group(path)
	{
		// Create
		DocumentRoutesGroup.POST("/", func(c *gin.Context) {
			db, collection := gnoSQL.GetDatabaseAndCollection(c.Param("DatabaseName"), c.Param("CollectionName"))
			handler.CreateDocument(c, db, collection)
		})

		// Read
		DocumentRoutesGroup.GET("/:id", func(c *gin.Context) {
			db, collection := gnoSQL.GetDatabaseAndCollection(c.Param("DatabaseName"), c.Param("CollectionName"))
			handler.ReadDocument(c, db, collection)
		})

		// Read by index
		DocumentRoutesGroup.POST("/filter", func(c *gin.Context) {
			db, collection := gnoSQL.GetDatabaseAndCollection(c.Param("DatabaseName"), c.Param("CollectionName"))
			handler.FilterDocument(c, db, collection)
		})

		// Update
		DocumentRoutesGroup.PUT("/:id", func(c *gin.Context) {
			db, collection := gnoSQL.GetDatabaseAndCollection(c.Param("DatabaseName"), c.Param("CollectionName"))
			handler.UpdateDocument(c, db, collection)
		})

		// Delete
		DocumentRoutesGroup.DELETE("/:id", func(c *gin.Context) {
			db, collection := gnoSQL.GetDatabaseAndCollection(c.Param("DatabaseName"), c.Param("CollectionName"))
			handler.DeleteDocument(c, db, collection)
		})

		// Get all data
		DocumentRoutesGroup.GET("/all-data", func(c *gin.Context) {
			db, collection := gnoSQL.GetDatabaseAndCollection(c.Param("DatabaseName"), c.Param("CollectionName"))
			handler.ReadAllDocument(c, db, collection)
		})
	}

}
