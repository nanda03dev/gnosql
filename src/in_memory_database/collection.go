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
	// Example: collectionName
	CollectionName string

	// Example: indexKeys
	IndexKeys []string
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

func (collection *Collection) Create(value Document) interface{} {
	collection.mu.Lock()
	defer collection.mu.Unlock()

	uniqueUuid := utils.Generate16DigitUUID()

	value["id"] = uniqueUuid
	value["created"] = utils.ExtractTimestampFromUUID(uniqueUuid)
	collection.DataMap[uniqueUuid] = value

	collection.createIndex(value)

	return value
}

func (collection *Collection) Read(id string) interface{} {
	collection.mu.RLock()
	defer collection.mu.RUnlock()
	return collection.DataMap[id]
}

func (collection *Collection) Filter(filters []GenericKeyValue) interface{} {
	collection.mu.RLock()
	defer collection.mu.RUnlock()

	var results []interface{}

	filtersWithoutIndex := make([]GenericKeyValue, 0)
	filtersWithIndex := make([]GenericKeyValue, 0)

outerLoop:
	for _, filter := range filters {
		for _, indexKey := range collection.IndexKeys {
			if indexKey == filter["key"] {
				filtersWithIndex = append(filtersWithIndex, filter)
				continue outerLoop
			}
		}
		filtersWithoutIndex = append(filtersWithoutIndex, filter)
	}

	var filteredData DataMap

	if len(filtersWithIndex) > 0 {
		filteredData = collection.filterDataByIndex(filtersWithIndex)
	} else {
		filteredData = collection.DataMap
	}

	fmt.Printf("\n Indexing filters count: %d ", len(filtersWithIndex))
	fmt.Printf("\n Non-indexing filters count: %d ", len(filtersWithoutIndex))
	fmt.Printf("\n Scanning %d documents \n", len(filteredData))

	for id := range filteredData {
		var isMatch bool = true
		document := collection.DataMap[id]
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

func (collection *Collection) FilterByIndexKey(request []GenericKeyValue) interface{} {
	collection.mu.RLock()
	defer collection.mu.RUnlock()

	var results = make([]interface{}, 0)

	resultIds := collection.filterDataByIndex(request)

	for eachId := range resultIds {
		results = append(results, collection.DataMap[eachId])
	}

	return results
}

func (collection *Collection) filterDataByIndex(request []GenericKeyValue) DataMap {
	isNotStarted := false
	resultIds := make(DataMap)
	filteredIndexMap := make(Index)

	for _, indexMap := range request {
		for index, indexIds := range collection.Index {
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

func (collection *Collection) Update(id string, updateInputData Document) interface{} {
	collection.mu.Lock()
	defer collection.mu.Unlock()

	if _, exists := collection.DataMap[id]; !exists {
		return fmt.Errorf("id '%s' not found", id)
	}

	collection.updateIndex(collection.DataMap[id], updateInputData)

	var existingData, _ = collection.DataMap[id]

	for key, value := range updateInputData {
		existingData[key] = value
	}

	collection.DataMap[id] = existingData

	return existingData
}

func (collection *Collection) Delete(id string) error {
	collection.mu.Lock()
	defer collection.mu.Unlock()

	if existingData, exists := collection.DataMap[id]; exists {
		delete(collection.DataMap, id)

		collection.deleteIndex(existingData)
	} else {
		return fmt.Errorf("id '%s' not found", id)
	}

	return nil
}

func (collection *Collection) GetIds() []string {
	collection.mu.RLock()
	defer collection.mu.RUnlock()

	keys := make([]string, 0, len(collection.DataMap))
	for key := range collection.DataMap {
		keys = append(keys, key)
	}
	return keys
}

func (collection *Collection) GetAllData() interface{} {
	collection.mu.RLock()
	defer collection.mu.RUnlock()

	details := make([]interface{}, 0, len(collection.DataMap))

	for _, value := range collection.DataMap {
		details = append(details, value)
	}

	return details
}

func (collection *Collection) Clear() {
	collection.mu.Lock()
	defer collection.mu.Unlock()
	collection.DataMap = make(DataMap)
	collection.Index = make(Index)
}

func (collection *Collection) GetIndexData() interface{} {
	collection.mu.RLock()
	defer collection.mu.RUnlock()

	return collection.Index
}

func (collection *Collection) Stats() interface{} {
	collection.mu.RLock()
	defer collection.mu.RUnlock()
	statsMap := make(map[string]interface{})
	statsMap["collectionName"] = collection.CollectionName
	statsMap["indexKeys"] = collection.IndexKeys
	statsMap["documents"] = len(collection.DataMap)

	return statsMap
}

func (collection *Collection) createIndex(body Document) {
	for _, eachIndex := range collection.IndexKeys {
		if indexName, ok := body[eachIndex]; ok {
			if id, ok := body["id"]; ok {
				collection.changeIndex(eachIndex, indexName.(string), id.(string), false)
			}
		}
	}
}

func (collection *Collection) updateIndex(oldData, updatedData Document) {
	for _, eachIndex := range collection.IndexKeys {
		if oldIndexValue, ok := oldData[eachIndex]; ok {
			if newIndexValue, ok := updatedData[eachIndex]; ok {
				var id string = oldData["id"].(string)
				collection.changeIndex(eachIndex, oldIndexValue.(string), id, true)
				collection.changeIndex(eachIndex, newIndexValue.(string), id, false)

			}
		}
	}
}
func (collection *Collection) deleteIndex(body Document) {
	for _, eachIndex := range collection.IndexKeys {
		if indexName, ok := body[eachIndex]; ok {
			if id, ok := body["id"]; ok {
				collection.changeIndex(eachIndex, indexName.(string), id.(string), true)
			}
		}
	}
}

func (collection *Collection) changeIndex(indexKey string, indexValue string, uniqueUuid string, isDelete bool) {
	if _, exists := collection.Index[indexKey]; !exists {
		collection.Index[indexKey] = make(IndexValue)
	}

	if uniqueUuid != "" {
		if _, exists := collection.Index[indexKey][indexValue]; !exists {
			collection.Index[indexKey][indexValue] = make(map[string]string)
			// Index[name][kumar] = {name:{nanda:{id1:id1, id2:id2 }}}
		}
		if isDelete {
			// delete id from map ex {name:{nanda:{id1:id1, id2:id2 }}}, delete id1 from this map

			delete(collection.Index[indexKey][indexValue], uniqueUuid)

			if len(collection.Index[indexKey][indexValue]) < 1 {
				delete(collection.Index[indexKey], indexValue)
			}
			return
		}

		collection.Index[indexKey][indexValue][uniqueUuid] = "Ok"
		// Index[name][kumar] = append([ids1,ids2], uniqueUuid)
	}
}
