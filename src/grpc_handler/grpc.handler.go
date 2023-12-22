package grpc_handler

import (
	"context"
	"encoding/json"
	"github.com/gin-gonic/gin"
	pb "gnosql/proto"
	"gnosql/src/in_memory_database"
	"gnosql/src/utils"
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

	db := s.GnoSQL.GetDB(req.DatabaseName)

	if db != nil {
		response.Data = ""
		response.Error = "Database already exists"
		return response, nil
	}

	var collectionsInput []in_memory_database.CollectionInput

	for _, EachInput := range req.Collections {
		collectionInput := in_memory_database.CollectionInput{
			CollectionName: EachInput.CollectionName,
			IndexKeys:      EachInput.GetIndexKeys(),
		}
		collectionsInput = append(collectionsInput, collectionInput)
	}

	s.GnoSQL.CreateDB(req.DatabaseName, collectionsInput)

	return response, nil

}

func (s *GnoSQLServer) DeleteDatabase(ctx context.Context, req *pb.DatabaseDeleteRequest) (*pb.DataStringResponse, error) {

	var response = &pb.DataStringResponse{
		Data: "Database deleted successfully",
	}

	db := s.GnoSQL.GetDB(req.DatabaseName)

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

	var db *in_memory_database.Database = s.GnoSQL.GetDB(req.DatabaseName)

	if db == nil {
		response.Data = ""
		response.Error = "Database not found"
		return response, nil
	}

	var collectionsInput []in_memory_database.CollectionInput

	for _, EachInput := range req.Collections {
		collectionInput := in_memory_database.CollectionInput{
			CollectionName: EachInput.CollectionName,
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

	var db *in_memory_database.Database = s.GnoSQL.GetDB(req.DatabaseName)

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

	var db *in_memory_database.Database = s.GnoSQL.GetDB(req.DatabaseName)

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

	db, collection := s.GnoSQL.GetDatabaseAndCollection(req.DatabaseName, req.CollectionName)

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

func (s *GnoSQLServer) CreateDocument(ctx context.Context, req *pb.DocumentCreateRequest) (*pb.DocumentCreateResponse, error) {
	response := &pb.DocumentCreateResponse{}

	db, collection := s.GnoSQL.GetDatabaseAndCollection(req.DatabaseName, req.CollectionName)

	if db == nil {
		response.Error = "Database not found"
		return response, nil
	}

	if collection == nil {
		response.Error = "Collection not found"
		return response, nil
	}

	var newDocument in_memory_database.Document

	// Convert JSON to Go struct
	UnMarsalErr := json.Unmarshal([]byte(req.Document), &newDocument)

	if UnMarsalErr != nil {
		response.Error = "Error while marshal document"
		return response, nil
	}

	uniqueUuid := utils.Generate16DigitUUID()

	newDocument["id"] = uniqueUuid

	var createEvent in_memory_database.Event = in_memory_database.Event{
		Type:      utils.EVENT_CREATE,
		EventData: newDocument,
	}

	collection.EventChannel <- createEvent

	responseDataString, MarshalErr := json.Marshal(newDocument)

	if MarshalErr != nil {
		response.Error = "Error while marshal document"
		return response, nil
	}

	response.Data = string(responseDataString)

	return response, nil
}

func (s *GnoSQLServer) ReadDocument(ctx context.Context, req *pb.DocumentReadRequest) (*pb.DocumentReadResponse, error) {
	response := &pb.DocumentReadResponse{}

	db, collection := s.GnoSQL.GetDatabaseAndCollection(req.DatabaseName, req.CollectionName)

	if db == nil {
		response.Error = "Database not found"
		return response, nil
	}

	if collection == nil {
		response.Error = "Collection not found"
		return response, nil
	}

	result := collection.Read(req.Id)

	responseDataString, MarshalErr := json.Marshal(result)

	if MarshalErr != nil {
		response.Error = "Error while marshal document"
		return response, nil
	}

	response.Data = string(responseDataString)

	return response, nil
}

func (s *GnoSQLServer) FilterDocument(ctx context.Context, req *pb.DocumentFilterRequest) (*pb.DocumentFilterResponse, error) {
	response := &pb.DocumentFilterResponse{}

	db, collection := s.GnoSQL.GetDatabaseAndCollection(req.DatabaseName, req.CollectionName)

	if db == nil {
		response.Error = "Database not found"
		return response, nil
	}

	if collection == nil {
		response.Error = "Collection not found"
		return response, nil
	}

	var filterQuery in_memory_database.MapInterface

	// Convert JSON to Go struct
	UnMarsalErr := json.Unmarshal([]byte(req.Filter), &filterQuery)

	if UnMarsalErr != nil {
		response.Error = "Error while marshal document"
		return response, nil
	}

	result := collection.Filter(filterQuery)

	responseDataString, MarshalErr := json.Marshal(result)

	if MarshalErr != nil {
		response.Error = "Error while marshal document"
		return response, nil
	}

	response.Data = string(responseDataString)

	return response, nil
}

func (s *GnoSQLServer) UpdateDocument(ctx context.Context, req *pb.DocumentUpdateRequest) (*pb.DocumentUpdateResponse, error) {
	response := &pb.DocumentUpdateResponse{}

	db, collection := s.GnoSQL.GetDatabaseAndCollection(req.DatabaseName, req.CollectionName)

	if db == nil {
		response.Error = "Database not found"
		return response, nil
	}

	if collection == nil {
		response.Error = "Collection not found"
		return response, nil
	}

	existingDocument := collection.Read(req.Id)

	if existingDocument == nil {
		response.Error = "document not found"
		return response, nil
	}

	var document in_memory_database.Document

	// Convert JSON to Go struct
	UnMarsalErr := json.Unmarshal([]byte(req.Document), &document)

	if UnMarsalErr != nil {
		response.Error = "Error while marshal document"
		return response, nil
	}

	for key, value := range document {
		existingDocument[key] = value
	}

	var updateEvent in_memory_database.Event = in_memory_database.Event{
		Type:      utils.EVENT_UPDATE,
		Id:        req.Id,
		EventData: existingDocument,
	}

	collection.EventChannel <- updateEvent

	responseDataString, MarshalErr := json.Marshal(existingDocument)

	if MarshalErr != nil {
		response.Error = "Error while marshal document"
		return response, nil
	}

	response.Data = string(responseDataString)

	return response, nil
}

func (s *GnoSQLServer) DeleteDocument(ctx context.Context, req *pb.DocumentDeleteRequest) (*pb.DocumentDeleteResponse, error) {
	response := &pb.DocumentDeleteResponse{}

	db, collection := s.GnoSQL.GetDatabaseAndCollection(req.DatabaseName, req.CollectionName)

	if db == nil {
		response.Error = "Database not found"
		return response, nil
	}

	if collection == nil {
		response.Error = "Collection not found"
		return response, nil
	}

	existingDocument := collection.Read(req.Id)

	if existingDocument == nil {
		response.Error = "document not found"
		return response, nil
	}

	var deleteEvent in_memory_database.Event = in_memory_database.Event{
		Type: utils.EVENT_DELETE,
		Id:   req.Id,
	}

	collection.EventChannel <- deleteEvent

	response.Data = utils.DOCUMENT_DELETE_SUCCESS_MSG

	return response, nil
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
