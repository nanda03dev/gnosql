package router

import (
	"encoding/json"
	"fmt"
	"gnosql/src/in_memory_database"
	"net/http"

	"github.com/gin-gonic/gin"
)

func LoadGnoSQLAndRoutes(router *gin.Engine, gnoSQL *in_memory_database.GnoSQL) {

	var databases []*in_memory_database.Database = gnoSQL.LoadAllDatabases()

	for _, database := range databases {
		GenerateCollectionRoutes(router, database)
		GenerateEntityRoutes(router, database, database.Collections)
	}

}

func GenerateDatabaseRoutes(router *gin.Engine, gnoSQL *in_memory_database.GnoSQL) {
	router.POST("/add-database", func(c *gin.Context) {
		var value map[string]interface{}

		if err := c.BindJSON(&value); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		databaseName := value["databaseName"].(string)

		if dbExists := gnoSQL.GetDatabase(databaseName); dbExists != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Database already exists"})
			return
		}

		db := gnoSQL.CreateDatabase(databaseName)

		GenerateCollectionRoutes(router, db)

		c.JSON(http.StatusCreated, gin.H{"data": "database deleted successfully"})

	})

	router.POST("/delete-database", func(c *gin.Context) {
		var value map[string]interface{}

		if err := c.BindJSON(&value); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		db := gnoSQL.GetDatabase(value["databaseName"].(string))

		if db == nil {
			c.JSON(http.StatusBadRequest, gin.H{"data": "Unexpected error while delete database"})
			return
		}

		gnoSQL.DeleteDatabase(db)

		c.JSON(http.StatusOK, gin.H{"data": "database deleted successfully"})

	})

	router.GET("/get-all-databases", func(c *gin.Context) {

		// Fetch all data from the database
		databaseNames := make([]string, 0)

		for _, database := range gnoSQL.Databases {
			if !database.IsDeleted {
				databaseNames = append(databaseNames, database.DatabaseName)
			}
		}
		// Serialize data to JSON
		responseData, _ := json.Marshal(databaseNames)

		// Send the JSON response
		c.Data(http.StatusOK, "application/json; charset=utf-8", responseData)
	})
}

func GenerateCollectionRoutes(router *gin.Engine, db *in_memory_database.Database) {

	path := fmt.Sprintf("/%s", db.DatabaseName)

	router.POST(path+"/add-collections", func(c *gin.Context) {
		if db.IsDeleted {
			c.JSON(http.StatusBadRequest, gin.H{"message": db.DatabaseName + " database deleted"})
			return
		}

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

		addedCollectionInstance := db.CreateCollections(collections)

		GenerateEntityRoutes(router, db, addedCollectionInstance)

		c.JSON(http.StatusCreated, gin.H{"data": "collection created successfully"})
	})

	router.DELETE(path+"/delete-collections", func(c *gin.Context) {
		if db.IsDeleted {
			c.JSON(http.StatusBadRequest, gin.H{"message": db.DatabaseName + " database deleted"})
			return
		}

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

	router.GET(path+"/get-all-collections", func(c *gin.Context) {
		if db.IsDeleted {
			c.JSON(http.StatusBadRequest, gin.H{"message": db.DatabaseName + " database deleted"})
			return
		}

		// Fetch all data from the database
		allCollections := db.Collections
		collections := make([]string, 0)

		for _, collection := range allCollections {
			collections = append(collections, collection.CollectionName)
		}
		// Serialize data to JSON
		responseData, _ := json.Marshal(collections)

		// Send the JSON response
		c.Data(http.StatusOK, "application/json; charset=utf-8", responseData)
	})

}

func GenerateEntityRoutes(router *gin.Engine, db *in_memory_database.Database, collections []*in_memory_database.Collection) {
	for _, collection := range collections {
		generateRoutes(router, db, collection)
	}
}

func generateRoutes(router *gin.Engine, db *in_memory_database.Database, collection *in_memory_database.Collection) {
	path := fmt.Sprintf("/%s/%s", db.DatabaseName, collection.CollectionName)

	// Create
	router.POST(path, func(c *gin.Context) {

		if db.IsDeleted {
			c.JSON(http.StatusBadRequest, gin.H{"message": db.DatabaseName + " database deleted"})
			return
		}

		if collection.IsDeleted {
			c.JSON(http.StatusBadRequest, gin.H{"message": collection.CollectionName + " collection deleted"})
			return
		}

		var value map[string]interface{}
		if err := c.BindJSON(&value); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		result := collection.Create(value)

		c.JSON(http.StatusCreated, gin.H{"data": result})
	})

	// Read
	router.GET(path+"/:id", func(c *gin.Context) {
		if db.IsDeleted {
			c.JSON(http.StatusBadRequest, gin.H{"message": db.DatabaseName + " database deleted"})
			return
		}

		if collection.IsDeleted {
			c.JSON(http.StatusBadRequest, gin.H{"message": collection.CollectionName + " collection deleted"})
			return
		}

		id := c.Param("id")
		value := collection.Read(id)
		c.JSON(http.StatusOK, gin.H{"data": value})
	})

	// Read by index
	router.POST(path+"/filter", func(c *gin.Context) {
		if db.IsDeleted {
			c.JSON(http.StatusBadRequest, gin.H{"message": db.DatabaseName + " database deleted"})
			return
		}

		if collection.IsDeleted {
			c.JSON(http.StatusBadRequest, gin.H{"message": collection.CollectionName + " collection deleted"})
			return
		}

		var value []in_memory_database.GenericKeyValue

		if err := c.BindJSON(&value); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		result := collection.Filter(value)

		c.JSON(http.StatusCreated, gin.H{"data": result})
	})

	// Read by index
	router.POST(path+"/filterbyindex", func(c *gin.Context) {
		if db.IsDeleted {
			c.JSON(http.StatusBadRequest, gin.H{"message": db.DatabaseName + " database deleted"})
			return
		}

		if collection.IsDeleted {
			c.JSON(http.StatusBadRequest, gin.H{"message": collection.CollectionName + " collection deleted"})
			return
		}

		var value []in_memory_database.GenericKeyValue

		if err := c.BindJSON(&value); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		result := collection.FilterByIndexKey(value)

		c.JSON(http.StatusCreated, gin.H{"data": result})
	})

	// Update
	router.PUT(path+"/:id", func(c *gin.Context) {
		if db.IsDeleted {
			c.JSON(http.StatusBadRequest, gin.H{"message": db.DatabaseName + " database deleted"})
			return
		}

		if collection.IsDeleted {
			c.JSON(http.StatusBadRequest, gin.H{"message": collection.CollectionName + " collection deleted"})
			return
		}

		id := c.Param("id")
		var value map[string]interface{}
		if err := c.BindJSON(&value); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		result := collection.Update(id, value)

		c.JSON(http.StatusOK, gin.H{"data": result})
	})

	// Delete
	router.DELETE(path+"/:id", func(c *gin.Context) {
		if db.IsDeleted {
			c.JSON(http.StatusBadRequest, gin.H{"message": db.DatabaseName + " database deleted"})
			return
		}

		if collection.IsDeleted {
			c.JSON(http.StatusBadRequest, gin.H{"message": collection.CollectionName + " collection deleted"})
			return
		}

		id := c.Param("id")
		if err := collection.Delete(id); err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "Data deleted successfully"})
	})

	// Define an API endpoint to get all data
	router.GET(path+"/all-data", func(c *gin.Context) {
		if db.IsDeleted {
			c.JSON(http.StatusBadRequest, gin.H{"message": db.DatabaseName + " database deleted"})
			return
		}

		if collection.IsDeleted {
			c.JSON(http.StatusBadRequest, gin.H{"message": collection.CollectionName + " collection deleted"})
			return
		}

		// Fetch all data from the database
		allData := collection.GetAllData()

		// Serialize data to JSON
		responseData, _ := json.Marshal(allData)

		// Send the JSON response
		c.Data(http.StatusOK, "application/json; charset=utf-8", responseData)
	})

	router.GET(path+"/index-data", func(c *gin.Context) {
		if db.IsDeleted {
			c.JSON(http.StatusBadRequest, gin.H{"message": db.DatabaseName + " database deleted"})
			return
		}

		if collection.IsDeleted {
			c.JSON(http.StatusBadRequest, gin.H{"message": collection.CollectionName + " collection deleted"})
			return
		}

		// Fetch all data from the database
		allData := collection.GetIndexData()

		// Serialize data to JSON
		responseData, _ := json.Marshal(allData)

		// Send the JSON response
		c.Data(http.StatusOK, "application/json; charset=utf-8", responseData)
	})
}
