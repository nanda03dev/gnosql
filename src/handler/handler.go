package handler

import (
	"fmt"
	"gnosql/src/global_constants"
	"gnosql/src/in_memory_database"
	"gnosql/src/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

// @Summary      Create new database
// @Description  To create new database
// @Tags         database
// @Produce      json
// @Param        database  body in_memory_database.DatabaseCreateRequest  true  "Database"
// @Success      200 "database created successfully"
// @Success      400 "Database already exists"
// @Router       /database/add [post]
func CreateDatabase(c *gin.Context, gnoSQL *in_memory_database.GnoSQL) {
	var requestBody in_memory_database.DatabaseCreateRequest

	if err := c.BindJSON(&requestBody); err != nil {
		c.JSON(http.StatusBadRequest, global_constants.ERROR_WHILE_BINDING_JSON)
		return
	}

	result, err := service.CreateDatabase(gnoSQL, requestBody.DatabaseName, requestBody.Collections)

	c.JSON(GetResponse(result, err))
}

// @Summary      Connect to database
// @Description  Connect database
// @Tags         database
// @Produce      json
// @Param        database  body in_memory_database.DatabaseCreateRequest  true  "Database"
// @Success      200 in_memory_database.DatabaseConnectResult
// @Success      400 "Something went wrong"
// @Router       /database/connect [post]
func ConnectDatabase(c *gin.Context, gnoSQL *in_memory_database.GnoSQL) {
	var requestBody in_memory_database.DatabaseCreateRequest

	if err := c.BindJSON(&requestBody); err != nil {
		c.JSON(http.StatusBadRequest, global_constants.ERROR_WHILE_BINDING_JSON)
		return
	}

	result := service.ConnectDatabase(gnoSQL, requestBody.DatabaseName, requestBody.Collections)

	c.JSON(GetResponse(result, nil))
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
	var requestBody in_memory_database.DatabaseDeleteRequest

	if err := c.BindJSON(&requestBody); err != nil {
		c.JSON(http.StatusBadRequest, global_constants.ERROR_WHILE_BINDING_JSON)
		return
	}

	result, err := service.DeleteDatabase(gnoSQL, requestBody.DatabaseName)

	c.JSON(GetResponse(result, err))
}

// @Summary      Get all database
// @Description  To get all database.
// @Tags         database
// @Produce      json
// @Success      200 {array} string
// @Router       /database/get-all [get]
func GetAllDatabases(c *gin.Context, gnoSQL *in_memory_database.GnoSQL) {
	result, err := service.GetAllDatabase(gnoSQL)
	c.JSON(GetResponse(result, err))
}

// @Summary      Load database to disk
// @Description  Load database to disk.
// @Tags         database
// @Produce      json
// @Success      200 {array} string
// @Router       /database/load-to-disk [get]
func LoadDatabaseToDisk(c *gin.Context, gnoSQL *in_memory_database.GnoSQL) {
	result, err := service.LoadToDisk(gnoSQL)

	c.JSON(GetResponse(result, err))
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
	var requestBody in_memory_database.CollectionCreateRequest

	if err := c.BindJSON(&requestBody); err != nil {
		c.JSON(http.StatusBadRequest, global_constants.ERROR_WHILE_BINDING_JSON)
		return
	}

	result, err := service.CreateCollections(gnoSQL, requestBody.DatabaseName, requestBody.Collections)

	c.JSON(GetResponse(result, err))
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

	var requestBody in_memory_database.CollectionDeleteRequest

	if err := c.BindJSON(&requestBody); err != nil {
		c.JSON(http.StatusBadRequest, global_constants.ERROR_WHILE_BINDING_JSON)
		return
	}

	result, err := service.DeleteCollections(gnoSQL, requestBody.DatabaseName, requestBody.Collections)

	c.JSON(GetResponse(result, err))
}

// @Summary      Get all collections
// @Description  To get all collections
// @Tags         collection
// @Produce      json
// @Param        databaseName  path      string  true  "databaseName"
// @Success      200 {array} string
// @Router       /collection/{databaseName}/get-all [get]
func GetAllCollections(c *gin.Context, gnoSQL *in_memory_database.GnoSQL) {
	var requestBody in_memory_database.CollectionGetAllRequest

	if err := c.BindJSON(&requestBody); err != nil {
		c.JSON(http.StatusBadRequest, global_constants.ERROR_WHILE_BINDING_JSON)
		return
	}

	result, err := service.GetAllCollections(gnoSQL, requestBody.DatabaseName)

	c.JSON(GetResponse(result, err))
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

	var requestBody in_memory_database.CollectionStatsRequest

	if err := c.BindJSON(&requestBody); err != nil {
		c.JSON(http.StatusBadRequest, global_constants.ERROR_WHILE_BINDING_JSON)
		return
	}

	result, err := service.GetCollectionStats(gnoSQL, requestBody.DatabaseName, requestBody.CollectionName)

	c.JSON(GetResponse(result, err))

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

	var requestBody in_memory_database.DocumentCreateRequest

	if err := c.BindJSON(&requestBody); err != nil {
		c.JSON(http.StatusBadRequest, global_constants.ERROR_WHILE_BINDING_JSON)
		return
	}

	result, err := service.DocumentCreate(gnoSQL, requestBody.DatabaseName, requestBody.CollectionName, requestBody.Document)

	c.JSON(GetResponse(result, err))
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
	var requestBody in_memory_database.DocumentReadRequest

	if err := c.BindJSON(&requestBody); err != nil {
		c.JSON(http.StatusBadRequest, global_constants.ERROR_WHILE_BINDING_JSON)
		return
	}

	result, err := service.DocumentRead(gnoSQL, requestBody.DatabaseName, requestBody.CollectionName, requestBody.DocId)

	c.JSON(GetResponse(result, err))
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
	var requestBody in_memory_database.DocumentFilterRequest

	if err := c.BindJSON(&requestBody); err != nil {
		c.JSON(http.StatusBadRequest, global_constants.ERROR_WHILE_BINDING_JSON)
		return
	}

	result, err := service.DocumentFilter(gnoSQL, requestBody.DatabaseName, requestBody.CollectionName, requestBody.Filter)
	fmt.Printf("\n result %v \n err %v", result, err)
	c.JSON(GetResponse(result, err))
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
	var requestBody in_memory_database.DocumentUpdateRequest

	if err := c.BindJSON(&requestBody); err != nil {
		c.JSON(http.StatusBadRequest, global_constants.ERROR_WHILE_BINDING_JSON)
		return
	}

	result, err := service.DocumentUpdate(gnoSQL, requestBody.DatabaseName, requestBody.CollectionName, requestBody.DocId, requestBody.Document)

	c.JSON(GetResponse(result, err))
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
	var requestBody in_memory_database.DocumentDeleteRequest

	if err := c.BindJSON(&requestBody); err != nil {
		c.JSON(http.StatusBadRequest, global_constants.ERROR_WHILE_BINDING_JSON)
		return
	}

	result, err := service.DocumentDelete(gnoSQL, requestBody.DatabaseName, requestBody.CollectionName, requestBody.DocId)

	c.JSON(GetResponse(result, err))
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
	var requestBody in_memory_database.DocumentGetAllRequest

	if err := c.BindJSON(&requestBody); err != nil {
		c.JSON(http.StatusBadRequest, global_constants.ERROR_WHILE_BINDING_JSON)
		return
	}

	result, err := service.DocumentGetAll(gnoSQL, requestBody.DatabaseName, requestBody.CollectionName)

	c.JSON(GetResponse(result, err))
}

func GetResponse(result interface{}, err error) (int, interface{}) {
	if err == nil {
		return http.StatusOK, result
	} else {
		return http.StatusBadRequest, err
	}
}
