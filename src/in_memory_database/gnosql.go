package in_memory_database

import (
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

	fileName := utils.GetDatabaseFileName(databaseName)
	filePath := utils.GetDatabaseFilePath(fileName)

	db := &Database{
		DatabaseName:         databaseName,
		DatabaseFileName:     fileName,
		DatabaseFileFullPath: filePath,
		Collections:          make([]*Collection, 0),
		Config:               Config,
		IsDeleted:            false,
	}

	db.CreateCollections(collectionsInput)

	db.SaveToFile()

	gnoSQL.Databases = append(gnoSQL.Databases, db)

	return db
}

func (gnoSQL *GnoSQL) LoadDatabase(dbValues Database) *Database {

	db := &Database{
		DatabaseName:         dbValues.DatabaseName,
		DatabaseFileName:     dbValues.DatabaseFileName,
		DatabaseFileFullPath: dbValues.DatabaseFileFullPath,
		Collections:          LoadCollections(dbValues.Collections),
		Config:               dbValues.Config,
	}

	go db.StartTimerToSaveFile()

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

	fileNames, err := utils.ReadFileNamesInDirectory(utils.GNOSQLFULLPATH)

	if err != nil {
		return databases
	}

	for _, fileName := range fileNames {
		if strings.Contains(fileName, "-db.json") {
			println("Loading database ", fileName)
			if databaseJson, err := ReadDatabaseJSONFile(fileName); err == nil {
				if !databaseJson.IsDeleted {
					var db *Database = gnoSQL.LoadDatabase(databaseJson)
					databases = append(databases, db)
				}
			}

		}
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
		database.SaveToFile()
	}
}

// // Function to be executed before the object is deleted
// func cleanupFunction(gnoSQL *GnoSQL) {
// 	println(" cleanupFunction called once its garbage collected ", gnoSQL.Databases)
// 	// Add your cleanup logic here
// 	// This function will be executed when the object is garbage collected
// }
