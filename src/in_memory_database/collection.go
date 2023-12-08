package in_memory_database

import (
	"cmp"
	"fmt"
	"gnosql/src/utils"
	"slices"
	"sync"
)

type GenericKeyValue map[string]interface{}

type IndexValue map[string]map[string]string

type Index map[string]IndexValue

type Document map[string]interface{}

type DataMap map[string]Document

type Collection struct {
	CollectionName string   `json:"CollectionName"`
	IsDeleted      bool     `json:"IsDeleted"`
	Index          Index    `json:"Index"`     // Index map Ex: {"city" :{ chennai: [id1, ids2]}}
	IndexKeys      []string `json:"IndexKeys"` // Index keys ["city", "pincode"]
	mu             sync.RWMutex
	DataMap        DataMap `json:"DataMap"`
}

type CollectionInput struct {
	CollectionName string
	IndexKeys      []string
}

func CreateCollection(collectionInput CollectionInput) *Collection {
	return &Collection{
		CollectionName: collectionInput.CollectionName,
		IndexKeys:      collectionInput.IndexKeys,
		DataMap:        make(DataMap),
		Index:          make(Index),
		mu:             sync.RWMutex{},
		IsDeleted:      false,
	}
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

func (db *Collection) Create(value Document) interface{} {
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

	fmt.Printf("\n Indexing filters count: %d ", len(filtersWithIndex))
	fmt.Printf("\n Non-indexing filters count: %d ", len(filtersWithoutIndex))
	fmt.Printf("\n Scanning %d documents \n", len(filteredData))

	for id := range filteredData {
		var isMatch bool = true
		document := db.DataMap[id]
		for _, filter := range filtersWithoutIndex {
			if value, ok := document[filter["key"].(string)]; ok {
				if value != filter["value"].(string) {
					isMatch = false
					break
				}
			}
		}

		if isMatch {
			results = append(results, document)
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
	filteredIndexMap := make(Index)

	for _, indexMap := range request {
		for index, indexIds := range db.Index {
			if indexMap["key"].(string) == index {
				filteredIndexMap[index] = indexIds
			}
		}
	}

	slices.SortFunc(request,
		func(a, b GenericKeyValue) int {
			keyToSearchA := a["key"].(string)
			valueToSearchA := a["value"].(string)

			keyToSearchB := b["key"].(string)
			valueToSearchB := b["value"].(string)

			indexIdsLenA := len(filteredIndexMap[keyToSearchA][valueToSearchA])
			//20 := len(filteredIndexMap[city][chennai]) chennai - 1000 - users
			indexIdsLenB := len(filteredIndexMap[keyToSearchB][valueToSearchB])
			//10 := len(filteredIndexMap[pincode][60100]) 600100 - 20 - users
			return cmp.Compare(indexIdsLenA, indexIdsLenB)
		})

outerLoop:
	for _, eachIndexMap := range request {

		keyToSearch := eachIndexMap["key"].(string)
		valueToSearch := eachIndexMap["value"].(string)

		if indexMap, exists := filteredIndexMap[keyToSearch]; exists {
			if idsMap, exists := indexMap[valueToSearch]; exists {

				if !isNotStarted {
					for eachId := range idsMap {
						resultIds[eachId] = Document{}
					}
					isNotStarted = true
				} else {
					tempIds := make(DataMap)
					for eachId := range resultIds {
						if _, exists := idsMap[eachId]; exists {
							tempIds[eachId] = Document{}
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

func (db *Collection) Update(id string, updateInputData Document) interface{} {
	db.mu.Lock()
	defer db.mu.Unlock()

	if _, exists := db.DataMap[id]; !exists {
		return fmt.Errorf("id '%s' not found", id)
	}

	db.updateIndex(db.DataMap[id], updateInputData)

	var existingData, _ = db.DataMap[id]

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

		db.deleteIndex(existingData)
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

func (db *Collection) createIndex(body Document) {
	for _, eachIndex := range db.IndexKeys {
		if indexName, ok := body[eachIndex]; ok {
			if id, ok := body["id"]; ok {
				db.changeIndex(eachIndex, indexName.(string), id.(string), false)
			}
		}
	}
}

func (db *Collection) updateIndex(oldData, updatedData Document) {
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
func (db *Collection) deleteIndex(body Document) {
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
