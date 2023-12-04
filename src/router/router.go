package router

import (
	"encoding/json"
	"fmt"
	"gnosql/src/in_memory_database"
	"net/http"

	"github.com/gin-gonic/gin"
)

func GenerateCollectionRoutes(router *gin.Engine, db *in_memory_database.Database) {

	router.POST("/add-collections", func(c *gin.Context) {
		var value []map[string]interface{}

		if err := c.BindJSON(&value); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		var collections []in_memory_database.CollectionInput

		for _, each := range value {
			if collectionName, ok := each["collectionName"].(string); ok {

				println("CollectionName ", collectionName)
				var indexKeys = make([]string, 0)

				for _, each := range each["indexKeys"].([]interface{}) {
					println("each ", each)
					indexKeys = append(indexKeys, each.(string))

				}

				collection := in_memory_database.CollectionInput{
					CollectionName: collectionName,
					IndexKeys:      indexKeys,
				}

				collections = append(collections, collection)

			}

		}

		addedCollectionInstance := db.AddCollections(collections)

		GenerateEntityRoutes(router, addedCollectionInstance)

		c.JSON(http.StatusCreated, gin.H{"data": "collection created successfully"})
	})

	router.DELETE("/delete-collections", func(c *gin.Context) {
		var value map[string][]string

		if err := c.BindJSON(&value); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// var indexKeys = make([]string, 0)
		if collections, ok := value["collections"]; ok {
			db.DeleteCollections(collections)

		}

		c.JSON(http.StatusCreated, gin.H{"data": "successfully deleted"})
	})

	router.GET("/get-all-collections", func(c *gin.Context) {

		// Fetch all data from the database
		allCollections := db.GetCollections()
		collections := make([]string, 0)

		for _, collection := range allCollections {
			collections = append(collections, collection.GetCollectionName())
		}
		// Serialize data to JSON
		responseData, _ := json.Marshal(collections)

		// Send the JSON response
		c.Data(http.StatusOK, "application/json; charset=utf-8", responseData)
	})

}

func GenerateEntityRoutes(router *gin.Engine, collections []*in_memory_database.Collection) {
	for _, collection := range collections {
		generateRoutes(router, collection)
	}
}

func generateRoutes(router *gin.Engine, db *in_memory_database.Collection) {
	entity := db.GetCollectionName()

	path := fmt.Sprintf("/%s", entity)

	// Create
	router.POST(path, func(c *gin.Context) {
		if db.IsDeleted() {
			c.JSON(http.StatusBadRequest, gin.H{"message": db.GetCollectionName() + " collection deleted"})
			return
		}

		var value map[string]interface{}
		if err := c.BindJSON(&value); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		result := db.Create(value)

		c.JSON(http.StatusCreated, gin.H{"data": result})
	})

	// Read
	router.GET(path+"/:id", func(c *gin.Context) {
		if db.IsDeleted() {
			c.JSON(http.StatusBadRequest, gin.H{"message": db.GetCollectionName() + " collection deleted"})
			return
		}

		id := c.Param("id")
		value := db.Read(id)
		c.JSON(http.StatusOK, gin.H{"data": value})
	})

	// Read by index
	router.POST(path+"/filter", func(c *gin.Context) {
		if db.IsDeleted() {
			c.JSON(http.StatusBadRequest, gin.H{"message": db.GetCollectionName() + " collection deleted"})
			return
		}

		var value in_memory_database.GenericKeyValue

		if err := c.BindJSON(&value); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		result := db.Filter(value)

		c.JSON(http.StatusCreated, gin.H{"data": result})
	})

	// Read by index
	router.POST(path+"/filterbyindex", func(c *gin.Context) {
		if db.IsDeleted() {
			c.JSON(http.StatusBadRequest, gin.H{"message": db.GetCollectionName() + " collection deleted"})
			return
		}

		var value in_memory_database.GenericKeyValue

		if err := c.BindJSON(&value); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		result := db.FilterByIndexKey(value)

		c.JSON(http.StatusCreated, gin.H{"data": result})
	})

	// Update
	router.PUT(path+"/:id", func(c *gin.Context) {
		if db.IsDeleted() {
			c.JSON(http.StatusBadRequest, gin.H{"message": db.GetCollectionName() + " collection deleted"})
			return
		}

		id := c.Param("id")
		var value map[string]interface{}
		if err := c.BindJSON(&value); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		result := db.Update(id, value)

		c.JSON(http.StatusOK, gin.H{"data": result})
	})

	// Delete
	router.DELETE(path+"/:id", func(c *gin.Context) {
		if db.IsDeleted() {
			c.JSON(http.StatusBadRequest, gin.H{"message": db.GetCollectionName() + " collection deleted"})
			return
		}

		id := c.Param("id")
		if err := db.Delete(id); err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "Data deleted successfully"})
	})

	// Define an API endpoint to get all data
	router.GET(path+"/all-data", func(c *gin.Context) {
		if db.IsDeleted() {
			c.JSON(http.StatusBadRequest, gin.H{"message": db.GetCollectionName() + " collection deleted"})
			return
		}

		// Fetch all data from the database
		allData := db.GetAllData()

		// Serialize data to JSON
		responseData, _ := json.Marshal(allData)

		// Send the JSON response
		c.Data(http.StatusOK, "application/json; charset=utf-8", responseData)
	})

	router.GET(path+"/index-data", func(c *gin.Context) {
		if db.IsDeleted() {
			c.JSON(http.StatusBadRequest, gin.H{"message": db.GetCollectionName() + " collection deleted"})
			return
		}

		// Fetch all data from the database
		allData := db.GetIndexData()

		// Serialize data to JSON
		responseData, _ := json.Marshal(allData)

		// Send the JSON response
		c.Data(http.StatusOK, "application/json; charset=utf-8", responseData)
	})
}
