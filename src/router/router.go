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
			handler.CreateDocument(c, gnoSQL)
		})

		// Read
		DocumentRoutesGroup.GET("/:id", func(c *gin.Context) {
			handler.ReadDocument(c, gnoSQL)
		})

		// Read by index
		DocumentRoutesGroup.POST("/filter", func(c *gin.Context) {
			handler.FilterDocument(c, gnoSQL)
		})

		// Update
		DocumentRoutesGroup.PUT("/:id", func(c *gin.Context) {
			handler.UpdateDocument(c, gnoSQL)
		})

		// Delete
		DocumentRoutesGroup.DELETE("/:id", func(c *gin.Context) {
			handler.DeleteDocument(c, gnoSQL)
		})

		// Get all data
		DocumentRoutesGroup.GET("/all-data", func(c *gin.Context) {
			handler.ReadAllDocument(c, gnoSQL)
		})
	}

}
