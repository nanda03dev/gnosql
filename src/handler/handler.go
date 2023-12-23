package handler

import (
	"github.com/gin-gonic/gin"
	"gnosql/src/in_memory_database"
	"gnosql/src/service"
	"gnosql/src/utils"
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
	var result = in_memory_database.DatabaseCreateResult{}

	if err := c.BindJSON(&value); err != nil {
		result.Error = utils.ERROR_WHILE_BINDING_JSON
		c.JSON(http.StatusBadRequest, result)
		return
	}

	CollectionsInterface := make([]interface{}, 0)

	if collections, exists := value["Collections"]; exists {
		CollectionsInterface = collections.([]interface{})
	}

	result = service.ServiceCreateDatabase(gnoSQL, value["DatabaseName"].(string), in_memory_database.ConvertToCollectionInputs(CollectionsInterface))

	c.JSON(http.StatusCreated, result)
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
	var value map[string]string

	var result = in_memory_database.DatabaseDeleteResult{}

	if err := c.BindJSON(&value); err != nil {
		result.Error = utils.ERROR_WHILE_BINDING_JSON
		c.JSON(http.StatusBadRequest, result)
		return
	}

	result = service.ServiceDeleteDatabase(gnoSQL, value["DatabaseName"])

	c.JSON(http.StatusOK, result)
}

// @Summary      Get all database
// @Description  To get all database.
// @Tags         database
// @Produce      json
// @Success      200 {array} string
// @Router       /database/get-all [get]
func GetAllDatabases(c *gin.Context, gnoSQL *in_memory_database.GnoSQL) {

	result := service.ServiceGetAllDatabase(gnoSQL)

	// Send the JSON response
	c.JSON(http.StatusOK, result)
}

// @Summary      Load database to disk
// @Description  Load database to disk.
// @Tags         database
// @Produce      json
// @Success      200 {array} string
// @Router       /database/load-to-disk [get]
func LoadDatabaseToDisk(c *gin.Context, gnoSQL *in_memory_database.GnoSQL) {
	result := service.ServiceLoadToDisk(gnoSQL)

	c.JSON(http.StatusOK, result)
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
func CreateCollection(c *gin.Context, gnoSQL *in_memory_database.GnoSQL) {
	var CollectionsInterface []interface{}
	var result = in_memory_database.CollectionCreateResult{}

	if err := c.BindJSON(&CollectionsInterface); err != nil {
		result.Error = utils.ERROR_WHILE_BINDING_JSON
		c.JSON(http.StatusBadRequest, result)
		return
	}

	result = service.ServiceCreateCollections(gnoSQL, c.Param("DatabaseName"),
		in_memory_database.ConvertToCollectionInputs(CollectionsInterface))

	c.JSON(http.StatusCreated, result)
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
func DeleteCollection(c *gin.Context, gnoSQL *in_memory_database.GnoSQL) {

	var value map[string][]string
	var result = in_memory_database.CollectionDeleteResult{}

	if err := c.BindJSON(&value); err != nil {
		result.Error = utils.ERROR_WHILE_BINDING_JSON
		c.JSON(http.StatusBadRequest, result)
		return
	}

	result = service.ServiceDeleteCollections(gnoSQL, c.Param("DatabaseName"), value["Collections"])

	c.JSON(http.StatusOK, result)
}

// @Summary      Get all collections
// @Description  To get all collections
// @Tags         collection
// @Produce      json
// @Param        databaseName  path      string  true  "databaseName"
// @Success      200 {array} string
// @Router       /collection/{databaseName}/get-all [get]
func GetAllCollections(c *gin.Context, gnoSQL *in_memory_database.GnoSQL) {

	result := service.ServiceGetAllCollections(gnoSQL, c.Param("DatabaseName"))
	// Send the JSON response
	c.JSON(http.StatusOK, result)
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
func CollectionStats(c *gin.Context, gnoSQL *in_memory_database.GnoSQL) {
	result := service.ServiceGetCollectionStats(gnoSQL, c.Param("DatabaseName"), c.Param("CollectionName"))

	c.JSON(http.StatusOK, result)

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
func CreateDocument(c *gin.Context, gnoSQL *in_memory_database.GnoSQL) {

	var document in_memory_database.Document

	var result = in_memory_database.DocumentCreateResult{}

	if err := c.BindJSON(&document); err != nil {
		result.Error = utils.ERROR_WHILE_BINDING_JSON
		c.JSON(http.StatusBadRequest, result)
		return
	}

	result = service.ServiceDocumentCreate(gnoSQL, c.Param("DatabaseName"),
		c.Param("CollectionName"), document)

	c.JSON(http.StatusCreated, result)
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
func ReadDocument(c *gin.Context, gnoSQL *in_memory_database.GnoSQL) {

	result := service.ServiceDocumentRead(gnoSQL, c.Param("DatabaseName"),
		c.Param("CollectionName"), c.Param("id"))

	c.JSON(http.StatusOK, result)
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
func FilterDocument(c *gin.Context, gnoSQL *in_memory_database.GnoSQL) {
	var filter in_memory_database.MapInterface

	var result = in_memory_database.DocumentFilterResult{}

	if err := c.BindJSON(&filter); err != nil {
		result.Error = utils.ERROR_WHILE_BINDING_JSON
		c.JSON(http.StatusBadRequest, result)
		return
	}

	result = service.ServiceDocumentFilter(gnoSQL, c.Param("DatabaseName"),
		c.Param("CollectionName"), filter)

	c.JSON(http.StatusOK, result)
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
func UpdateDocument(c *gin.Context, gnoSQL *in_memory_database.GnoSQL) {
	var document in_memory_database.Document

	var result = in_memory_database.DocumentUpdateResult{}

	if err := c.BindJSON(&document); err != nil {
		result.Error = utils.ERROR_WHILE_BINDING_JSON
		c.JSON(http.StatusBadRequest, result)
		return
	}

	result = service.ServiceDocumentUpdate(gnoSQL, c.Param("DatabaseName"),
		c.Param("CollectionName"), c.Param("id"), document)

	c.JSON(http.StatusOK, result)
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
func DeleteDocument(c *gin.Context, gnoSQL *in_memory_database.GnoSQL) {

	result := service.ServiceDocumentDelete(gnoSQL, c.Param("DatabaseName"),
		c.Param("CollectionName"), c.Param("id"))

	c.JSON(http.StatusOK, result)
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
func ReadAllDocument(c *gin.Context, gnoSQL *in_memory_database.GnoSQL) {

	result := service.ServiceDocumentGetAll(gnoSQL, c.Param("DatabaseName"),
		c.Param("CollectionName"))

	c.JSON(http.StatusOK, result)
}
