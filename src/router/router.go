package router

import (
	"basic_database/src/in_memory_database"
	"encoding/json"
	"fmt"
	"net/http"
	"github.com/gin-gonic/gin"
)

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
