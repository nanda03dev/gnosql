package in_memory_database

import (
	"cmp"
	"fmt"
	"gnosql/src/utils"
	"os"
	"slices"
	"sync"
)

type Document map[string]interface{}

type DocumentsMap map[string]Document // Ex: {id1: {...Document1}, id2: {...Document2} }

type DocumentIds []string

type IndexMap map[string]IndexIdsmap //  Ex: { city :{ chennai: {id1: ok , ids2: ok}}}

type IndexIdsmap map[string]MapString // Ex: { chennai: {id1: ok , ids2: ok}}

type Collection struct {
	CollectionName     string       `json:"CollectionName"`
	IndexMap           IndexMap     `json:"IndexMap"`  // Ex: { city :{ chennai: {id1: ok , ids2: ok}}}
	IndexKeys          []string     `json:"IndexKeys"` // Ex: [ "city", "pincode"]
	DocumentsMap       DocumentsMap `json:"DocumentsMap"`
	DocumentIds        DocumentIds  `json:"DocumentIds"`
	CollectionFileName string       `json:"CollectionFileName"`
	CollectionFullPath string       `json:"CollectionFullPath"`
	mu                 sync.RWMutex
}

type CollectionFileStruct struct {
	CollectionName     string       `json:"CollectionName"`
	IndexMap           IndexMap     `json:"IndexMap"`  // Ex: { city :{ chennai: {id1: ok , ids2: ok}}}
	IndexKeys          []string     `json:"IndexKeys"` // Ex: [ "city", "pincode"]
	DocumentsMap       DocumentsMap `json:"DocumentsMap"`
	DocumentIds        DocumentIds  `json:"DocumentIds"`
	CollectionFileName string       `json:"CollectionFileName"`
	CollectionFullPath string       `json:"CollectionFullPath"`
}

type CollectionInput struct {
	// Example: collectionName
	CollectionName string

	// Example: indexKeys
	IndexKeys []string
}

func CreateCollection(collectionInput CollectionInput, db *Database) *Collection {

	fileName := utils.GetCollectionFileName(collectionInput.CollectionName)
	fullPath := utils.GetCollectionFilePath(db.DatabaseName, fileName)

	collection :=
		&Collection{
			CollectionName:     collectionInput.CollectionName,
			IndexKeys:          collectionInput.IndexKeys,
			DocumentsMap:       make(DocumentsMap),
			DocumentIds:        make(DocumentIds, 0),
			IndexMap:           make(IndexMap),
			mu:                 sync.RWMutex{},
			CollectionFileName: fileName,
			CollectionFullPath: fullPath,
		}

	collection.SaveCollectionToFile()

	return collection
}

func LoadCollections(collectionsGob []CollectionFileStruct) []*Collection {
	var collections = make([]*Collection, 0)

	for _, collectionGob := range collectionsGob {
		collection := &Collection{
			CollectionName:     collectionGob.CollectionName,
			IndexKeys:          collectionGob.IndexKeys,
			DocumentsMap:       collectionGob.DocumentsMap,
			DocumentIds:        collectionGob.DocumentIds,
			IndexMap:           collectionGob.IndexMap,
			CollectionFileName: collectionGob.CollectionFileName,
			CollectionFullPath: collectionGob.CollectionFullPath,
			mu:                 sync.RWMutex{},
		}

		collections = append(collections, collection)
	}

	return collections
}

func (collection *Collection) DeleteCollection() {
	utils.DeleteFile(collection.CollectionFullPath)
	collection.Clear()
}

func (collection *Collection) Create(value Document) Document {
	collection.mu.Lock()
	defer collection.mu.Unlock()

	uniqueUuid := utils.Generate16DigitUUID()

	var document Document = value
	document["id"] = uniqueUuid
	document["created"] = utils.ExtractTimestampFromUUID(uniqueUuid).String()

	collection.DocumentsMap[uniqueUuid] = document

	collection.DocumentIds = append(collection.DocumentIds, uniqueUuid)

	collection.createIndex(document)

	return document
}

func (collection *Collection) Read(id string) Document {
	collection.mu.RLock()
	defer collection.mu.RUnlock()
	return collection.DocumentsMap[id]
}

func (collection *Collection) Filter(reqFilter MapInterface) []Document {
	collection.mu.RLock()
	defer collection.mu.RUnlock()

	filters := make([]MapInterface, 0)

	for key, value := range reqFilter {
		temp := make(MapInterface)
		temp["key"] = key
		temp["value"] = value
		filters = append(filters, temp)
	}

	filtersWithoutIndex := make([]MapInterface, 0)
	filtersWithIndex := make([]MapInterface, 0)

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

	// if filter have index keys, first filter ids based on
	if len(filtersWithIndex) > 0 {
		filteredIds = collection.filterWithIndex(filtersWithIndex)
	} else {
		filteredIds = collection.DocumentIds
	}

	fmt.Printf("\n Indexing filters count: %d ", len(filtersWithIndex))
	fmt.Printf("\n Non-indexing filters count: %d ", len(filtersWithoutIndex))
	fmt.Printf("\n Scanning %d documents \n", len(filteredIds))

	filteredIdsLength := len(filteredIds)

	workerCount := 4
	// Use a WaitGroup to wait for the goroutine to finish
	var wg sync.WaitGroup
	wg.Add(workerCount)

	// Create a channel to communicate results
	resultChannel := make(chan Document, filteredIdsLength/workerCount)

	for i := 0; i < workerCount; i++ {
		go collection.filterWithoutIndex(&wg, resultChannel, filtersWithoutIndex, filteredIds[i*filteredIdsLength/workerCount:(i+1)*filteredIdsLength/workerCount])
	}

	// Use another goroutine to close the result channel when the filtering is done
	go func() {
		// Close the result channel to signal that no more values will be sent,
		//then only resultChannel for loop will end, otherwise it will continult wait
		defer close(resultChannel)

		// Wait for the worker goroutine to finish
		wg.Wait()

	}()

	var results = make([]Document, 0)

	// Retrieve the results from the channel in a loop
	for result := range resultChannel {
		results = append(results, result)
	}

	return results

}

func (collection *Collection) filterWithoutIndex(wg *sync.WaitGroup, resultChannel chan Document, filter []MapInterface, filteredIds DocumentIds) {
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

func (collection *Collection) filterWithIndex(filters []MapInterface) DocumentIds {

	filteredIndexMap := make(IndexMap)

	for _, filter := range filters {
		for index, indexIds := range collection.IndexMap {
			if filter["key"].(string) == index {
				filteredIndexMap[index] = indexIds
			}
		}
	}

	slices.SortFunc(filters,
		func(a, b MapInterface) int {
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

func (collection *Collection) Update(id string, updateInputData Document) Document {
	collection.mu.Lock()
	defer collection.mu.Unlock()

	if _, exists := collection.DocumentsMap[id]; !exists {
		return nil
	}

	collection.updateIndex(collection.DocumentsMap[id], updateInputData)

	var existingDocument, _ = collection.DocumentsMap[id]

	for key, value := range updateInputData {
		existingDocument[key] = value
	}

	collection.DocumentsMap[id] = existingDocument

	return existingDocument
}

func (collection *Collection) Delete(id string) error {
	collection.mu.Lock()
	defer collection.mu.Unlock()

	if document, exists := collection.DocumentsMap[id]; exists {
		delete(collection.DocumentsMap, id)

		collection.deleteIndex(document)
	} else {
		return fmt.Errorf("id '%s' not found", id)
	}

	return nil
}

func (collection *Collection) GetIds() []string {
	return collection.DocumentIds
}

func (collection *Collection) GetAllData() []Document {
	collection.mu.RLock()
	defer collection.mu.RUnlock()

	documents := make([]Document, 0, len(collection.DocumentsMap))

	for _, document := range collection.DocumentsMap {
		documents = append(documents, document)
	}

	return documents
}

func (collection *Collection) Clear() {
	collection.mu.Lock()
	defer collection.mu.Unlock()
	collection.DocumentsMap = make(DocumentsMap)
	collection.IndexMap = make(IndexMap)
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
	if _, exists := collection.IndexMap[indexKey]; !exists {
		collection.IndexMap[indexKey] = make(IndexIdsmap)
	}

	if uniqueUuid != "" {
		if _, exists := collection.IndexMap[indexKey][indexValue]; !exists {
			collection.IndexMap[indexKey][indexValue] = make(MapString)
			// IndexMap[name][kumar] = {name:{nanda:{id1:id1, id2:id2 }}}
		}
		if isDelete {
			// delete id from map ex {name:{nanda:{id1:id1, id2:id2 }}}, delete id1 from this map

			delete(collection.IndexMap[indexKey][indexValue], uniqueUuid)

			if len(collection.IndexMap[indexKey][indexValue]) < 1 {
				delete(collection.IndexMap[indexKey], indexValue)
			}
			return
		}

		collection.IndexMap[indexKey][indexValue][uniqueUuid] = "ok"
		// IndexMap[name][kumar] = append([ids1,ids2], uniqueUuid)
	}
}

func (collection *Collection) SaveCollectionToFile() {
	fmt.Printf("\n Writing to collection : %s to disk \n", collection.CollectionName)

	temp := CollectionFileStruct{
		CollectionName:     collection.CollectionName,
		IndexKeys:          collection.IndexKeys,
		DocumentsMap:       collection.DocumentsMap,
		DocumentIds:        collection.DocumentIds,
		IndexMap:           collection.IndexMap,
		CollectionFileName: collection.CollectionFileName,
		CollectionFullPath: collection.CollectionFullPath,
	}

	gobData, err := utils.EncodeGob(temp)

	if err != nil {
		fmt.Println("GOB encoding error:", err)
	}

	err = utils.SaveToFile(collection.CollectionFullPath, gobData)

	if err != nil {
		fmt.Println("Error saving collection GOB to file:", err)
	}

	fmt.Println("GOB data saved to ", collection.CollectionName)
}

func ConvertToCollectionInputs(collectionsInterface []interface{}) []CollectionInput {
	var collectionsInput []CollectionInput

	for _, each := range collectionsInterface {
		if collectionName, ok := each.(map[string]interface{})["collectionName"].(string); ok {
			var indexKeys = make([]string, 0)

			for _, each := range each.(map[string]interface{})["indexKeys"].([]interface{}) {
				indexKeys = append(indexKeys, each.(string))
			}

			collectionInput := CollectionInput{
				CollectionName: collectionName,
				IndexKeys:      indexKeys,
			}

			collectionsInput = append(collectionsInput, collectionInput)
		}
	}
	return collectionsInput
}

func ReadCollectionGobFile(filePath string) (CollectionFileStruct, error) {
	var gobData CollectionFileStruct

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
