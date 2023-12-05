package in_memory_database

import (
	"fmt"
	"gnosql/src/utils"
	"sync"
)

type GenericKeyValue map[string]interface{}

type IndexValue map[string][]string

type Index map[string]IndexValue

type Data map[string]interface{}

type DocumentInput map[string]interface{}

type Collection struct {
	collectionName string
	deleted        bool
	index          Index    // index map Ex: {"city" :{ chennai: [id1, ids2]}}
	indexKeys      []string // index keys ["city"]
	mu             sync.RWMutex
	data           Data
}

type CollectionInput struct {
	CollectionName string
	IndexKeys      []string
}

type CollectionOutput map[string]*Collection

func CreateCollections(collectionsInput []CollectionInput) CollectionOutput {
	collectionInstances := make(CollectionOutput)

	for _, collectionInput := range collectionsInput {
		collectionInstance := &Collection{
			collectionName: collectionInput.CollectionName,
			indexKeys:      collectionInput.IndexKeys,
			data:           make(Data),
			index:          make(Index),
			mu:             sync.RWMutex{},
			deleted:        false,
		}
		collectionInstances[collectionInput.CollectionName] = collectionInstance

	}

	return collectionInstances
}

func (db *Collection) GetCollectionName() string {
	return db.collectionName
}

func (db *Collection) IsDeleted() bool {
	return db.deleted
}

func (db *Collection) Create(value DocumentInput) interface{} {
	db.mu.Lock()
	defer db.mu.Unlock()

	uniqueUuid := utils.Generate16DigitUUID()

	value["id"] = uniqueUuid
	value["created"] = utils.ExtractTimestampFromUUID(uniqueUuid)
	db.data[uniqueUuid] = value

	db.createIndex(value)

	return value
}

func (db *Collection) Read(id string) interface{} {
	db.mu.RLock()
	defer db.mu.RUnlock()
	return db.data[id]
}

func (db *Collection) Filter(filters []GenericKeyValue) interface{} {
	db.mu.RLock()
	defer db.mu.RUnlock()

	var results []interface{}

	for _, eachData := range db.data {
		isMatch := true
		for _, filter := range filters {
			if value, ok := eachData.(DocumentInput)[filter["key"].(string)]; ok {
				if value != filter["value"].(string) {
					isMatch = false
					break
				}
			}
		}
		if isMatch {
			results = append(results, eachData)
		}

	}
	return results

}

func (db *Collection) FilterByIndexKey(request []GenericKeyValue) interface{} {
	db.mu.RLock()
	defer db.mu.RUnlock()

	var results = make([]interface{}, 0)

	isNotStarted := false

	resultIds := make(map[string]bool)

outerLoop:
	for _, eachIndexMap := range request {
		if indexMap, exists := db.index[eachIndexMap["key"].(string)]; exists {
			if ids, exists := indexMap[eachIndexMap["value"].(string)]; exists {
				if !isNotStarted {
					for _, eachId := range ids {
						resultIds[eachId] = true
					}
					isNotStarted = true
				} else {
					tempIds := make(map[string]bool)
					for _, eachId := range ids {
						if resultIds[eachId] {
							tempIds[eachId] = true
						}
					}
					if len(tempIds) == 0 {
						break outerLoop
					}
					resultIds = tempIds
				}
			}
		}
	}

	for eachId := range resultIds {
		results = append(results, db.data[eachId])
	}

	return results
}

func (db *Collection) Update(id string, updateInputData DocumentInput) interface{} {
	db.mu.Lock()
	defer db.mu.Unlock()

	if _, exists := db.data[id]; !exists {
		return fmt.Errorf("id '%s' not found", id)
	}

	db.updateIndex(db.data[id].(DocumentInput), updateInputData)

	var existingData, _ = db.data[id].(DocumentInput)

	for key, value := range updateInputData {
		existingData[key] = value
	}

	db.data[id] = existingData

	return existingData
}

func (db *Collection) Delete(id string) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	if existingData, exists := db.data[id]; exists {
		delete(db.data, id)

		db.deleteIndex(existingData.(DocumentInput))
	} else {
		return fmt.Errorf("id '%s' not found", id)
	}

	return nil
}

func (db *Collection) GetIds() []string {
	db.mu.RLock()
	defer db.mu.RUnlock()

	keys := make([]string, 0, len(db.data))
	for key := range db.data {
		keys = append(keys, key)
	}
	return keys
}

func (db *Collection) GetAllData() interface{} {
	db.mu.RLock()
	defer db.mu.RUnlock()

	details := make([]interface{}, 0, len(db.data))

	for _, value := range db.data {
		details = append(details, value)
	}

	return details
}

func (db *Collection) Clear() {
	db.mu.Lock()
	defer db.mu.Unlock()
	db.data = make(Data)
	db.index = make(Index)
}

func (db *Collection) GetIndexData() interface{} {
	db.mu.RLock()
	defer db.mu.RUnlock()

	return db.index
}

func (db *Collection) createIndex(body DocumentInput) {
	for _, eachIndex := range db.indexKeys {
		if indexName, ok := body[eachIndex]; ok {
			if id, ok := body["id"]; ok {
				db.changeIndex(eachIndex, indexName.(string), id.(string), false)
			}
		}
	}
}

func (db *Collection) updateIndex(oldData, updatedData DocumentInput) {
	for _, eachIndex := range db.indexKeys {
		if oldIndexValue, ok := oldData[eachIndex]; ok {
			if newIndexValue, ok := updatedData[eachIndex]; ok {
				var id string = oldData["id"].(string)
				db.changeIndex(eachIndex, oldIndexValue.(string), id, true)
				db.changeIndex(eachIndex, newIndexValue.(string), id, false)

			}
		}
	}
}
func (db *Collection) deleteIndex(body DocumentInput) {
	for _, eachIndex := range db.indexKeys {
		if indexName, ok := body[eachIndex]; ok {
			if id, ok := body["id"]; ok {
				db.changeIndex(eachIndex, indexName.(string), id.(string), true)
			}
		}
	}
}

func (db *Collection) changeIndex(indexKey string, indexValue string, uniqueUuid string, isDelete bool) {
	if _, exists := db.index[indexKey]; !exists {
		db.index[indexKey] = make(IndexValue)
	}

	if uniqueUuid != "" {
		if _, exists := db.index[indexKey][indexValue]; !exists {
			db.index[indexKey][indexValue] = make([]string, 0, 100000)
		}
		if isDelete {
			// delete id from array
			var updatedIndexValues = utils.DeleteElement(db.index[indexKey][indexValue], uniqueUuid)
			if len(updatedIndexValues) > 0 {
				db.index[indexKey][indexValue] = updatedIndexValues
			} else {
				delete(db.index[indexKey], indexValue)
			}
			return
		}

		db.index[indexKey][indexValue] = append(db.index[indexKey][indexValue], uniqueUuid)
		// index[name][kumar] = append([ids1,ids2], uniqueUuid)
	}
}
