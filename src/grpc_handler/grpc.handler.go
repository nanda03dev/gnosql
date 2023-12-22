package grpc_handler

import (
	"context"
	"encoding/json"
	"github.com/gin-gonic/gin"
	pb "gnosql/proto"
	"gnosql/src/in_memory_database"
	"gnosql/src/service"
	"gnosql/src/utils"
	"net/http"
)

type GnoSQLServer struct {
	pb.UnimplementedGnoSQLServiceServer

	GnoSQL *in_memory_database.GnoSQL
}

func (s *GnoSQLServer) CreateNewDatabase(ctx context.Context, req *pb.DatabaseCreateRequest) (*pb.DataStringResponse, error) {
	response := &pb.DataStringResponse{}

	var collectionsInput = ConvertReqToCollectionInput(req.GetCollections())

	result := service.ServiceCreateDatabase(s.GnoSQL, req.DatabaseName, collectionsInput)

	response.Data = result.Data
	response.Error = result.Error
	return response, nil

}

func (s *GnoSQLServer) DeleteDatabase(ctx context.Context, req *pb.DatabaseDeleteRequest) (*pb.DataStringResponse, error) {

	var response = &pb.DataStringResponse{}

	result := service.ServiceDeleteDatabase(s.GnoSQL, req.DatabaseName)

	response.Data = result.Data
	response.Error = result.Error
	return response, nil
}

func (s *GnoSQLServer) GetAllDatabases(ctx context.Context, req *pb.NoRequestBody) (*pb.DatabaseGetAllResult, error) {
	var response = &pb.DatabaseGetAllResult{}

	result := service.ServiceGetAllDatabase(s.GnoSQL)

	response.Data = result.Data
	response.Error = result.Error

	return response, nil
}

func (s *GnoSQLServer) LoadToDisk(ctx context.Context, req *pb.NoRequestBody) (*pb.DataStringResponse, error) {
	var response = &pb.DataStringResponse{}

	result := service.ServiceLoadToDisk(s.GnoSQL)

	response.Data = result.Data
	response.Error = result.Error

	return response, nil
}

func (s *GnoSQLServer) CreateNewCollection(ctx context.Context, req *pb.CollectionCreateRequest) (*pb.DataStringResponse, error) {
	response := &pb.DataStringResponse{}

	var collectionsInput = ConvertReqToCollectionInput(req.GetCollections())

	result := service.ServiceCreateCollections(s.GnoSQL, req.DatabaseName, collectionsInput)

	response.Data = result.Data
	response.Error = result.Error

	return response, nil
}

func (s *GnoSQLServer) DeleteCollections(ctx context.Context, req *pb.CollectionDeleteRequest) (*pb.DataStringResponse, error) {
	response := &pb.DataStringResponse{}

	result := service.ServiceDeleteCollections(s.GnoSQL, req.DatabaseName, req.GetCollections())

	response.Data = result.Data
	response.Error = result.Error

	return response, nil

}

func (s *GnoSQLServer) GetAllCollections(ctx context.Context, req *pb.CollectionGetAllRequest) (*pb.CollectionGetAllResult, error) {

	response := &pb.CollectionGetAllResult{}

	result := service.ServiceGetAllCollections(s.GnoSQL, req.DatabaseName)

	response.Data = result.Data
	response.Error = result.Error

	return response, nil
}

func (s *GnoSQLServer) GetCollectionStats(ctx context.Context, req *pb.CollectionStatsRequest) (*pb.CollectionStatsResponse, error) {

	response := &pb.CollectionStatsResponse{}

	result := service.ServiceGetCollectionStats(s.GnoSQL, req.DatabaseName, req.CollectionName)

	response.Data = &pb.CollectionStats{
		CollectionName: result.Data.CollectionName,
		IndexKeys:      result.Data.IndexKeys,
		Documents:      int32(result.Data.Documents),
	}

	response.Error = result.Error

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

func ConvertReqToCollectionInput(collections []*pb.CollectionInput) []in_memory_database.CollectionInput {

	var collectionsInput []in_memory_database.CollectionInput

	for _, EachInput := range collections {
		collectionInput := in_memory_database.CollectionInput{
			CollectionName: EachInput.CollectionName,
			IndexKeys:      EachInput.IndexKeys,
		}
		collectionsInput = append(collectionsInput, collectionInput)
	}

	return collectionsInput
}
