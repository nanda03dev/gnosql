package in_memory_database

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

type Config map[string]interface{}

type Database struct {
	DatabaseName         string        `json:"DatabaseName"`
	DatabaseFileName     string        `json:"DatabaseFileName"`
	DatabaseFileFullPath string        `json:"DatabaseFileFullPath"`
	Collections          []*Collection `json:"Collections"`
	Config               Config        `json:"Config"`
	IsDeleted            bool          `json:"IsDeleted"`
}

func (db *Database) AddCollections(newCollections []CollectionInput) []*Collection {
	var oldCollections []*Collection = db.Collections

	var createdCollections CollectionOutput = CreateCollections(newCollections)

	var newCollectionInstances []*Collection

	for _, collection := range createdCollections {
		newCollectionInstances = append(newCollectionInstances, collection)
		oldCollections = append(oldCollections, collection)
	}

	db.Collections = oldCollections

	return newCollectionInstances
}

func (db *Database) DeleteCollections(collectionNamesToDelete []string) *Database {
	var Collections []*Collection = db.Collections

	for _, collectionNameToDelete := range collectionNamesToDelete {
		for collectionIndex, collection := range Collections {
			if collectionNameToDelete == collection.GetCollectionName() {
				collection.Clear()

				collection.IsDeleted = true

				Collections[collectionIndex] = collection
			}
		}
	}

	db.Collections = Collections

	return db
}

func (db *Database) GetCollections() []*Collection {
	return db.Collections
}

func (db *Database) GetCollection(collectionName string) (*Collection, error) {
	for _, eachCollection := range db.Collections {
		if eachCollection.GetCollectionName() == collectionName {
			return eachCollection, nil
		}
	}
	return nil, nil
}

func (db *Database) SaveToFile() error {
	fmt.Printf("\n Writing to database : %s to disk \n", db.DatabaseName)

	data, err := json.MarshalIndent(db, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(db.DatabaseFileFullPath, data, 0644)
}

func (db *Database) StartTimerToSaveFile() {
	for range time.Tick(30 * time.Second) {
		go db.SaveToFile()
	}
}

func ReadDatabaseJSONFile(filePath string) (Database, error) {
	var jsonData Database

	// Read the Databse JSON file
	fileData, err := os.ReadFile(filePath)

	if err != nil {
		fmt.Printf("\n Datebase file %s reading, Error %v", filePath, err)
		return jsonData, err
	}

	err = json.Unmarshal(fileData, &jsonData)

	if err != nil {
		fmt.Printf("\n Datebase file %s Unmarshall , Error %v", filePath, err)

		return jsonData, err
	}

	return jsonData, nil
}
