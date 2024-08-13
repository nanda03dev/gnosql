package in_memory_database

type MapString map[string]string // Ex: { "name": "name" }
type MapStrings map[string]string
type MapInterface map[string]interface{}

type Result struct {
	Data  interface{} `json:"data"`
	Error string      `json:"error"`
}

type DatabaseCreateRequest struct {
	DatabaseName string            `json:"databaseName"`
	Collections  []CollectionInput `json:"collections"`
}

type DatabaseCreateResult struct {
	Data  string `json:"data"`
	Error string `json:"error"`
}

type DatabaseResult struct {
	DatabaseName string   `json:"databaseName"`
	Collections  []string `json:"collections"`
}

type DatabaseConnectResult struct {
	Data DatabaseResult `json:"data"`
}

type DatabaseDeleteRequest struct {
	DatabaseName string `json:"databaseName"`
}
type DatabaseDeleteResult struct {
	Data string `json:"data"`
}

type DatabaseGetAllResult struct {
	Data []string `json:"data"`
}

type DatabaseLoadToDiskResult struct {
	Data string `json:"data"`
}

type CollectionCreateRequest struct {
	DatabaseName string            `json:"databaseName"`
	Collections  []CollectionInput `json:"collections"`
}

type CollectionCreateResult struct {
	Data string `json:"data"`
}

type CollectionDeleteRequest struct {
	DatabaseName string   `json:"databaseName"`
	Collections  []string `json:"collections"`
}

type CollectionDeleteResult struct {
	Data string `json:"data"`
}

type CollectionGetAllRequest struct {
	DatabaseName string `json:"databaseName"`
}

type CollectionGetAllResult struct {
	Data []string `json:"data"`
}

type CollectionStatsRequest struct {
	DatabaseName   string `json:"databaseName"`
	CollectionName string `json:"collectionName"`
}

type CollectionStatsResult struct {
	Data CollectionStats
}

type DocumentCreateRequest struct {
	DatabaseName   string   `json:"databaseName"`
	CollectionName string   `json:"collectionName"`
	Document       Document `json:"document"`
}

type DocumentCreateResult struct {
	Data Document `json:"data"`
}

type DocumentReadRequest struct {
	DatabaseName   string `json:"databaseName"`
	CollectionName string `json:"collectionName"`
	DocId          string `json:"docId"`
}

type DocumentReadResult struct {
	Data Document `json:"data"`
}

type DocumentFilterRequest struct {
	DatabaseName   string `json:"databaseName"`
	CollectionName string `json:"collectionName"`
	Filter         MapInterface
}

type DocumentFilterResult struct {
	Data []Document `json:"data"`
}

type DocumentUpdateRequest struct {
	DatabaseName   string   `json:"databaseName"`
	CollectionName string   `json:"collectionName"`
	DocId          string   `json:"docId"`
	Document       Document `json:"document"`
}

type DocumentUpdateResult struct {
	Data Document `json:"data"`
}

type DocumentDeleteRequest struct {
	DatabaseName   string `json:"databaseName"`
	CollectionName string `json:"collectionName"`
	DocId          string `json:"docId"`
}

type DocumentDeleteResult struct {
	Data string `json:"data"`
}

type DocumentGetAllRequest struct {
	DatabaseName   string `json:"databaseName"`
	CollectionName string `json:"collectionName"`
}

type DocumentGetAllResult struct {
	Data []Document `json:"data"`
}
