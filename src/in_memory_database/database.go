package in_memory_database

type Config map[string]interface{}

type Database struct {
	collections []*Collection
	config      Config
}

func CreateDatabase() *Database {
	config := make(Config)
	config["version"] = 1

	return &Database{
		collections: make([]*Collection, 0),
		config:      config,
	}
}

func (db *Database) DeleteCollections(collectionNamesToDelete []string) *Database {
	var collections []*Collection = db.collections

	for _, collectionNameToDelete := range collectionNamesToDelete {
		for collectionIndex, collection := range collections {
			if collectionNameToDelete == collection.GetCollectionName() {
				collection.Clear()

				collection.deleted = true

				collections[collectionIndex] = collection
			}
		}
	}

	db.collections = collections

	return db
}

func (db *Database) AddCollections(newCollections []CollectionInput) []*Collection {
	var oldCollections []*Collection = db.collections

	var createdCollections CollectionOutput = CreateCollections(newCollections)

	var newCollectionInstances []*Collection

	for _, collection := range createdCollections {
		newCollectionInstances = append(newCollectionInstances, collection)
		oldCollections = append(oldCollections, collection)
	}

	db.collections = oldCollections

	return newCollectionInstances
}

func (db *Database) GetCollections() []*Collection {
	return db.collections
}

func (db *Database) GetCollection(collectionName string) (*Collection, error) {
	for _, eachCollection := range db.collections {
		if eachCollection.GetCollectionName() == collectionName {
			return eachCollection, nil
		}
	}
	return nil, nil
}
