package service

import (
	"errors"
	"gnosql/src/common"
	"gnosql/src/global_constants"
	"gnosql/src/in_memory_database"
)

func ConnectDatabase(gnoSQL *in_memory_database.GnoSQL, DatabaseName string, collectionsInput []in_memory_database.CollectionInput) in_memory_database.DatabaseConnectResult {
	var result = in_memory_database.DatabaseConnectResult{}

	db := gnoSQL.GetDB(DatabaseName)

	if db == nil {
		db = gnoSQL.CreateDB(DatabaseName, collectionsInput)
	} else {
		db.CreateColls(collectionsInput)
	}

	result.Data = in_memory_database.DatabaseResult{
		DatabaseName: db.DatabaseName,
		Collections:  db.GetCollectionNames(),
	}

	return result
}

func CreateDatabase(gnoSQL *in_memory_database.GnoSQL, DatabaseName string, collectionsInput []in_memory_database.CollectionInput) (in_memory_database.DatabaseCreateResult, error) {
	var result = in_memory_database.DatabaseCreateResult{}

	db := gnoSQL.GetDB(DatabaseName)

	if db != nil {
		return result, errors.New("Database already exists")
	}

	gnoSQL.CreateDB(DatabaseName, collectionsInput)

	result.Data = global_constants.DATABASE_CREATE_SUCCESS_MSG

	return result, nil
}

func DeleteDatabase(gnoSQL *in_memory_database.GnoSQL, DatabaseName string) (in_memory_database.DatabaseDeleteResult, error) {
	var result = in_memory_database.DatabaseDeleteResult{}

	db := gnoSQL.GetDB(DatabaseName)

	if err := validateDatabase(db); err != nil {
		return result, err
	}

	gnoSQL.DeleteDB(db)

	result.Data = global_constants.DATABASE_DELETE_SUCCESS_MSG

	return result, nil
}

func GetAllDatabase(gnoSQL *in_memory_database.GnoSQL) (in_memory_database.DatabaseGetAllResult, error) {
	var result = in_memory_database.DatabaseGetAllResult{}

	databaseNames := make([]string, 0)

	for _, database := range gnoSQL.Databases {
		databaseNames = append(databaseNames, database.DatabaseName)
	}

	result.Data = databaseNames

	return result, nil
}

func LoadToDisk(gnoSQL *in_memory_database.GnoSQL) (in_memory_database.DatabaseLoadToDiskResult, error) {
	var result = in_memory_database.DatabaseLoadToDiskResult{}

	go gnoSQL.WriteAllDBs()

	result.Data = global_constants.DATABASE_LOAD_TO_DISK_MSG
	return result, nil
}

func CreateCollections(gnoSQL *in_memory_database.GnoSQL, DatabaseName string, collectionsInput []in_memory_database.CollectionInput) (in_memory_database.CollectionCreateResult, error) {
	var result = in_memory_database.CollectionCreateResult{}

	db := gnoSQL.GetDB(DatabaseName)

	if err := validateDatabase(db); err != nil {
		return result, err
	}

	db.CreateColls(collectionsInput)

	result.Data = global_constants.COLLECTION_CREATE_SUCCESS_MSG

	return result, nil
}

func DeleteCollections(gnoSQL *in_memory_database.GnoSQL, DatabaseName string, collections []string) (in_memory_database.CollectionDeleteResult, error) {
	var result = in_memory_database.CollectionDeleteResult{}

	db := gnoSQL.GetDB(DatabaseName)

	if err := validateDatabase(db); err != nil {
		return result, err
	}

	db.DeleteColls(collections)

	result.Data = global_constants.COLLECTION_DELETE_SUCCESS_MSG

	return result, nil
}

func GetAllCollections(gnoSQL *in_memory_database.GnoSQL, DatabaseName string) (in_memory_database.CollectionGetAllResult, error) {
	var result = in_memory_database.CollectionGetAllResult{}

	db := gnoSQL.GetDB(DatabaseName)

	if err := validateDatabase(db); err != nil {
		return result, err
	}

	allCollections := db.Collections

	collections := make([]string, 0)

	for _, collection := range allCollections {
		collections = append(collections, collection.CollectionName)
	}

	result.Data = collections

	return result, nil
}

func GetCollectionStats(gnoSQL *in_memory_database.GnoSQL, DatabaseName string, CollectionName string) (in_memory_database.CollectionStatsResult, error) {
	var result = in_memory_database.CollectionStatsResult{}

	db, collection := gnoSQL.GetDatabaseAndCollection(DatabaseName, CollectionName)

	if err := validateDatabaseAndCollection(db, collection); err != nil {
		return result, err
	}

	stats := collection.Stats()
	result.Data = stats

	return result, nil
}

func DocumentCreate(gnoSQL *in_memory_database.GnoSQL,
	DatabaseName string, CollectionName string, document in_memory_database.Document) (in_memory_database.DocumentCreateResult, error) {

	var result = in_memory_database.DocumentCreateResult{}

	db, collection := gnoSQL.GetDatabaseAndCollection(DatabaseName, CollectionName)

	if err := validateDatabaseAndCollection(db, collection); err != nil {
		return result, err
	}

	if document["docId"] == nil {
		document["docId"] = common.Generate16DigitUUID()
	}

	var createEvent in_memory_database.Event = GenerateCreateEvent(document)

	go in_memory_database.AddIncomingRequest(db.DatabaseName, collection.CollectionName, createEvent)

	result.Data = document

	return result, nil
}

func DocumentRead(gnoSQL *in_memory_database.GnoSQL,
	DatabaseName string, CollectionName string, id string) (in_memory_database.DocumentReadResult, error) {

	var result = in_memory_database.DocumentReadResult{}

	db, collection := gnoSQL.GetDatabaseAndCollection(DatabaseName, CollectionName)

	if err := validateDatabaseAndCollection(db, collection); err != nil {
		return result, err
	}

	existingDocument := collection.Read(id)

	if existingDocument == nil {
		return result, errors.New(global_constants.DOCUMENT_NOT_FOUND_MSG)
	}

	result.Data = existingDocument

	return result, nil
}

func DocumentFilter(gnoSQL *in_memory_database.GnoSQL,
	DatabaseName string, CollectionName string, filter in_memory_database.MapInterface) (in_memory_database.DocumentFilterResult, error) {

	var result = in_memory_database.DocumentFilterResult{}

	db, collection := gnoSQL.GetDatabaseAndCollection(DatabaseName, CollectionName)

	if err := validateDatabaseAndCollection(db, collection); err != nil {
		return result, err
	}

	documents := collection.Filter(filter)

	result.Data = documents

	return result, nil
}

func DocumentUpdate(gnoSQL *in_memory_database.GnoSQL,
	DatabaseName string, CollectionName string, id string,
	document in_memory_database.Document) (in_memory_database.DocumentUpdateResult, error) {

	var result = in_memory_database.DocumentUpdateResult{}

	db, collection := gnoSQL.GetDatabaseAndCollection(DatabaseName, CollectionName)

	if err := validateDatabaseAndCollection(db, collection); err != nil {
		return result, err
	}

	existingDocument := collection.Read(id)
	if existingDocument == nil {
		return result, errors.New(global_constants.DOCUMENT_NOT_FOUND_MSG)
	}

	for key, value := range document {
		existingDocument[key] = value
	}

	var updateEvent in_memory_database.Event = GenerateUpdateEvent(existingDocument)

	go in_memory_database.AddIncomingRequest(db.DatabaseName, collection.CollectionName, updateEvent)

	result.Data = existingDocument

	return result, nil
}

func DocumentDelete(gnoSQL *in_memory_database.GnoSQL,
	DatabaseName string, CollectionName string, id string) (in_memory_database.DocumentDeleteResult, error) {

	var result = in_memory_database.DocumentDeleteResult{}

	db, collection := gnoSQL.GetDatabaseAndCollection(DatabaseName, CollectionName)

	if err := validateDatabaseAndCollection(db, collection); err != nil {
		return result, err
	}

	existingDocument := collection.Read(id)

	if existingDocument == nil {
		return result, errors.New(global_constants.DOCUMENT_NOT_FOUND_MSG)
	}

	var deleteEvent in_memory_database.Event = GenerateDeleteEvent(id)

	go in_memory_database.AddIncomingRequest(db.DatabaseName, collection.CollectionName, deleteEvent)

	result.Data = global_constants.DOCUMENT_DELETE_SUCCESS_MSG

	return result, nil
}

func DocumentGetAll(gnoSQL *in_memory_database.GnoSQL,
	DatabaseName string, CollectionName string) (in_memory_database.DocumentGetAllResult, error) {

	var result = in_memory_database.DocumentGetAllResult{}

	db, collection := gnoSQL.GetDatabaseAndCollection(DatabaseName, CollectionName)

	if err := validateDatabaseAndCollection(db, collection); err != nil {
		return result, err
	}

	documents := collection.GetAllData()

	result.Data = documents

	return result, nil
}

// validateDatabase checks if db is nil, returns an error if it is
func validateDatabase(db *in_memory_database.Database) error {
	if db == nil {
		return errors.New(global_constants.DATABASE_NOT_FOUND_MSG)
	}
	return nil
}

// validateCollection checks if collection is nil, returns an error if it is
func validateCollection(collection *in_memory_database.Collection) error {
	if collection == nil {
		return errors.New(global_constants.COLLECTION_NOT_FOUND_MSG)
	}
	return nil
}

// validateDatabaseAndCollection checks if db or collection are nil, returns an error if either is nil
func validateDatabaseAndCollection(db *in_memory_database.Database, collection *in_memory_database.Collection) error {
	if err := validateDatabase(db); err != nil {
		return err
	}
	return validateCollection(collection)
}

func GenerateCreateEvent(document in_memory_database.Document) in_memory_database.Event {
	var EventDocument = make(in_memory_database.Document)

	for key, value := range document {
		EventDocument[key] = value
	}

	return in_memory_database.Event{
		Type:      global_constants.EVENT_CREATE,
		EventData: EventDocument,
	}
}

func GenerateUpdateEvent(document in_memory_database.Document) in_memory_database.Event {
	var EventDocument = make(in_memory_database.Document)
	id := document["docId"]

	for key, value := range document {
		EventDocument[key] = value
	}

	return in_memory_database.Event{
		Id:        id.(string),
		Type:      global_constants.EVENT_UPDATE,
		EventData: EventDocument,
	}
}

func GenerateDeleteEvent(id string) in_memory_database.Event {
	return in_memory_database.Event{
		Id:   id,
		Type: global_constants.EVENT_DELETE,
	}
}
