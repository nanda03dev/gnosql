package grpc_handler

import (
	"context"
	"encoding/json"
	"fmt"

	// "fmt"
	pb "gnosql/proto"
	"gnosql/src/in_memory_database"
	"gnosql/src/utils"

	"github.com/gin-gonic/gin"

	// "log"
	"net/http"
)

type GnoSQLServer struct {
	pb.UnimplementedGnoSQLServiceServer

	GnoSQL *in_memory_database.GnoSQL
}

func (s *GnoSQLServer) CreateNewDatabase(ctx context.Context, req *pb.DatabaseCreateRequest) (*pb.DataStringResponse, error) {
	response := &pb.DataStringResponse{
		Data: "Database successfully created",
	}

	db := s.GnoSQL.GetDB(req.GetDatabaseName())

	if db != nil {
		response.Data = ""
		response.Error = "Database already exists"
		return response, nil
	}

	var collectionsInput []in_memory_database.CollectionInput

	for _, EachInput := range req.Collections {
		collectionInput := in_memory_database.CollectionInput{
			CollectionName: EachInput.GetCollectionName(),
			IndexKeys:      EachInput.GetIndexKeys(),
		}
		collectionsInput = append(collectionsInput, collectionInput)
	}

	fmt.Printf("\n collectionsInput %v \n", req.String())

	s.GnoSQL.CreateDB(req.GetDatabaseName(), collectionsInput)

	return response, nil

}

func (s *GnoSQLServer) DeleteDatabase(ctx context.Context, req *pb.DatabaseDeleteRequest) (*pb.DataStringResponse, error) {

	var response = &pb.DataStringResponse{
		Data: "Database deleted successfully",
	}

	db := s.GnoSQL.GetDB(req.GetDatabaseName())

	if db == nil {
		response.Error = "Database not found"
		return response, nil
	}

	s.GnoSQL.DeleteDB(db)

	return response, nil
}

func (s *GnoSQLServer) GetAllDatabases(ctx context.Context, req *pb.NoRequestBody) (*pb.DatabaseGetAllResult, error) {
	// Implement the logic similar to your HTTP handler
	databaseNames := make([]string, 0)

	for _, database := range s.GnoSQL.Databases {
		databaseNames = append(databaseNames, database.DatabaseName)
	}
	fmt.Printf("Database %v ", databaseNames)

	response := &pb.DatabaseGetAllResult{
		Data: databaseNames,
	}

	return response, nil
}

func (s *GnoSQLServer) LoadToDisk(ctx context.Context, req *pb.NoRequestBody) (*pb.DataStringResponse, error) {

	go s.GnoSQL.WriteAllDBs()

	response := &pb.DataStringResponse{
		Data: "database to file disk started.",
	}

	return response, nil
}

func (s *GnoSQLServer) CreateNewCollection(ctx context.Context, req *pb.CollectionCreateRequest) (*pb.DataStringResponse, error) {

	response := &pb.DataStringResponse{
		Data: utils.COLLECTION_CREATE_SUCCESS_MSG,
	}

	var db *in_memory_database.Database = s.GnoSQL.GetDB(req.GetDatabaseName())

	if db == nil {
		response.Data = ""
		response.Error = "Database not found"
		return response, nil
	}

	var collectionsInput []in_memory_database.CollectionInput

	for _, EachInput := range req.Collections {
		collectionInput := in_memory_database.CollectionInput{
			CollectionName: EachInput.GetCollectionName(),
			IndexKeys:      EachInput.GetIndexKeys(),
		}
		collectionsInput = append(collectionsInput, collectionInput)
	}

	db.CreateColls(collectionsInput)

	return response, nil

}

func (s *GnoSQLServer) DeleteCollections(ctx context.Context, req *pb.CollectionDeleteRequest) (*pb.DataStringResponse, error) {

	response := &pb.DataStringResponse{
		Data: utils.COLLECTION_DELETE_SUCCESS_MSG,
	}

	var db *in_memory_database.Database = s.GnoSQL.GetDB(req.GetDatabaseName())

	if db == nil {
		response.Data = ""
		response.Error = "Database not found"
		return response, nil
	}

	db.DeleteColls(req.GetCollections())

	return response, nil

}

func (s *GnoSQLServer) GetAllCollections(ctx context.Context, req *pb.CollectionGetAllRequest) (*pb.CollectionGetAllResult, error) {

	response := &pb.CollectionGetAllResult{}

	var db *in_memory_database.Database = s.GnoSQL.GetDB(req.GetDatabaseName())

	if db == nil {
		response.Error = "Database not found"
		return response, nil
	}

	allCollections := db.Collections

	collections := make([]string, 0)

	for _, collection := range allCollections {
		collections = append(collections, collection.CollectionName)
	}

	response.Data = collections

	return response, nil
}

func (s *GnoSQLServer) GetCollectionStats(ctx context.Context, req *pb.CollectionStatsRequest) (*pb.CollectionStatsResponse, error) {

	response := &pb.CollectionStatsResponse{}

	db, collection := s.GnoSQL.GetDatabaseAndCollection(req.GetDatabaseName(), req.GetCollectionName())

	if db == nil {
		response.Error = "Database not found"
		return response, nil
	}

	if collection == nil {
		response.Error = "Collection not found"
		return response, nil
	}

	stats := collection.Stats()

	response.Data = &pb.CollectionStats{
		CollectionName: stats.CollectionName,
		IndexKeys:      stats.IndexKeys,
		Documents:      int32(stats.Documents),
	}

	return response, nil
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
