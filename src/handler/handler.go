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
// @Description  To create a new database
// @Tags         database
// @Accept       json
// @Produce      json
// @Param        requestBody  body  in_memory_database.DatabaseCreateRequest  true  "Database creation request containing databaseName and collections"
// @Success      200  {object}  in_memory_database.DatabaseCreateResult  "Database created successfully"
// @Failure      400  {object}  map[string]string  "Database already exists or error while binding JSON"
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
// @Description  Connect to an existing database
// @Tags         database
// @Accept       json
// @Produce      json
// @Param        requestBody  body in_memory_database.DatabaseCreateRequest true "databaseName, collections"
// @Success      200  {object}  in_memory_database.DatabaseConnectResult  "Connected successfully"
// @Failure      400  {object}  map[string]string  "Something went wrong or error while binding JSON"
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
// @Description  To delete a database
// @Tags         database
// @Accept       json
// @Produce      json
// @Param        requestBody  body  in_memory_database.DatabaseDeleteRequest true "databaseName"
// @Success      200  {object}  in_memory_database.DatabaseDeleteResult  "Database deleted successfully"
// @Failure      400  {object}  map[string]string  "Unexpected error while deleting database or error while binding JSON"
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

// @Summary      Get all databases
// @Description  Retrieve a list of all databases
// @Tags         database
// @Produce      json
// @Success      200  {array}  in_memory_database.DatabaseGetAllResult  "List of all databases"
// @Failure      500  {object}  map[string]string  "Internal server error"
// @Router       /database/get-all [get]
func GetAllDatabases(c *gin.Context, gnoSQL *in_memory_database.GnoSQL) {
	result, err := service.GetAllDatabase(gnoSQL)
	c.JSON(GetResponse(result, err))
}

// @Summary      Load database to disk
// @Description  Load database to disk for persistence
// @Tags         database
// @Produce      json
// @Success      200  {object}  map[string]string  "Database loaded to disk successfully"
// @Failure      500  {object}  map[string]string  "Error loading database to disk"
// @Router       /database/load-to-disk [get]
func LoadDatabaseToDisk(c *gin.Context, gnoSQL *in_memory_database.GnoSQL) {
	result, err := service.LoadToDisk(gnoSQL)

	c.JSON(GetResponse(result, err))
}

// @Summary      Create new collection
// @Description  To create a new collection in a specific database
// @Tags         collection
// @Accept       json
// @Produce      json
// @Param        requestBody  body in_memory_database.CollectionCreateRequest true "databaseName, collections"
// @Success      200  {object}  in_memory_database.CollectionCreateResult  "Collection created successfully"
// @Failure      400  {object}  map[string]string  "Collection already exists or error while binding JSON"
// @Router       /collection/add [post]
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
// @Description  To delete a collection from a specific database
// @Tags         collection
// @Accept       json
// @Produce      json
// @Param        requestBody  body  in_memory_database.CollectionDeleteRequest true  "databaseName, collections"
// @Success      200  {object}  in_memory_database.CollectionDeleteResult  "Collection deleted successfully"
// @Failure      400  {object}  map[string]string  "Unexpected error while deleting collection or error while binding JSON"
// @Router       /collection/delete [post]
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
// @Description  Retrieve all collections from a specific database
// @Tags         collection
// @Produce      json
// @Param        requestBody  body  in_memory_database.CollectionGetAllRequest true "databaseName"
// @Success      200  {array}   in_memory_database.CollectionGetAllResult  "List of all collections"
// @Failure      400  {object}  map[string]string  "Error while fetching collections or invalid database"
// @Router       /collection/get-all [post]
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
// @Description  Retrieve statistics for a specific collection in a database
// @Tags         collection
// @Produce      json
// @Param        requestBody  body  in_memory_database.CollectionStatsRequest true "databaseName, collectionName"
// @Success      200  {object}  in_memory_database.IndexMap  "Collection statistics"
// @Failure      400  {object}  map[string]string  "Database or Collection not found or deleted"
// @Router       /collection/stats [post]
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
// @Param        requestBody  body  in_memory_database.DocumentCreateRequest true  "databaseName, collectionName"
// @Success      200 "Document created successfully"
// @Success      400 "Database/Collection deleted"
// @Router       /document/add [post]
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
// @Param        requestBody  body  in_memory_database.DocumentReadRequest true "databaseName, collectionName, docId"
// @Success      200 {object}  in_memory_database.Document
// @Success   	 400 "Database/Collection deleted"
// @Router       /document/{id} [get]
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
// @Param        requestBody  body   in_memory_database.DocumentFilterRequest true "databaseName, collectionName, filter"
// @Success      200 {array}  in_memory_database.Document
// @Success   	 400 "Database/Collection deleted"
// @Router       /document/filter [post]
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
// @Param        requestBody  body  in_memory_database.DocumentUpdateRequest true "databaseName, collectionName, docId, document"
// @Success      200 {object} in_memory_database.Document
// @Success      400 "Database/Collection deleted"
// @Router       /document/{id} [post]
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
// @Param        requestBody  body  in_memory_database.DocumentDeleteRequest true "databaseName, collectionName, docId"
// @Success      200 {object} in_memory_database.Document
// @Success      400 "Database/Collection deleted"
// @Router       /document/{id} [post]
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
// @Param        requestBody  body   in_memory_database.DocumentGetAllRequest true "databaseName, collectionName"
// @Success      200 {array}  in_memory_database.Document
// @Success   	 400 "Database/Collection deleted"
// @Router       /document/all-data [post]
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
