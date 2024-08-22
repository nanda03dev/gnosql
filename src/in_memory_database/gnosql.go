package in_memory_database

import (
	"fmt"
	"gnosql/src/common"
	"gnosql/src/global_constants"
	"path/filepath"
	"strings"
)

type GnoSQL struct {
	Databases []*Database
}

func CreateGnoSQL() *GnoSQL {
	gnoSQL := &GnoSQL{
		Databases: make([]*Database, 0),
	}
	return gnoSQL
}

func (gnoSQL *GnoSQL) CreateDB(databaseName string, collectionsInput []CollectionInput) *Database {
	var db *Database = CreateDatabase(databaseName, collectionsInput)
	gnoSQL.Databases = append(gnoSQL.Databases, db)
	return db
}

func (gnoSQL *GnoSQL) LoadDB(database DatabaseFileStruct) *Database {
	var db *Database = LoadDatabase(database)
	gnoSQL.Databases = append(gnoSQL.Databases, db)
	return db
}

func (gnoSQL *GnoSQL) DeleteDB(db *Database) bool {
	var databases []*Database = make([]*Database, 0)

	for _, database := range gnoSQL.Databases {
		if database.DatabaseName != db.DatabaseName {
			databases = append(databases, database)
		} else {
			database.DeleteDatabase()
		}
	}

	gnoSQL.Databases = databases

	return true
}

func (gnoSQL *GnoSQL) LoadAllDBs() {
	// Read all database folder from gnosqlpath
	databaseFolders, err := common.ReadFoldersInDirectory(global_constants.GNOSQL_FULL_PATH)
	if err != nil {
		fmt.Println("Error while reading database folders", fmt.Sprintf("%v", err))
	}

	fmt.Printf("\n Loading databases ")
	// Read database and all colelctions one by one
	for _, eachDatabaseFolder := range databaseFolders {
		fileNames, err := common.ReadFileNamesInDirectory(eachDatabaseFolder)
		if err != nil {
			fmt.Println("Error while reading collection files", fmt.Sprintf("%v", err))
		}

		var db *Database
		var collectionFileStructs []CollectionFileStruct

		// filter fileName "-db.gob", "-collection.gob"
		for _, fileName := range fileNames {
			if strings.Contains(fileName, global_constants.DB_EXTENSION) {
				if databaseGob, err := ReadDatabaseGobFile(fileName); err == nil {
					db = gnoSQL.LoadDB(databaseGob)
				}
			}

		}

		collectionFolders, err := common.ReadFoldersInDirectory(eachDatabaseFolder)

		if err != nil {
			fmt.Println("Error while reading collection folder files", fmt.Sprintf("%v", err))
		}

		for _, eachCollectionFolder := range collectionFolders {
			fileNames, err := common.ReadFileNamesInDirectory(eachCollectionFolder)
			if err != nil {
				fmt.Println("Error while reading collection files", fmt.Sprintf("%v", err))
			}

			var collectionFile CollectionFileStruct
			var documentMaps DocumentsMap = make(DocumentsMap)

			for _, fileName := range fileNames {
				if strings.Contains(fileName, global_constants.COLLECTION_EXTENSION) {
					if collectionGob, err := common.ReadFileAndDecodeGOB[CollectionFileStruct](fileName); err == nil {
						collectionFile = collectionGob
					}
				}
				if strings.Contains(fileName, global_constants.COLLECTION_BATCH_EXTENSION) {
					if collectionDataGob, err := common.ReadFileAndDecodeGOB[BatchDocuments](fileName); err == nil {
						var dataFileName = filepath.Base(fileName)
						if strings.Contains(dataFileName, global_constants.COLLECTION_BATCH_EXTENSION) {
							documentMaps[dataFileName] = collectionDataGob
						}
					}
				}
			}

			collectionFile.DocumentsMap = documentMaps
			collectionFileStructs = append(collectionFileStructs, collectionFile)

		}

		db.Collections = db.LoadColls(collectionFileStructs)
		fmt.Printf("\n\t Database Name : %v ", db.DatabaseName)
		fmt.Printf("\n\t Collections Names : %v \n", db.GetCollectionNames())

	}
	fmt.Printf("\n ----- All databases loaded ----- \n")

}

func (gnoSQL *GnoSQL) GetDB(databaseName string) *Database {
	for _, database := range gnoSQL.Databases {
		if database.DatabaseName == databaseName {
			return database
		}
	}
	return nil
}

func (gnoSQL *GnoSQL) GetDatabaseAndCollection(databaseName string, collectionName string) (*Database, *Collection) {
	for _, database := range gnoSQL.Databases {
		if database.DatabaseName == databaseName {
			for _, collection := range database.Collections {
				if collection.CollectionName == collectionName {
					return database, collection
				}
			}
			return database, nil
		}
	}
	return nil, nil
}

func (gnoSQL *GnoSQL) WriteAllDBs() {
	for _, database := range gnoSQL.Databases {
		database.SaveDatabaseToFile()

		for _, collection := range database.Collections {
			collection.SaveCollectionToFile()
		}
	}
}
