package router

type DatabaseRequestInput struct {
	// Example: databaseName
	DatabaseName string

	// Example: collections
	Collections []map[string]interface{}
}
