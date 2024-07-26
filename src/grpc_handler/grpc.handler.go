package grpc_handler

import (
	"context"
	"encoding/json"
	"fmt"
	pb "gnosql/proto"
	"gnosql/src/in_memory_database"
	"gnosql/src/service"
	"gnosql/src/utils"
)

type GnoSQLServer struct {
	pb.UnimplementedGnoSQLServiceServer

	GnoSQL *in_memory_database.GnoSQL
}

func (s *GnoSQLServer) CreateNewDatabase(ctx context.Context,
	req *pb.DatabaseCreateRequest) (*pb.DatabaseCreateResponse, error) {

	response := &pb.DatabaseCreateResponse{}
	var collectionsInput = ConvertReqToCollectionInput(req.GetCollections())

	result := service.CreateDatabase(s.GnoSQL, req.DatabaseName, collectionsInput)

	response.Data = result.Data
	response.Error = result.Error
	return response, nil

}

func (s *GnoSQLServer) DeleteDatabase(ctx context.Context, req *pb.DatabaseDeleteRequest) (*pb.DatabaseDeleteResponse, error) {

	var response = &pb.DatabaseDeleteResponse{}

	result := service.DeleteDatabase(s.GnoSQL, req.DatabaseName)

	response.Data = result.Data
	response.Error = result.Error
	return response, nil
}

func (s *GnoSQLServer) GetAllDatabases(ctx context.Context, req *pb.NoRequestBody) (*pb.DatabaseGetAllResponse, error) {
	var response = &pb.DatabaseGetAllResponse{}

	result := service.GetAllDatabase(s.GnoSQL)

	response.Data = result.Data
	response.Error = result.Error

	return response, nil
}

func (s *GnoSQLServer) LoadToDisk(ctx context.Context, req *pb.NoRequestBody) (*pb.LoadToDiskResponse, error) {
	var response = &pb.LoadToDiskResponse{}

	result := service.LoadToDisk(s.GnoSQL)

	response.Data = result.Data
	response.Error = result.Error

	return response, nil
}

func (s *GnoSQLServer) CreateNewCollection(ctx context.Context, req *pb.CollectionCreateRequest) (*pb.CollectionCreateResponse, error) {
	response := &pb.CollectionCreateResponse{}

	var collectionsInput = ConvertReqToCollectionInput(req.GetCollections())

	result := service.CreateCollections(s.GnoSQL, req.DatabaseName, collectionsInput)

	response.Data = result.Data
	response.Error = result.Error

	return response, nil
}

func (s *GnoSQLServer) DeleteCollections(ctx context.Context, req *pb.CollectionDeleteRequest) (*pb.CollectionDeleteResponse, error) {
	response := &pb.CollectionDeleteResponse{}

	result := service.DeleteCollections(s.GnoSQL, req.DatabaseName, req.GetCollections())

	response.Data = result.Data
	response.Error = result.Error

	return response, nil

}

func (s *GnoSQLServer) GetAllCollections(ctx context.Context, req *pb.CollectionGetAllRequest) (*pb.CollectionGetAllResponse, error) {

	response := &pb.CollectionGetAllResponse{}

	result := service.GetAllCollections(s.GnoSQL, req.DatabaseName)

	response.Data = result.Data
	response.Error = result.Error

	return response, nil
}

func (s *GnoSQLServer) GetCollectionStats(ctx context.Context, req *pb.CollectionStatsRequest) (*pb.CollectionStatsResponse, error) {

	response := &pb.CollectionStatsResponse{}

	result := service.GetCollectionStats(s.GnoSQL, req.DatabaseName, req.CollectionName)

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

	var newDocument in_memory_database.Document

	// Convert JSON to Go struct
	UnMarsalErr := json.Unmarshal([]byte(req.Document), &newDocument)

	if UnMarsalErr != nil {
		response.Error = utils.ERROR_WHILE_UNMARSHAL_JSON
		return response, nil
	}

	result := service.DocumentCreate(s.GnoSQL, req.DatabaseName, req.CollectionName, newDocument)

	Data, Error := ConvertDocumentMapToString(result.Data, result.Error)

	response.Data = Data
	response.Error = Error

	return response, nil
}

func (s *GnoSQLServer) ReadDocument(ctx context.Context, req *pb.DocumentReadRequest) (*pb.DocumentReadResponse, error) {
	response := &pb.DocumentReadResponse{}

	result := service.DocumentRead(s.GnoSQL, req.DatabaseName, req.CollectionName, req.Id)

	Data, Error := ConvertDocumentMapToString(result.Data, result.Error)

	response.Data = Data
	response.Error = Error

	return response, nil
}

func (s *GnoSQLServer) FilterDocument(ctx context.Context, req *pb.DocumentFilterRequest) (*pb.DocumentFilterResponse, error) {
	response := &pb.DocumentFilterResponse{}

	var filter in_memory_database.MapInterface

	UnMarsalErr := json.Unmarshal([]byte(req.Filter), &filter)

	if UnMarsalErr != nil {
		response.Error = utils.ERROR_WHILE_UNMARSHAL_JSON
		return response, nil
	}

	result := service.DocumentFilter(s.GnoSQL, req.DatabaseName, req.CollectionName, filter)

	Data, Error := ConvertDocumentMapsToString(result.Data, result.Error)

	response.Data = Data
	response.Error = Error

	return response, nil
}

func (s *GnoSQLServer) UpdateDocument(ctx context.Context, req *pb.DocumentUpdateRequest) (*pb.DocumentUpdateResponse, error) {
	response := &pb.DocumentUpdateResponse{}

	var document in_memory_database.Document

	UnMarsalErr := json.Unmarshal([]byte(req.Document), &document)

	if UnMarsalErr != nil {
		response.Error = utils.ERROR_WHILE_UNMARSHAL_JSON
		return response, nil
	}

	result := service.DocumentUpdate(s.GnoSQL, req.DatabaseName, req.CollectionName, req.Id, document)

	Data, Error := ConvertDocumentMapToString(result.Data, result.Error)

	response.Data = Data
	response.Error = Error

	return response, nil
}

func (s *GnoSQLServer) DeleteDocument(ctx context.Context, req *pb.DocumentDeleteRequest) (*pb.DocumentDeleteResponse, error) {
	response := &pb.DocumentDeleteResponse{}

	result := service.DocumentDelete(s.GnoSQL, req.DatabaseName, req.CollectionName, req.Id)

	response.Data = result.Data
	response.Error = result.Error
	return response, nil
}

func (s *GnoSQLServer) GetAllDocuments(ctx context.Context, req *pb.DocumentGetAllRequest) (*pb.DocumentGetAllResponse, error) {
	response := &pb.DocumentGetAllResponse{}

	result := service.DocumentGetAll(s.GnoSQL, req.DatabaseName, req.CollectionName)

	Data, Error := ConvertDocumentMapsToString(result.Data, result.Error)

	response.Data = Data
	response.Error = Error

	return response, nil
}
func ConvertReqToCollectionInput(collections []*pb.CollectionInput) []in_memory_database.CollectionInput {
	fmt.Printf("ConvertReqToCollectionInput")

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

func ConvertDocumentMapToString(document in_memory_database.Document, gRPCError string) (string, string) {
	responseDataString, MarshalErr := json.Marshal(document)

	if MarshalErr != nil {
		return "", utils.ERROR_WHILE_MARSHAL_JSON
	}
	return string(responseDataString), gRPCError
}

func ConvertDocumentMapsToString(document []in_memory_database.Document, gRPCError string) (string, string) {
	responseDataString, MarshalErr := json.Marshal(document)

	if MarshalErr != nil {
		return "", utils.ERROR_WHILE_MARSHAL_JSON
	}
	return string(responseDataString), gRPCError
}
