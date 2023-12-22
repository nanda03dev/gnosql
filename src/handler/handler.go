package handler

import (
	"encoding/json"
	"fmt"
	"gnosql/src/in_memory_database"
	"gnosql/src/utils"
	"net/http"

	"github.com/gin-gonic/gin"
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

	fmt.Printf("\n value %v \n ", value)

	databaseName := value["DatabaseName"].(string)
	collectionsInterface := make([]interface{}, 0)

	if collections, exists := value["Collections"]; exists {
		collectionsInterface = collections.([]interface{})
	}

	if dbExists := gnoSQL.GetDB(databaseName); dbExists != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": "Database already exists"})
		return
	}

	gnoSQL.CreateDB(databaseName, in_memory_database.ConvertToCollectionInputs(collectionsInterface))

	c.JSON(http.StatusCreated, gin.H{"Data": "database created successfully"})
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

	db := gnoSQL.GetDB(value["DatabaseName"].(string))

	if db == nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": "Unexpected error while delete database"})
		return
	}

	gnoSQL.DeleteDB(db)

	c.JSON(http.StatusOK, gin.H{"Data": "database deleted successfully"})
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
		databaseNames = append(databaseNames, database.DatabaseName)
	}

	// Send the JSON response
	c.JSON(http.StatusOK, gin.H{"Data": databaseNames})
}

// @Summary      Load database to disk
// @Description  Load database to disk.
// @Tags         database
// @Produce      json
// @Success      200 {array} string
// @Router       /database/load-to-disk [get]
func LoadDatabaseToDisk(c *gin.Context, gnoSQL *in_memory_database.GnoSQL) {
	go gnoSQL.WriteAllDBs()
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

	db.CreateColls(in_memory_database.ConvertToCollectionInputs(CollectionsInterface))

	c.JSON(http.StatusCreated, gin.H{"Data": utils.COLLECTION_CREATE_SUCCESS_MSG})
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
	if collections, ok := value["Collections"]; ok {
		db.DeleteColls(collections)
	}

	c.JSON(http.StatusOK, gin.H{"Data": "collection successfully deleted"})
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

	// Send the JSON response
	c.JSON(http.StatusOK, gin.H{"Data": collections})
}

// @Summary      Collection stats
// @Description  Collection stats
// @Tags         collection
// @Produce      json
// @Param        databaseName  path      string  true  "databaseName"
// @Param        collectionName  path      string  true  "collectionName"
// @Success      200 {object}  in_memory_database.IndexMap
// @Success   	 400 "Database/Collection deleted"
// @Router       /collection/{databaseName}/{collectionName}/stats [get]
func CollectionStats(c *gin.Context, db *in_memory_database.Database, collection *in_memory_database.Collection) {
	// Send the JSON response
	c.JSON(http.StatusOK, gin.H{"Data": collection.Stats()})
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

	var value in_memory_database.Document
	if err := c.BindJSON(&value); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	uniqueUuid := utils.Generate16DigitUUID()

	value["id"] = uniqueUuid

	var createEvent in_memory_database.Event = in_memory_database.Event{
		Type:      utils.EVENT_CREATE,
		EventData: value,
	}

	collection.EventChannel <- createEvent

	c.JSON(http.StatusCreated, gin.H{"Data": value})
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
	id := c.Param("id")
	result := collection.Read(id)
	c.JSON(http.StatusOK, gin.H{"Data": result})
}

// @Summary      Filter document
// @Description  Filter document
// @Tags         document
// @Produce      json
// @Param        databaseName  path      string  true  "databaseName"
// @Param        collectionName  path      string  true  "collectionName"
// @Param        document body in_memory_database.MapInterface  true  "MapInterface"
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

	c.JSON(http.StatusOK, gin.H{"Data": result})
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
	var value in_memory_database.Document
	if err := c.BindJSON(&value); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var existingDocument in_memory_database.Document = collection.Read(id)

	if existingDocument == nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": "document not found"})
		return
	}

	for key, value := range value {
		existingDocument[key] = value
	}

	var updateEvent in_memory_database.Event = in_memory_database.Event{
		Type:      utils.EVENT_UPDATE,
		Id:        id,
		EventData: existingDocument,
	}

	collection.EventChannel <- updateEvent

	c.JSON(http.StatusOK, gin.H{"Data": existingDocument})
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

	if err := collection.Read(id); err == nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": "document not found"})
		return
	}

	var deleteEvent in_memory_database.Event = in_memory_database.Event{
		Type: utils.EVENT_DELETE,
		Id:   id,
	}

	collection.EventChannel <- deleteEvent

	c.JSON(http.StatusOK, gin.H{"Data": "Data deleted successfully"})
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
