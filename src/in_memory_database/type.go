package in_memory_database

type MapString map[string]string // Ex: { "name": "name" }
type MapStrings map[string]string
type MapInterface map[string]interface{}

type Result struct {
	Data  interface{} `json:"Data"`
	Error string      `json:"Error"`
}

type DatabaseCreateResult struct {
	Data  string
	Error string
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

type CollectionCreateResult struct {
	Data  string
	Error string
}

type CollectionDeleteResult struct {
	Data  string
	Error string
}

type CollectionGetAllResult struct {
	Data  []string
	Error string
}

type CollectionStatsResult struct {
	Data  CollectionStats
	Error string
}
