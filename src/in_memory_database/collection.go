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

type DocumentsMap map[string]Document

type DocumentIds []string

type Collection struct {
	CollectionName string   `json:"CollectionName"`
	IsDeleted      bool     `json:"IsDeleted"`
	Index          Index    `json:"Index"`     // Index map Ex: {"city" :{ chennai: [id1, ids2]}}
	IndexKeys      []string `json:"IndexKeys"` // Index keys ["city", "pincode"]
	mu             sync.RWMutex
	DocumentsMap   DocumentsMap `json:"DocumentsMap"`
	DocumentIds    DocumentIds  `json:"DocumentIds"`
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
		DocumentsMap:   make(DocumentsMap),
		DocumentIds:    make(DocumentIds, 0),
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
			DocumentsMap:   collection.DocumentsMap,
			DocumentIds:    collection.DocumentIds,
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

	var document Document = value
	document["id"] = uniqueUuid
	document["created"] = utils.ExtractTimestampFromUUID(uniqueUuid)

	collection.DocumentsMap[uniqueUuid] = document

	collection.DocumentIds = append(collection.DocumentIds, uniqueUuid)

	collection.createIndex(document)

	return value
}

func (collection *Collection) Read(id string) interface{} {
	collection.mu.RLock()
	defer collection.mu.RUnlock()
	return collection.DocumentsMap[id]
}

func (collection *Collection) Filter(filters []GenericKeyValue) interface{} {
	collection.mu.RLock()
	defer collection.mu.RUnlock()

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

	var filteredIds DocumentIds

	if len(filtersWithIndex) > 0 {
		filteredIds = collection.filterWithIndex(filtersWithIndex)
	} else {
		filteredIds = collection.DocumentIds
	}

	fmt.Printf("\n Indexing filters count: %d ", len(filtersWithIndex))
	fmt.Printf("\n Non-indexing filters count: %d ", len(filtersWithoutIndex))
	fmt.Printf("\n Scanning %d documents \n", len(filteredIds))

	// Create a channel to communicate results
	resultChannel := make(chan Document)

	workerCount := 4
	// Use a WaitGroup to wait for the goroutine to finish
	var wg sync.WaitGroup
	wg.Add(workerCount)

	filteredIdsLength := len(filteredIds)

	for i := 0; i < workerCount; i++ {
		go collection.filterWithoutIndex(&wg, resultChannel, filtersWithoutIndex, filteredIds[i*filteredIdsLength/workerCount:(i+1)*filteredIdsLength/workerCount])
	}

	// go collection.filterWithoutIndex(&wg, resultChannel, filtersWithoutIndex, filteredIds)

	// Use another goroutine to close the result channel when the filtering is done
	go func() {
		// Close the result channel to signal that no more values will be sent,
		//then only resultChannel for loop will end, otherwise it will continult wait
		defer close(resultChannel)

		// Wait for the worker goroutine to finish
		wg.Wait()

	}()

	var results = make([]interface{}, 0)

	// Retrieve the results from the channel in a loop
	for result := range resultChannel {
		results = append(results, result)
	}

	return results

}

func (collection *Collection) filterWithoutIndex(wg *sync.WaitGroup, resultChannel chan Document, filter []GenericKeyValue, filteredIds DocumentIds) {
	defer wg.Done()
	
	for _, id := range filteredIds {
		var isMatch bool = true
		document := collection.DocumentsMap[id]

		for _, filter := range filter {
			if value, ok := document[filter["key"].(string)]; ok {
				if value != filter["value"].(string) {
					isMatch = false
					break
				}
			}
		}

		if isMatch {
			resultChannel <- document
		}

	}
}

func (collection *Collection) filterWithIndex(filters []GenericKeyValue) DocumentIds {

	filteredIndexMap := make(Index)

	for _, indexMap := range filters {
		for index, indexIds := range collection.Index {
			if indexMap["key"].(string) == index {
				filteredIndexMap[index] = indexIds
			}
		}
	}

	slices.SortFunc(filters,
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

	isNotStarted := false
	resultIdsMap := make(map[string]bool)
	filteredIds := make([]string, 0)

outerLoop:
	for _, eachIndexMap := range filters {

		keyToSearch := eachIndexMap["key"].(string)
		valueToSearch := eachIndexMap["value"].(string)

		if indexMap, exists := filteredIndexMap[keyToSearch]; exists {
			if idsMap, exists := indexMap[valueToSearch]; exists {

				if !isNotStarted {
					for eachId := range idsMap {
						resultIdsMap[eachId] = true
						filteredIds = append(filteredIds, eachId)
					}
					isNotStarted = true
				} else {
					tempIdsMap := make(map[string]bool)
					tempIds := make([]string, 0)

					for eachId := range resultIdsMap {
						if _, exists := idsMap[eachId]; exists {
							tempIdsMap[eachId] = true
							tempIds = append(tempIds, eachId)
						}
					}

					if len(tempIdsMap) == 0 {
						break outerLoop
					}

					resultIdsMap = tempIdsMap
					filteredIds = tempIds
				}
			}
		}
	}
	return filteredIds
}

func (collection *Collection) Update(id string, updateInputData Document) interface{} {
	collection.mu.Lock()
	defer collection.mu.Unlock()

	if _, exists := collection.DocumentsMap[id]; !exists {
		return fmt.Errorf("id '%s' not found", id)
	}

	collection.updateIndex(collection.DocumentsMap[id], updateInputData)

	var existingData, _ = collection.DocumentsMap[id]

	for key, value := range updateInputData {
		existingData[key] = value
	}

	collection.DocumentsMap[id] = existingData

	return existingData
}

func (collection *Collection) Delete(id string) error {
	collection.mu.Lock()
	defer collection.mu.Unlock()

	if existingData, exists := collection.DocumentsMap[id]; exists {
		delete(collection.DocumentsMap, id)

		collection.deleteIndex(existingData)
	} else {
		return fmt.Errorf("id '%s' not found", id)
	}

	return nil
}

func (collection *Collection) GetIds() []string {
	collection.mu.RLock()
	defer collection.mu.RUnlock()

	keys := make([]string, 0, len(collection.DocumentsMap))
	for key := range collection.DocumentsMap {
		keys = append(keys, key)
	}
	return keys
}

func (collection *Collection) GetAllData() interface{} {
	collection.mu.RLock()
	defer collection.mu.RUnlock()

	details := make([]interface{}, 0, len(collection.DocumentsMap))

	for _, value := range collection.DocumentsMap {
		details = append(details, value)
	}

	return details
}

func (collection *Collection) Clear() {
	collection.mu.Lock()
	defer collection.mu.Unlock()
	collection.DocumentsMap = make(DocumentsMap)
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
	statsMap["documents"] = len(collection.DocumentsMap)

	return statsMap
}

func (collection *Collection) createIndex(document Document) {
	for _, eachIndex := range collection.IndexKeys {
		if indexName, ok := document[eachIndex]; ok {
			if id, ok := document["id"]; ok {
				collection.changeIndex(eachIndex, indexName.(string), id.(string), false)
			}
		}
	}
}

func (collection *Collection) updateIndex(oldDocument Document, updatedDocument Document) {
	for _, eachIndex := range collection.IndexKeys {
		if oldIndexValue, ok := oldDocument[eachIndex]; ok {
			if newIndexValue, ok := updatedDocument[eachIndex]; ok {
				var id string = oldDocument["id"].(string)
				collection.changeIndex(eachIndex, oldIndexValue.(string), id, true)
				collection.changeIndex(eachIndex, newIndexValue.(string), id, false)

			}
		}
	}
}
func (collection *Collection) deleteIndex(document Document) {
	for _, eachIndex := range collection.IndexKeys {
		if indexName, ok := document[eachIndex]; ok {
			if id, ok := document["id"]; ok {
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

		collection.Index[indexKey][indexValue][uniqueUuid] = "ok"
		// Index[name][kumar] = append([ids1,ids2], uniqueUuid)
	}
}
