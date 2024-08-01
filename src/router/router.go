package router

import (
	_ "gnosql/docs"
	"gnosql/src/handler"
	"gnosql/src/in_memory_database"
	"gnosql/src/seed"
	"gnosql/src/service"
	// "html/template"
	"net/http"

	"github.com/gin-gonic/gin"
)

type FilterQuery struct {
	DatabaseName   string `form:"databaseName"`
	CollectionName string `form:"collectionName"`
	DocumentId     string `form:"documentId"`
}

func RouterInit(ginRouter *gin.Engine, gnoSQL *in_memory_database.GnoSQL) {
	SeedRoute(ginRouter, gnoSQL)
	DatabaseRoutes(ginRouter, gnoSQL)
	CollectionRoutes(ginRouter, gnoSQL)
	DocumentRoutes(ginRouter, gnoSQL)
	UIRoutes(ginRouter, gnoSQL)
}

func UIRoutes(ginRouter *gin.Engine, gnoSQL *in_memory_database.GnoSQL) {
	ginRouter.GET("/gnosql-ui", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", nil)
	})
	ginRouter.POST("/gnosql-ui/submit", func(c *gin.Context) {
		var filterQuery FilterQuery
		var response []in_memory_database.Document

		if err := c.ShouldBind(&filterQuery); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		var filter in_memory_database.MapInterface

		if len(filterQuery.DocumentId) > 10 {
			filter = in_memory_database.MapInterface{
				"docId": filterQuery.DocumentId,
			}
		}

		if len(filterQuery.CollectionName) > 0 {
			result := service.DocumentFilter(gnoSQL, filterQuery.DatabaseName, filterQuery.CollectionName, filter)

			if len(result.Data) > 0 {
				response = result.Data
			}

		} else {
			result := service.GetAllCollections(gnoSQL, filterQuery.DatabaseName)

			if len(result.Data) > 0 {

				for _, collectionName := range result.Data {
					response = append(response, in_memory_database.Document{"collectionName": collectionName})
				}
			}
		}

		c.HTML(http.StatusOK, "table.html", gin.H{"data": response})
	})
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
