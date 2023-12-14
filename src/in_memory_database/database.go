package in_memory_database

import (
	"fmt"
	"gnosql/src/utils"
	"os"
	"time"
)

type Config MapInterface

type Database struct {
	DatabaseName         string        `json:"DatabaseName"`
	DatabaseFileName     string        `json:"DatabaseFileName"`
	DatabaseFileFullPath string        `json:"DatabaseFileFullPath"`
	Collections          []*Collection `json:"Collections"`
	Config               Config        `json:"Config"`
	IsDeleted            bool          `json:"IsDeleted"`
}

func (db *Database) ClearDatabase() {
	utils.DeleteFile(db.DatabaseFileFullPath)

	for _, collection := range db.Collections {
		collection.Clear()
	}
}

func (db *Database) CreateCollections(collectionsInput []CollectionInput) []*Collection {
	var collections []*Collection = make([]*Collection, 0)
	for _, collectionInput := range collectionsInput {
		if IsCollectionExists := db.GetCollection(collectionInput.CollectionName); IsCollectionExists == nil {
			collection := CreateCollection(collectionInput)
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

// func (db *Database) SaveToFile() error {
// 	fmt.Printf("\n Writing to database : %s to disk \n", db.DatabaseName)

// 	data, err := json.Marshal(db)
// 	if err != nil {
// 		return err
// 	}

// 	return os.WriteFile(db.DatabaseFileFullPath, data, 0644)
// }

func (db *Database) SaveToFile() {
	fmt.Printf("\n Writing to database : %s to disk \n", db.DatabaseName)

	// Convert struct to gob
	gobData, err := utils.EncodeGob(db)

	if err != nil {
		fmt.Println("GOB encoding error:", err)
	}

	// Save gob to file
	err = utils.SaveToFile(db.DatabaseFileFullPath, gobData)
	if err != nil {
		fmt.Println("Error saving GOB to file:", err)
	}
	fmt.Println("GOB data saved to data.gob")
}

func (db *Database) StartTimerToSaveFile() {
	for range time.Tick(2 * time.Hour) {
		go db.SaveToFile()
	}
}

// func ReadDatabaseJsonFile(filePath string) (Database, error) {
// 	var jsonData Database

// 	// Read the Databse JSON file
// 	fileData, err := os.ReadFile(filePath)

// 	if err != nil {
// 		fmt.Printf("\n Datebase file %s reading, Error %v", filePath, err)
// 		return jsonData, err
// 	}

// 	err = json.Unmarshal(fileData, &jsonData)

// 	if err != nil {
// 		fmt.Printf("\n Datebase file %s Unmarshall , Error %v", filePath, err)

// 		return jsonData, err
// 	}

// 	return jsonData, nil
// }

func ReadDatabaseGobFile(filePath string) (Database, error) {
	var gobData Database

	// Read the Databse JSON file
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
