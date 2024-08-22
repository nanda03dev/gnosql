package in_memory_database

import (
	"fmt"
	"gnosql/src/common"
	"os"
)

type Config MapInterface

type Database struct {
	DatabaseName string        `json:"DatabaseName"`
	Collections  []*Collection `json:"Collections"`
	Config       Config        `json:"Config"`
}

type DatabaseFileStruct struct {
	DatabaseName string `json:"DatabaseName"`
	Config       Config `json:"Config"`
}

func CreateDatabase(databaseName string, collectionsInput []CollectionInput) *Database {
	Config := make(Config)
	Config["version"] = 1

	db := &Database{
		DatabaseName: databaseName,
		Collections:  make([]*Collection, 0),
		Config:       Config,
	}

	common.CreateFolder(common.GetDatabaseFolderPath(databaseName))

	db.SaveDatabaseToFile()
	db.CreateColls(collectionsInput)

	return db
}

func LoadDatabase(database DatabaseFileStruct) *Database {
	return &Database{
		DatabaseName: database.DatabaseName,
		Collections:  make([]*Collection, 0),
		Config:       database.Config,
	}
}

func (db *Database) DeleteDatabase() {
	common.DeleteFolder(common.GetDatabaseFolderPath(db.DatabaseName))
	for _, collection := range db.Collections {
		collection.DeleteCollection(false)
	}
}

func (db *Database) CreateColls(collectionsInput []CollectionInput) []*Collection {
	var collections []*Collection = make([]*Collection, 0)
	for _, collectionInput := range collectionsInput {
		if IsCollectionExists := db.GetColl(collectionInput.CollectionName); IsCollectionExists == nil {
			collection := CreateCollection(collectionInput, db)
			db.Collections = append(db.Collections, collection)
			collections = append(collections, collection)
		}
	}
	return collections
}

func (db *Database) DeleteColls(collectionNamesToDelete []string) *Database {
	var Collections []*Collection = make([]*Collection, 0)

	for _, collection := range db.Collections {
		ToBeDeleted := false
		for _, collectionNameToDelete := range collectionNamesToDelete {
			if collectionNameToDelete == collection.CollectionName {
				ToBeDeleted = true
			}
		}

		if !ToBeDeleted {
			Collections = append(Collections, collection)
		} else {
			collection.DeleteCollection(true)
		}

	}

	db.Collections = Collections

	return db
}

func (db *Database) LoadColls(collectionsGob []CollectionFileStruct) []*Collection {
	return LoadCollections(collectionsGob)
}

func (db *Database) GetColl(collectionName string) *Collection {
	for _, collection := range db.Collections {
		if collection.CollectionName == collectionName {
			return collection
		}
	}
	return nil
}
func (db *Database) GetCollectionNames() []string {
	var colelctionNames []string
	for _, collection := range db.Collections {
		colelctionNames = append(colelctionNames, collection.CollectionName)
	}
	return colelctionNames
}

func (db *Database) SaveDatabaseToFile() {
	fmt.Printf("\n Writing to database : %s to disk \n", db.DatabaseName)

	// Convert struct to gob
	temp := DatabaseFileStruct{
		DatabaseName: db.DatabaseName,
		Config:       db.Config,
	}

	gobData, err := common.EncodeGob(temp)

	if err != nil {
		fmt.Println("GOB encoding error:", err)
	}

	err = common.SaveToFile(common.GetDatabaseFilePath(db.DatabaseName, common.GetDatabaseFileName(db.DatabaseName)), gobData)

	if err != nil {
		fmt.Println("Error saving database GOB to file:", err)
	}

	fmt.Println("GOB data saved to ", db.DatabaseName)
}

func ReadDatabaseGobFile(filePath string) (DatabaseFileStruct, error) {
	var gobData DatabaseFileStruct

	fileData, err := os.ReadFile(filePath)

	if err != nil {
		fmt.Printf("\n Datebase file %s reading, Error %v", filePath, err)
		return gobData, err
	}

	err = common.DecodeGob(fileData, &gobData)

	if err != nil {
		fmt.Printf("\n Datebase file %s decoding , Error %v", filePath, err)

		return gobData, err
	}

	return gobData, nil
}
