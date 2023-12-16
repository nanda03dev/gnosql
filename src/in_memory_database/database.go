package in_memory_database

import (
	"fmt"
	"gnosql/src/utils"
	"os"
)

type Config MapInterface

type Database struct {
	DatabaseName         string        `json:"DatabaseName"`
	DatabaseFileName     string        `json:"DatabaseFileName"`
	DatabaseFileFullPath string        `json:"DatabaseFileFullPath"`
	DatabaseFolderPath   string        `json:"DatabaseFolderPath"`
	Collections          []*Collection `json:"Collections"`
	Config               Config        `json:"Config"`
	IsDeleted            bool          `json:"IsDeleted"`
}

type DatabaseFileStruct struct {
	DatabaseName         string `json:"DatabaseName"`
	DatabaseFileName     string `json:"DatabaseFileName"`
	DatabaseFileFullPath string `json:"DatabaseFileFullPath"`
	DatabaseFolderPath   string `json:"DatabaseFolderPath"`
	Config               Config `json:"Config"`
	IsDeleted            bool   `json:"IsDeleted"`
}

func (db *Database) ClearDatabase() {
	os.RemoveAll(db.DatabaseFolderPath)

	for _, collection := range db.Collections {
		collection.Clear()
	}
}

func (db *Database) CreateCollections(collectionsInput []CollectionInput) []*Collection {
	var collections []*Collection = make([]*Collection, 0)
	for _, collectionInput := range collectionsInput {
		if IsCollectionExists := db.GetCollection(collectionInput.CollectionName); IsCollectionExists == nil {
			collection := CreateCollection(collectionInput, db)
			db.Collections = append(db.Collections, collection)
			collections = append(collections, collection)
		}
	}

	return collections
}

func (db *Database) DeleteCollections(collectionNamesToDelete []string) *Database {
	var Collections []*Collection = db.Collections

	for _, collectionNameToDelete := range collectionNamesToDelete {
		for collectionIndex, collection := range Collections {
			if collectionNameToDelete == collection.CollectionName {
				collection.Clear()

				collection.IsDeleted = true

				Collections[collectionIndex] = collection
			}
		}
	}

	db.Collections = Collections

	return db
}

func (db *Database) GetCollection(collectionName string) *Collection {
	for _, collection := range db.Collections {
		if collection.CollectionName == collectionName {
			return collection
		}
	}
	return nil
}

func (db *Database) SaveDatabaseToFile() {
	fmt.Printf("\n Writing to database : %s to disk \n", db.DatabaseName)

	// Convert struct to gob
	temp := DatabaseFileStruct{
		DatabaseName:         db.DatabaseName,
		DatabaseFileName:     db.DatabaseFileName,
		DatabaseFileFullPath: db.DatabaseFileFullPath,
		Config:               db.Config,
		IsDeleted:            db.IsDeleted,
	}

	gobData, err := utils.EncodeGob(temp)

	if err != nil {
		fmt.Println("GOB encoding error:", err)
	}

	err = utils.SaveToFile(db.DatabaseFileFullPath, gobData)

	if err != nil {
		fmt.Println("Error saving GOB to file:", err)
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

	err = utils.DecodeGob(fileData, &gobData)

	if err != nil {
		fmt.Printf("\n Datebase file %s decoding , Error %v", filePath, err)

		return gobData, err
	}

	return gobData, nil
}
