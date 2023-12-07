package in_memory_database

import (
	"fmt"
	"gnosql/src/utils"
	"sync"
)

type GenericKeyValue map[string]interface{}

type IndexValue map[string]map[string]string

type Index map[string]IndexValue

type DataMap map[string]interface{}

type Document map[string]interface{}

type DocumentInput map[string]interface{}

type Collection struct {
	CollectionName string   `json:"CollectionName"`
	IsDeleted      bool     `json:"IsDeleted"`
	Index          Index    `json:"Index"`     // Index map Ex: {"city" :{ chennai: [id1, ids2]}}
	IndexKeys      []string `json:"IndexKeys"` // Index keys ["city"]
	mu             sync.RWMutex
	DataMap        DataMap `json:"DataMap"`
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
			CollectionName: collectionInput.CollectionName,
			IndexKeys:      collectionInput.IndexKeys,
			DataMap:        make(DataMap),
			Index:          make(Index),
			mu:             sync.RWMutex{},
			IsDeleted:      false,
		}
		collectionInstances[collectionInput.CollectionName] = collectionInstance

	}

	return collectionInstances
}

func LoadCollections(collections []*Collection) []*Collection {
	var collectionInstances = make([]*Collection, 0)

	for _, collection := range collections {
		collectionInstance := &Collection{
			CollectionName: collection.CollectionName,
			IndexKeys:      collection.IndexKeys,
			DataMap:        collection.DataMap,
			Index:          collection.Index,
			mu:             sync.RWMutex{},
			IsDeleted:      collection.IsDeleted,
		}

		collectionInstances = append(collectionInstances, collectionInstance)
	}

	return collectionInstances
}

func (db *Collection) GetCollectionName() string {
	return db.CollectionName
}

func (db *Collection) Create(value DocumentInput) interface{} {
	db.mu.Lock()
	defer db.mu.Unlock()

	uniqueUuid := utils.Generate16DigitUUID()

	value["id"] = uniqueUuid
	value["created"] = utils.ExtractTimestampFromUUID(uniqueUuid)
	db.DataMap[uniqueUuid] = value

	db.createIndex(value)

	return value
}

func (db *Collection) Read(id string) interface{} {
	db.mu.RLock()
	defer db.mu.RUnlock()
	return db.DataMap[id]
}

func (db *Collection) Filter(filters []GenericKeyValue) interface{} {
	db.mu.RLock()
	defer db.mu.RUnlock()

	var results []interface{}

	filtersWithoutIndex := make([]GenericKeyValue, 0)
	filtersWithIndex := make([]GenericKeyValue, 0)

outerLoop:
	for _, filter := range filters {
		for _, indexKey := range db.IndexKeys {
			if indexKey == filter["key"] {
				filtersWithIndex = append(filtersWithIndex, filter)
				continue outerLoop
			}
		}
		filtersWithoutIndex = append(filtersWithoutIndex, filter)
	}

	var filteredData DataMap

	if len(filtersWithIndex) > 0 {
		filteredData = db.filterDataByIndex(filtersWithIndex)
	} else {
		filteredData = db.DataMap
	}

	println(" indexing filters count %v ", len(filtersWithIndex))
	println(" Non-indexing filters count %v ", len(filtersWithoutIndex))
	println(" Scanning %v documents", len(filteredData))

	for id := range filteredData {
		isMatch := true
		eachData := db.DataMap[id]
		for _, filter := range filtersWithoutIndex {
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

	resultIds := db.filterDataByIndex(request)

	for eachId := range resultIds {
		results = append(results, db.DataMap[eachId])
	}

	return results
}

func (db *Collection) filterDataByIndex(request []GenericKeyValue) DataMap {
	isNotStarted := false
	resultIds := make(DataMap)

outerLoop:
	for _, eachIndexMap := range request {
		if indexMap, exists := db.Index[eachIndexMap["key"].(string)]; exists {
			if idsMap, exists := indexMap[eachIndexMap["value"].(string)]; exists {
				if !isNotStarted {
					for eachId := range idsMap {
						resultIds[eachId] = true
					}
					isNotStarted = true
				} else {
					tempIds := make(DataMap)

					for eachId := range resultIds {
						if _, exists := idsMap[eachId]; exists {
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
	return resultIds
}

func (db *Collection) Update(id string, updateInputData DocumentInput) interface{} {
	db.mu.Lock()
	defer db.mu.Unlock()

	if _, exists := db.DataMap[id]; !exists {
		return fmt.Errorf("id '%s' not found", id)
	}

	db.updateIndex(db.DataMap[id].(DocumentInput), updateInputData)

	var existingData, _ = db.DataMap[id].(DocumentInput)

	for key, value := range updateInputData {
		existingData[key] = value
	}

	db.DataMap[id] = existingData

	return existingData
}

func (db *Collection) Delete(id string) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	if existingData, exists := db.DataMap[id]; exists {
		delete(db.DataMap, id)

		db.deleteIndex(existingData.(DocumentInput))
	} else {
		return fmt.Errorf("id '%s' not found", id)
	}

	return nil
}

func (db *Collection) GetIds() []string {
	db.mu.RLock()
	defer db.mu.RUnlock()

	keys := make([]string, 0, len(db.DataMap))
	for key := range db.DataMap {
		keys = append(keys, key)
	}
	return keys
}

func (db *Collection) GetAllData() interface{} {
	db.mu.RLock()
	defer db.mu.RUnlock()

	details := make([]interface{}, 0, len(db.DataMap))

	for _, value := range db.DataMap {
		details = append(details, value)
	}

	return details
}

func (db *Collection) Clear() {
	db.mu.Lock()
	defer db.mu.Unlock()
	db.DataMap = make(DataMap)
	db.Index = make(Index)
}

func (db *Collection) GetIndexData() interface{} {
	db.mu.RLock()
	defer db.mu.RUnlock()

	return db.Index
}

func (db *Collection) createIndex(body DocumentInput) {
	for _, eachIndex := range db.IndexKeys {
		if indexName, ok := body[eachIndex]; ok {
			if id, ok := body["id"]; ok {
				db.changeIndex(eachIndex, indexName.(string), id.(string), false)
			}
		}
	}
}

func (db *Collection) updateIndex(oldData, updatedData DocumentInput) {
	for _, eachIndex := range db.IndexKeys {
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
	for _, eachIndex := range db.IndexKeys {
		if indexName, ok := body[eachIndex]; ok {
			if id, ok := body["id"]; ok {
				db.changeIndex(eachIndex, indexName.(string), id.(string), true)
			}
		}
	}
}

func (db *Collection) changeIndex(indexKey string, indexValue string, uniqueUuid string, isDelete bool) {
	if _, exists := db.Index[indexKey]; !exists {
		db.Index[indexKey] = make(IndexValue)
	}

	if uniqueUuid != "" {
		if _, exists := db.Index[indexKey][indexValue]; !exists {
			db.Index[indexKey][indexValue] = make(map[string]string)
			// Index[name][kumar] = {name:{nanda:{id1:id1, id2:id2 }}}
		}
		if isDelete {
			// delete id from map ex {name:{nanda:{id1:id1, id2:id2 }}}, delete id1 from this map

			delete(db.Index[indexKey][indexValue], uniqueUuid)

			if len(db.Index[indexKey][indexValue]) < 1 {
				delete(db.Index[indexKey], indexValue)
			}
			return
		}

		db.Index[indexKey][indexValue][uniqueUuid] = "Ok"
		// Index[name][kumar] = append([ids1,ids2], uniqueUuid)
	}
}
