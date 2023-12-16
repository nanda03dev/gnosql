package in_memory_database

import (
	"fmt"
	"gnosql/src/utils"
	"strings"
)

type GnoSQL struct {
	Databases []*Database
}

func CreateGnoSQL() *GnoSQL {
	gnoSQL := &GnoSQL{
		Databases: make([]*Database, 0),
	}

	// TODO function added, not tested, keeping this for future use
	// runtime.SetFinalizer(gnoSQL, cleanupFunction)

	return gnoSQL
}

func (gnoSQL *GnoSQL) CreateDatabase(databaseName string, collectionsInput []CollectionInput) *Database {
	Config := make(Config)
	Config["version"] = 1

	databasePath := utils.GetDatabaseFolderPath(databaseName)
	fileName := utils.GetDatabaseFileName(databaseName)
	filePath := utils.GetDatabaseFilePath(databaseName, fileName)

	db := &Database{
		DatabaseName:         databaseName,
		DatabaseFileName:     fileName,
		DatabaseFileFullPath: filePath,
		DatabaseFolderPath:   databasePath,
		Collections:          make([]*Collection, 0),
		Config:               Config,
		IsDeleted:            false,
	}

	utils.CreateFolder(databasePath)
	db.SaveDatabaseToFile()
	db.CreateCollections(collectionsInput)
	gnoSQL.Databases = append(gnoSQL.Databases, db)

	return db
}

func (gnoSQL *GnoSQL) LoadDatabase(dbValues DatabaseFileStruct) *Database {

	db := &Database{
		DatabaseName:         dbValues.DatabaseName,
		DatabaseFileName:     dbValues.DatabaseFileName,
		DatabaseFileFullPath: dbValues.DatabaseFileFullPath,
		Collections:          make([]*Collection, 0),
		Config:               dbValues.Config,
	}

	gnoSQL.Databases = append(gnoSQL.Databases, db)

	return db
}

func (gnoSQL *GnoSQL) DeleteDatabase(db *Database) bool {
	var databases []*Database = make([]*Database, 0)

	for _, database := range gnoSQL.Databases {
		if database.DatabaseName != db.DatabaseName {
			// database.IsDeleted = true
			databases = append(databases, database)
		} else {
			database.ClearDatabase()
		}
	}

	gnoSQL.Databases = databases

	return true
}

func (gnoSQL *GnoSQL) LoadAllDatabases() []*Database {
	var databases = make([]*Database, 0)

	folders, err := utils.ReadFoldersInDirectory(utils.GNOSQLFULLPATH)

	if err != nil {
		print("error while reading folders ")
		return databases
	}

	fmt.Println("Database folders:", fmt.Sprintf("%v", folders))

	for _, eachDatabaseFolder := range folders {
		fileNames, err := utils.ReadFileNamesInDirectory(eachDatabaseFolder)
		fmt.Println("database & collections files:", fmt.Sprintf("%v", fileNames))

		if err != nil {
			return databases
		}

		var db *Database
		var collectionsGobData []CollectionFileStruct

		for _, fileName := range fileNames {
			if strings.Contains(fileName, "-db.gob") {
				println("Loading database ", fileName)
				if databaseJson, err := ReadDatabaseGobFile(fileName); err == nil {
					if !databaseJson.IsDeleted {
						db = gnoSQL.LoadDatabase(databaseJson)
						databases = append(databases, db)
					}
				}
			}
		}

		for _, fileName := range fileNames {
			if strings.Contains(fileName, "-collection.gob") {
				println("Loading collection ", fileName)
				if collectionGobData, err := ReadCollectionGobFile(fileName); err == nil {
					if !collectionGobData.IsDeleted {
						collectionsGobData = append(collectionsGobData, collectionGobData)
					}
				}
			}
		}

		db.Collections = LoadCollections(collectionsGobData)
	}

	return databases
}

func (gnoSQL *GnoSQL) GetDatabase(databaseName string) *Database {
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

func (gnoSQL *GnoSQL) WriteAllDatabases() {
	for _, database := range gnoSQL.Databases {
		database.SaveDatabaseToFile()

		for _, collection := range database.Collections {
			collection.SaveCollectionToFile()
		}
	}
}

// // Function to be executed before the object is deleted
// func cleanupFunction(gnoSQL *GnoSQL) {
// 	println(" cleanupFunction called once its garbage collected ", gnoSQL.Databases)
// 	// Add your cleanup logic here
// 	// This function will be executed when the object is garbage collected
// }
