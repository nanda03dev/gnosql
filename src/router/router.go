package router

import (
	_ "gnosql/docs"
	"gnosql/src/handler"
	"gnosql/src/in_memory_database"
	"gnosql/src/seed"
	"net/http"

	"github.com/gin-gonic/gin"
)

func RouterInit(ginRouter *gin.Engine, gnoSQL *in_memory_database.GnoSQL) {
	SeedRoute(ginRouter, gnoSQL)
	DatabaseRoutes(ginRouter, gnoSQL)
	CollectionRoutes(ginRouter, gnoSQL)
	DocumentRoutes(ginRouter, gnoSQL)
}

// @Summary      generate seed database
// @Description  This will create generate seed database.
// @Tags         generate-seed-data
// @Produce      json
// @Success      200
// @Router       /generate-seed-data [get]
func SeedRoute(ginRouter *gin.Engine, gnoSQL *in_memory_database.GnoSQL) {
	ginRouter.GET("/generate-seed-data", func(c *gin.Context) {
		var database *in_memory_database.Database = seed.SeedData(gnoSQL)
		if database == nil {
			c.JSON(http.StatusBadRequest, gin.H{"status": "Seed database and routes exists already"})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"status": "Seed database and routes created"})
	})
}

func DatabaseRoutes(ginRouter *gin.Engine, gnoSQL *in_memory_database.GnoSQL) {
	path := "/database"

	DatabaseRoutesGroup := ginRouter.Group(path)
	{
		DatabaseRoutesGroup.POST("/connect", func(c *gin.Context) {
			handler.ConnectDatabase(c, gnoSQL)
		})

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

func CollectionRoutes(ginRouter *gin.Engine, gnoSQL *in_memory_database.GnoSQL) {

	path := "/collection"

	CollectionRoutesGroup := ginRouter.Group(path)
	{
		CollectionRoutesGroup.POST("/add", func(c *gin.Context) {
			handler.CreateCollection(c, gnoSQL)
		})

		CollectionRoutesGroup.POST("/delete", func(c *gin.Context) {
			handler.DeleteCollection(c, gnoSQL)
		})

		CollectionRoutesGroup.POST("/get-all", func(c *gin.Context) {
			handler.GetAllCollections(c, gnoSQL)
		})

		// Get collection stats
		CollectionRoutesGroup.POST("/stats", func(c *gin.Context) {
			handler.CollectionStats(c, gnoSQL)
		})
	}

}

func DocumentRoutes(ginRouter *gin.Engine, gnoSQL *in_memory_database.GnoSQL) {
	path := "/document/"

	DocumentRoutesGroup := ginRouter.Group(path)
	{
		// Create
		DocumentRoutesGroup.POST("/add", func(c *gin.Context) {
			handler.CreateDocument(c, gnoSQL)
		})

		// Read
		DocumentRoutesGroup.POST("/find", func(c *gin.Context) {
			handler.ReadDocument(c, gnoSQL)
		})

		// Read by index
		DocumentRoutesGroup.POST("/filter", func(c *gin.Context) {
			handler.FilterDocument(c, gnoSQL)
		})

		// Update
		DocumentRoutesGroup.POST("/update", func(c *gin.Context) {
			handler.UpdateDocument(c, gnoSQL)
		})

		// Delete
		DocumentRoutesGroup.POST("/delete", func(c *gin.Context) {
			handler.DeleteDocument(c, gnoSQL)
		})

		// Get all data
		DocumentRoutesGroup.POST("/all-data", func(c *gin.Context) {
			handler.ReadAllDocument(c, gnoSQL)
		})
	}
}
