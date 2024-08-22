package grpc_handler

import (
	"context"
	"encoding/json"
	"errors"
	pb "gnosql/proto"
	"gnosql/src/global_constants"
	"gnosql/src/in_memory_database"
	"gnosql/src/service"
)

type GnoSQLServer struct {
	pb.UnimplementedGnoSQLServiceServer

	GnoSQL *in_memory_database.GnoSQL
}

func (s *GnoSQLServer) CreateNewDatabase(ctx context.Context,
	req *pb.DatabaseCreateRequest) (*pb.DatabaseCreateResponse, error) {

	response := &pb.DatabaseCreateResponse{}
	var collectionsInput = ConvertReqToCollectionInput(req.GetCollections())

	result, err := service.CreateDatabase(s.GnoSQL, req.DatabaseName, collectionsInput)

	response.Data = result.Data
	return response, err

}

func (s *GnoSQLServer) ConnectDatabase(ctx context.Context,
	req *pb.DatabaseCreateRequest) (*pb.DatabaseConnectResponse, error) {

	response := &pb.DatabaseConnectResponse{}
	var collectionsInput = ConvertReqToCollectionInput(req.GetCollections())

	result := service.ConnectDatabase(s.GnoSQL, req.DatabaseName, collectionsInput)

	response.Data = &pb.DatabaseResponse{
		DatabaseName: result.Data.DatabaseName,
		Collections:  result.Data.Collections,
	}

	return response, nil

}

func (s *GnoSQLServer) DeleteDatabase(ctx context.Context, req *pb.DatabaseDeleteRequest) (*pb.DatabaseDeleteResponse, error) {
	var response = &pb.DatabaseDeleteResponse{}

	result, err := service.DeleteDatabase(s.GnoSQL, req.DatabaseName)
	response.Data = result.Data

	return response, err
}

func (s *GnoSQLServer) GetAllDatabases(ctx context.Context, req *pb.NoRequestBody) (*pb.DatabaseGetAllResponse, error) {
	var response = &pb.DatabaseGetAllResponse{}

	result, err := service.GetAllDatabase(s.GnoSQL)
	response.Data = result.Data

	return response, err
}

func (s *GnoSQLServer) LoadToDisk(ctx context.Context, req *pb.NoRequestBody) (*pb.LoadToDiskResponse, error) {
	var response = &pb.LoadToDiskResponse{}

	result, err := service.LoadToDisk(s.GnoSQL)
	response.Data = result.Data

	return response, err
}

func (s *GnoSQLServer) CreateNewCollection(ctx context.Context, req *pb.CollectionCreateRequest) (*pb.CollectionCreateResponse, error) {
	response := &pb.CollectionCreateResponse{}
	var collectionsInput = ConvertReqToCollectionInput(req.GetCollections())

	result, err := service.CreateCollections(s.GnoSQL, req.DatabaseName, collectionsInput)
	response.Data = result.Data

	return response, err
}

func (s *GnoSQLServer) DeleteCollections(ctx context.Context, req *pb.CollectionDeleteRequest) (*pb.CollectionDeleteResponse, error) {
	response := &pb.CollectionDeleteResponse{}

	result, err := service.DeleteCollections(s.GnoSQL, req.DatabaseName, req.GetCollections())
	response.Data = result.Data

	return response, err

}

func (s *GnoSQLServer) GetAllCollections(ctx context.Context, req *pb.CollectionGetAllRequest) (*pb.CollectionGetAllResponse, error) {

	response := &pb.CollectionGetAllResponse{}

	result, err := service.GetAllCollections(s.GnoSQL, req.DatabaseName)
	response.Data = result.Data

	return response, err
}

func (s *GnoSQLServer) GetCollectionStats(ctx context.Context, req *pb.CollectionStatsRequest) (*pb.CollectionStatsResponse, error) {

	response := &pb.CollectionStatsResponse{}

	result, err := service.GetCollectionStats(s.GnoSQL, req.DatabaseName, req.CollectionName)

	response.Data = &pb.CollectionStats{
		CollectionName: result.Data.CollectionName,
		IndexKeys:      result.Data.IndexKeys,
		Documents:      int32(result.Data.Documents),
	}

	return response, err
}

func (s *GnoSQLServer) CreateDocument(ctx context.Context, req *pb.DocumentCreateRequest) (*pb.DocumentCreateResponse, error) {
	response := &pb.DocumentCreateResponse{}

	var newDocument in_memory_database.Document

	// Convert JSON to Go struct
	UnMarsalErr := json.Unmarshal([]byte(req.Document), &newDocument)

	if UnMarsalErr != nil {
		return response, errors.New(global_constants.ERROR_WHILE_UNMARSHAL_JSON)
	}

	result, err := service.DocumentCreate(s.GnoSQL, req.DatabaseName, req.CollectionName, newDocument)

	if err != nil {
		return response, err
	}

	resultString, err := ConvertDocumentMapToString(result.Data)

	response.Data = resultString

	return response, err
}

func (s *GnoSQLServer) ReadDocument(ctx context.Context, req *pb.DocumentReadRequest) (*pb.DocumentReadResponse, error) {
	response := &pb.DocumentReadResponse{}

	result, err := service.DocumentRead(s.GnoSQL, req.DatabaseName, req.CollectionName, req.DocId)

	resultString, err := ConvertDocumentMapToString(result.Data)

	if err != nil {
		return response, err
	}

	response.Data = resultString

	return response, err
}

func (s *GnoSQLServer) FilterDocument(ctx context.Context, req *pb.DocumentFilterRequest) (*pb.DocumentFilterResponse, error) {
	response := &pb.DocumentFilterResponse{}

	var filter in_memory_database.MapInterface

	UnMarsalErr := json.Unmarshal([]byte(req.Filter), &filter)

	if UnMarsalErr != nil {
		return response, errors.New(global_constants.ERROR_WHILE_UNMARSHAL_JSON)
	}

	result, err := service.DocumentFilter(s.GnoSQL, req.DatabaseName, req.CollectionName, filter)

	if err != nil {
		return response, err
	}

	resultString, err := ConvertDocumentMapsToString(result.Data)

	response.Data = resultString

	return response, err
}

func (s *GnoSQLServer) UpdateDocument(ctx context.Context, req *pb.DocumentUpdateRequest) (*pb.DocumentUpdateResponse, error) {
	response := &pb.DocumentUpdateResponse{}

	var document in_memory_database.Document

	UnMarsalErr := json.Unmarshal([]byte(req.Document), &document)

	if UnMarsalErr != nil {
		return response, errors.New(global_constants.ERROR_WHILE_UNMARSHAL_JSON)
	}

	result, err := service.DocumentUpdate(s.GnoSQL, req.DatabaseName, req.CollectionName, req.DocId, document)
	if err != nil {
		return response, err
	}

	resultString, err := ConvertDocumentMapToString(result.Data)

	response.Data = resultString

	return response, err
}

func (s *GnoSQLServer) DeleteDocument(ctx context.Context, req *pb.DocumentDeleteRequest) (*pb.DocumentDeleteResponse, error) {
	response := &pb.DocumentDeleteResponse{}

	result, err := service.DocumentDelete(s.GnoSQL, req.DatabaseName, req.CollectionName, req.DocId)
	if err != nil {
		return response, err
	}

	response.Data = result.Data
	return response, nil
}

func (s *GnoSQLServer) GetAllDocuments(ctx context.Context, req *pb.DocumentGetAllRequest) (*pb.DocumentGetAllResponse, error) {
	response := &pb.DocumentGetAllResponse{}

	result, err := service.DocumentGetAll(s.GnoSQL, req.DatabaseName, req.CollectionName)

	if err != nil {
		return response, err
	}

	resultString, err := ConvertDocumentMapsToString(result.Data)

	response.Data = resultString

	return response, err
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

func ConvertDocumentMapToString(document in_memory_database.Document) (string, error) {

	responseDataString, MarshalErr := json.Marshal(document)

	if MarshalErr != nil {
		return "", errors.New(global_constants.ERROR_WHILE_MARSHAL_JSON)
	}

	return string(responseDataString), nil

}

func ConvertDocumentMapsToString(document []in_memory_database.Document) (string, error) {
	responseDataString, MarshalErr := json.Marshal(document)

	if MarshalErr != nil {
		return "", errors.New(global_constants.ERROR_WHILE_MARSHAL_JSON)
	}
	return string(responseDataString), nil
}
