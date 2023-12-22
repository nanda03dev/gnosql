package service

import (
	"fmt"
	"gnosql/src/in_memory_database"
	"gnosql/src/utils"
)

func ServiceCreateDatabase(gnoSQL *in_memory_database.GnoSQL, DatabaseName string, collectionsInput []in_memory_database.CollectionInput) in_memory_database.DatabaseCreateResult {
	var result = in_memory_database.DatabaseCreateResult{}

	db := gnoSQL.GetDB(DatabaseName)

	fmt.Printf("\n collectionsInput %v \n ", collectionsInput)

	if db != nil {
		result.Error = "Database already exists"
		return result
	}

	gnoSQL.CreateDB(DatabaseName, collectionsInput)

	result.Data = utils.DATABASE_CREATE_SUCCESS_MSG

	return result
}

func ServiceDeleteDatabase(gnoSQL *in_memory_database.GnoSQL, DatabaseName string) in_memory_database.DatabaseDeleteResult {
	var result = in_memory_database.DatabaseDeleteResult{}

	fmt.Printf("\n DatabaseName %v\n ", DatabaseName)
	db := gnoSQL.GetDB(DatabaseName)

	if db == nil {
		result.Error = utils.DATABASE_NOT_FOUND_MSG

		return result
	}

	gnoSQL.DeleteDB(db)

	result.Data = utils.DATABASE_DELETE_SUCCESS_MSG

	return result
}

func ServiceGetAllDatabase(gnoSQL *in_memory_database.GnoSQL) in_memory_database.DatabaseGetAllResult {
	var result = in_memory_database.DatabaseGetAllResult{}

	databaseNames := make([]string, 0)

	for _, database := range gnoSQL.Databases {
		databaseNames = append(databaseNames, database.DatabaseName)
	}

	result.Data = databaseNames

	return result
}

func ServiceLoadToDisk(gnoSQL *in_memory_database.GnoSQL) in_memory_database.DatabaseLoadToDiskResult {
	var result = in_memory_database.DatabaseLoadToDiskResult{}

	go gnoSQL.WriteAllDBs()

	result.Data = utils.DATABASE_LOAD_TO_DISK_MSG
	return result
}

func ServiceCreateCollections(gnoSQL *in_memory_database.GnoSQL, DatabaseName string, collectionsInput []in_memory_database.CollectionInput) in_memory_database.CollectionCreateResult {
	var result = in_memory_database.CollectionCreateResult{}

	db := gnoSQL.GetDB(DatabaseName)

	if db == nil {
		result.Error = utils.DATABASE_NOT_FOUND_MSG
		return result
	}

	db.CreateColls(collectionsInput)

	result.Data = utils.COLLECTION_CREATE_SUCCESS_MSG

	return result
}

func ServiceDeleteCollections(gnoSQL *in_memory_database.GnoSQL, DatabaseName string, collections []string) in_memory_database.CollectionDeleteResult {
	var result = in_memory_database.CollectionDeleteResult{}

	db := gnoSQL.GetDB(DatabaseName)

	if db == nil {
		result.Error = utils.DATABASE_NOT_FOUND_MSG
		return result
	}

	db.DeleteColls(collections)

	result.Data = utils.COLLECTION_DELETE_SUCCESS_MSG

	return result
}

func ServiceGetAllCollections(gnoSQL *in_memory_database.GnoSQL, DatabaseName string) in_memory_database.CollectionGetAllResult {
	var result = in_memory_database.CollectionGetAllResult{}

	db := gnoSQL.GetDB(DatabaseName)

	if db == nil {
		result.Error = utils.DATABASE_NOT_FOUND_MSG
		return result
	}

	allCollections := db.Collections

	collections := make([]string, 0)

	for _, collection := range allCollections {
		collections = append(collections, collection.CollectionName)
	}

	result.Data = collections

	return result
}

func ServiceGetCollectionStats(gnoSQL *in_memory_database.GnoSQL, DatabaseName string, CollectionName string) in_memory_database.CollectionStatsResult {
	var result = in_memory_database.CollectionStatsResult{}

	db, collection := gnoSQL.GetDatabaseAndCollection(DatabaseName, CollectionName)

	if db == nil {
		result.Error = utils.DATABASE_NOT_FOUND_MSG
		return result
	}

	if collection == nil {
		result.Error = utils.COLLECTION_NOT_FOUND_MSG
		return result
	}

	stats := collection.Stats()
	result.Data = stats

	return result
}
