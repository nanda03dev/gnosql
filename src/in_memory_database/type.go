package in_memory_database

type MapString map[string]string // Ex: { "name": "name" }
type MapStrings map[string]string
type MapInterface map[string]interface{}

type Result struct {
	Data  interface{} `json:"Data"`
	Error string      `json:"Error"`
}

type DatabaseCreateRequest struct {
	DatabaseName string
	Collections  []CollectionInput
}

type DatabaseCreateResult struct {
	Data  string
	Error string
}

type DatabaseDeleteRequest struct {
	DatabaseName string
}
type DatabaseDeleteResult struct {
	Data  string
	Error string
}

type DatabaseGetAllResult struct {
	Data  []string
	Error string
}

type DatabaseLoadToDiskResult struct {
	Data  string
	Error string
}

type CollectionCreateRequest struct {
	DatabaseName string
	Collections  []CollectionInput
}

type CollectionCreateResult struct {
	Data  string
	Error string
}

type CollectionDeleteRequest struct {
	DatabaseName string
	Collections  []string
}

type CollectionDeleteResult struct {
	Data  string
	Error string
}

type CollectionGetAllRequest struct {
	DatabaseName string
}

type CollectionGetAllResult struct {
	Data  []string
	Error string
}

type CollectionStatsRequest struct {
	DatabaseName   string
	CollectionName string
}

type CollectionStatsResult struct {
	Data  CollectionStats
	Error string
}

type DocumentCreateRequest struct {
	DatabaseName   string
	CollectionName string
	Document       Document
}

type DocumentCreateResult struct {
	Data  Document
	Error string
}

type DocumentReadRequest struct {
	DatabaseName   string
	CollectionName string
	Id             string
}

type DocumentReadResult struct {
	Data  Document
	Error string
}

type DocumentFilterRequest struct {
	DatabaseName   string
	CollectionName string
	Filter         MapInterface
}

type DocumentFilterResult struct {
	Data  []Document
	Error string
}

type DocumentUpdateRequest struct {
	DatabaseName   string
	CollectionName string
	Id             string
	Document       Document
}

type DocumentUpdateResult struct {
	Data  Document
	Error string
}

type DocumentDeleteRequest struct {
	DatabaseName   string
	CollectionName string
	Id             string
}

type DocumentDeleteResult struct {
	Data  string
	Error string
}

type DocumentGetAllRequest struct {
	DatabaseName   string
	CollectionName string
}

type DocumentGetAllResult struct {
	Data  []Document
	Error string
}
