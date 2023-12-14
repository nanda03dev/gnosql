package handler

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"gnosql/src/in_memory_database"
	"net/http"
)

// @Summary      Create new database
// @Description  To create new database
// @Tags         database
// @Produce      json
// @Param        database  body router.DatabaseRequestInput  true  "Database"
// @Success      200 "database created successfully"
// @Success      400 "Database already exists"
// @Router       /database/add [post]
func CreateDatabase(c *gin.Context, gnoSQL *in_memory_database.GnoSQL) {
	var value in_memory_database.MapInterface

	if err := c.BindJSON(&value); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	databaseName := value["databaseName"].(string)
	collectionsInterface := make([]interface{}, 0)

	if collections, exists := value["collections"]; exists {
		collectionsInterface = collections.([]interface{})
	}

	if dbExists := gnoSQL.GetDatabase(databaseName); dbExists != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Database already exists"})
		return
	}

	gnoSQL.CreateDatabase(databaseName, in_memory_database.ConvertToCollectionInputs(collectionsInterface))

	c.JSON(http.StatusCreated, gin.H{"data": "database created successfully"})
}

// @Summary      Delete database
// @Description  To delete database
// @Tags         database
// @Produce      json
// @Param        database  body router.DatabaseRequestInput  true  "Database"
// @Success      200 "database deleted successfully"
// @Success      400 "Unexpected error while delete database"
// @Router       /database/delete [post]
func DeleteDatabase(c *gin.Context, gnoSQL *in_memory_database.GnoSQL) {
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
}

// @Summary      Get all database
// @Description  To get all database.
// @Tags         database
// @Produce      json
// @Success      200 {array} string
// @Router       /database/get-all [get]
func GetAllDatabases(c *gin.Context, gnoSQL *in_memory_database.GnoSQL) {
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
}

// @Summary      Load database to disk
// @Description  Load database to disk.
// @Tags         database
// @Produce      json
// @Success      200 {array} string
// @Router       /database/load-to-disk [get]
func LoadDatabaseToDisk(c *gin.Context, gnoSQL *in_memory_database.GnoSQL) {
	go gnoSQL.WriteAllDatabases()
	c.JSON(http.StatusOK, gin.H{"status": "database to file disk started."})
}

// @Summary      Create new collection
// @Description  To create new collection.
// @Tags         collection
// @Produce      json
// @Param        databaseName  path      string  true  "databaseName"
// @Param        collection  body in_memory_database.CollectionInput  true  "Collection"
// @Success      200 "collection created successfully"
// @Success      400 "collection already exists"
// @Router       /collection/{databaseName}/add [post]
func CreateCollection(c *gin.Context, db *in_memory_database.Database) {

	var CollectionsInterface []interface{}

	if err := c.BindJSON(&CollectionsInterface); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	db.CreateCollections(in_memory_database.ConvertToCollectionInputs(CollectionsInterface))

	c.JSON(http.StatusCreated, gin.H{"data": "collection created successfully"})
}

// @Summary      Delete collection
// @Description  To delete collection
// @Tags         collection
// @Produce      json
// @Param        databaseName  path      string  true  "databaseName"
// @Param        collection  body router.DatabaseRequestInput  true  "collection"
// @Success      200 "collection deleted successfully"
// @Success      400 "Unexpected error while delete collection"
// @Router       /collection/{databaseName}/delete [post]
func DeleteCollection(c *gin.Context, db *in_memory_database.Database) {

	var value map[string][]string

	if err := c.BindJSON(&value); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// var indexKeys = make([]string, 0)
	if collections, ok := value["collections"]; ok {
		db.DeleteCollections(collections)

	}

	c.JSON(http.StatusOK, gin.H{"data": "successfully deleted"})
}

// @Summary      Get all collections
// @Description  To get all collections
// @Tags         collection
// @Produce      json
// @Param        databaseName  path      string  true  "databaseName"
// @Success      200 {array} string
// @Router       /collection/{databaseName}/get-all [get]
func GetAllCollections(c *gin.Context, db *in_memory_database.Database) {

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
}

// @Summary      Collection stats
// @Description  Collection stats
// @Tags         collection
// @Produce      json
// @Param        databaseName  path      string  true  "databaseName"
// @Param        collectionName  path      string  true  "collectionName"
// @Success      200 {object}  in_memory_database.Index
// @Success   	 400 "Database/Collection deleted"
// @Router       /collection/{databaseName}/{collectionName}/stats [get]
func CollectionStats(c *gin.Context, db *in_memory_database.Database, collection *in_memory_database.Collection) {
	// Serialize data to JSON
	responseData, _ := json.Marshal(collection.Stats())

	// Send the JSON response
	c.Data(http.StatusOK, "application/json; charset=utf-8", responseData)
}

// @Summary      Create new document
// @Description  To create new document
// @Tags         document
// @Produce      json
// @Param        databaseName  path      string  true  "databaseName"
// @Param        collectionName  path      string  true  "collectionName"
// @Param        document  body in_memory_database.Document  true  "Document"
// @Success      200 "Document created successfully"
// @Success      400 "Database/Collection deleted"
// @Router       /document/{databaseName}/{collectionName}/ [post]
func CreateDocument(c *gin.Context, db *in_memory_database.Database, collection *in_memory_database.Collection) {

	var value map[string]interface{}
	if err := c.BindJSON(&value); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result := collection.Create(value)

	c.JSON(http.StatusCreated, gin.H{"data": result})
}

// @Summary      Read by id
// @Description  Read document by id.
// @Tags         document
// @Produce      json
// @Param        databaseName  path      string  true  "databaseName"
// @Param        collectionName  path      string  true  "collectionName"
// @Param        id  path      string  true  "search document by id"
// @Success      200 {object}  in_memory_database.Document
// @Success   	 400 "Database/Collection deleted"
// @Router       /document/{databaseName}/{collectionName}/{id} [get]
func ReadDocument(c *gin.Context, db *in_memory_database.Database, collection *in_memory_database.Collection) {

	var value map[string]interface{}
	if err := c.BindJSON(&value); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result := collection.Create(value)

	c.JSON(http.StatusOK, gin.H{"data": result})
}

// @Summary      Filter document
// @Description  Filter document
// @Tags         document
// @Produce      json
// @Param        databaseName  path      string  true  "databaseName"
// @Param        collectionName  path      string  true  "collectionName"
// @Param        document body in_memory_database.GenericKeyValue  true  "GenericKeyValue"
// @Success      200 {array}  in_memory_database.Document
// @Success   	 400 "Database/Collection deleted"
// @Router       /document/{databaseName}/{collectionName}/filter [post]
func FilterDocument(c *gin.Context, db *in_memory_database.Database, collection *in_memory_database.Collection) {

	var value in_memory_database.MapInterface

	if err := c.BindJSON(&value); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result := collection.Filter(value)

	c.JSON(http.StatusOK, gin.H{"data": result})
}

// @Summary      Update document
// @Description  To update document
// @Tags         document
// @Produce      json
// @Param        databaseName  path      string  true  "databaseName"
// @Param        collectionName  path      string  true  "collectionName"
// @Param        id  path      string  true  "update document by id"
// @Param        document  body in_memory_database.Document  true  "Document"
// @Success      200 {object} in_memory_database.Document
// @Success      400 "Database/Collection deleted"
// @Router       /document/{databaseName}/{collectionName}/{id} [put]
func UpdateDocument(c *gin.Context, db *in_memory_database.Database, collection *in_memory_database.Collection) {

	id := c.Param("id")
	var value map[string]interface{}
	if err := c.BindJSON(&value); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result := collection.Update(id, value)

	if result == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "document not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": result})
}

// @Summary      Delete document
// @Description  To delete document
// @Tags         document
// @Produce      json
// @Param        databaseName  path      string  true  "databaseName"
// @Param        collectionName  path      string  true  "collectionName"
// @Param        id  path      string  true  "delete document by id"
// @Success      200 {object} in_memory_database.Document
// @Success      400 "Database/Collection deleted"
// @Router       /document/{databaseName}/{collectionName}/{id} [delete]
func DeleteDocument(c *gin.Context, db *in_memory_database.Database, collection *in_memory_database.Collection) {

	id := c.Param("id")
	if err := collection.Delete(id); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Data deleted successfully"})
}

// @Summary      Read all document
// @Description  Read all document
// @Tags         document
// @Produce      json
// @Param        databaseName  path      string  true  "databaseName"
// @Param        collectionName  path      string  true  "collectionName"
// @Success      200 {array}  in_memory_database.Document
// @Success   	 400 "Database/Collection deleted"
// @Router       /document/{databaseName}/{collectionName}/all-data [get]
func ReadAllDocument(c *gin.Context, db *in_memory_database.Database, collection *in_memory_database.Collection) {

	// Fetch all data from the database
	allData := collection.GetAllData()

	// Serialize data to JSON
	responseData, _ := json.Marshal(allData)

	// Send the JSON response
	c.Data(http.StatusOK, "application/json; charset=utf-8", responseData)
}
