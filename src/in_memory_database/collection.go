package in_memory_database

import (
	"cmp"
	"fmt"
	"gnosql/src/utils"
	"os"
	"slices"
	"sync"
	"time"
)

type Document map[string]interface{}

type DocumentsMap map[string]Document // Ex: {id1: {...Document1}, id2: {...Document2} }

type DocumentIds []string

type IndexMap map[string]IndexIdsmap //  Ex: { city :{ chennai: {id1: ok , ids2: ok}}}

type IndexIdsmap map[string]MapString // Ex: { chennai: {id1: ok , ids2: ok}}

type Event struct {
	Type      string
	Id        string
	EventData Document
}

type CollectionStats struct {
	CollectionName string   `json:"CollectionName"`
	IndexKeys      []string `json:"IndexKeys"`
	Documents      int      `json:"Documents"`
}

const EventChannelSize = 1 * 10 * 100 * 1000
const TimerToSaveToDisk = 100 * time.Minute

type Collection struct {
	CollectionName     string       `json:"CollectionName"`
	IndexMap           IndexMap     `json:"IndexMap"`  // Ex: { city :{ chennai: {id1: ok , ids2: ok}}}
	IndexKeys          []string     `json:"IndexKeys"` // Ex: [ "city", "pincode"]
	DocumentsMap       DocumentsMap `json:"DocumentsMap"`
	DocumentIds        DocumentIds  `json:"DocumentIds"`
	CollectionFileName string       `json:"CollectionFileName"`
	CollectionFullPath string       `json:"CollectionFullPath"`
	EventChannel       chan Event
	mu                 sync.RWMutex
	IsChanged          bool
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
			CollectionFileName: fileName,
			CollectionFullPath: fullPath,
			mu:                 sync.RWMutex{},
			EventChannel:       make(chan Event, EventChannelSize),
			IsChanged:          true,
		}

	collection.StartInternalFunctions()

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
			EventChannel:       make(chan Event, EventChannelSize),
			mu:                 sync.RWMutex{},
			IsChanged:          false,
		}

		collection.StartInternalFunctions()

		collections = append(collections, collection)
	}
	return collections
}

func (collection *Collection) DeleteCollection() {
	utils.DeleteFile(collection.CollectionFullPath)
	collection.Clear()
}

func (collection *Collection) Create(document Document) Document {
	collection.mu.RLock()
	defer collection.mu.RUnlock()

	uniqueUuid := document["id"].(string)

	document["created"] = utils.ExtractTimestampFromUUID(uniqueUuid).String()

	collection.DocumentsMap[uniqueUuid] = document

	collection.DocumentIds = append(collection.DocumentIds, uniqueUuid)

	collection.createIndex(document)

	collection.IsChanged = true

	return document
}

func (collection *Collection) Read(id string) Document {
	return collection.DocumentsMap[id]
}

func (collection *Collection) Filter(reqFilter MapInterface) []Document {
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

	var ResultChannelSize = filteredIdsLength / workerCount

	// Create a channel to communicate results
	resultChannel := make(chan Document, ResultChannelSize)

	for i := 0; i < workerCount; i++ {
		start := i * filteredIdsLength / workerCount
		end := (i + 1) * filteredIdsLength / workerCount
		go collection.filterWithoutIndex(&wg, resultChannel, filtersWithoutIndex, filteredIds, start, end)
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

func (collection *Collection) filterWithoutIndex(wg *sync.WaitGroup, resultChannel chan Document, filter []MapInterface, filteredIds DocumentIds, start int, end int) {
	defer wg.Done()

	for i := start; i < end; i++ {
		id := filteredIds[i]
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

func (collection *Collection) Update(id string, updatedDocument Document) Document {
	collection.mu.RLock()
	defer collection.mu.RUnlock()

	if _, exists := collection.DocumentsMap[id]; !exists {
		return nil
	}

	collection.updateIndex(collection.DocumentsMap[id], updatedDocument)

	collection.DocumentsMap[id] = updatedDocument

	collection.IsChanged = true
	return updatedDocument
}

func (collection *Collection) Delete(id string) error {
	collection.mu.RLock()
	defer collection.mu.RUnlock()

	if document, exists := collection.DocumentsMap[id]; exists {
		delete(collection.DocumentsMap, id)

		collection.deleteIndex(document)
	} else {
		return fmt.Errorf("id '%s' not found", id)
	}

	collection.IsChanged = true
	return nil
}

func (collection *Collection) GetIds() []string {
	return collection.DocumentIds
}

func (collection *Collection) GetAllData() []Document {
	documents := make([]Document, 0, len(collection.DocumentsMap))

	for _, document := range collection.DocumentsMap {
		documents = append(documents, document)
	}

	return documents
}

func (collection *Collection) Clear() {
	collection.mu.RLock()
	defer collection.mu.RUnlock()
	collection.DocumentsMap = make(DocumentsMap)
	collection.IndexMap = make(IndexMap)
	collection.IsChanged = true
}

func (collection *Collection) Stats() CollectionStats {
	var statsMap = CollectionStats{
		CollectionName: collection.CollectionName,
		IndexKeys:      collection.IndexKeys,
		Documents:      len(collection.DocumentsMap),
	}
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

func (collection *Collection) changeIndex(indexKey string, indexValue string, id string, isDelete bool) {
	if _, exists := collection.IndexMap[indexKey]; !exists {
		collection.IndexMap[indexKey] = make(IndexIdsmap)
	}

	if id != "" {
		if _, exists := collection.IndexMap[indexKey][indexValue]; !exists {
			collection.IndexMap[indexKey][indexValue] = make(MapString)
			// IndexMap[city][chennai] = {city:{chennai:{id1: ok, id2:ok }}}
		}
		if isDelete {
			// delete id from map example: {city:{chennai:{ id1: ok, id2: ok }}}, delete id1 from this map
			// after deleted {city:{chennai:{ id2: ok }}}

			delete(collection.IndexMap[indexKey][indexValue], id)

			if len(collection.IndexMap[indexKey][indexValue]) < 1 {
				delete(collection.IndexMap[indexKey], indexValue)
			}
			return
		}

		collection.IndexMap[indexKey][indexValue][id] = "ok"
		// IndexMap[city][chennai] = {city:{chennai:{ id1: ok, ...exiting-ids}}}
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

	collection.IsChanged = false
}

func ConvertToCollectionInputs(collectionsInterface []interface{}) []CollectionInput {
	var collectionsInput []CollectionInput

	for _, each := range collectionsInterface {
		if collectionName, ok := each.(map[string]interface{})["CollectionName"].(string); ok {
			var indexKeys = make([]string, 0)

			for _, each := range each.(map[string]interface{})["IndexKeys"].([]interface{}) {
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

func (collection *Collection) StartTimerToSaveFile() {
	for range time.Tick(TimerToSaveToDisk) {
		collection.EventChannel <- Event{Type: utils.EVENT_SAVE_TO_DISK}
	}
}

func (collection *Collection) StartInternalFunctions() {
	go collection.StartTimerToSaveFile()
	go collection.EventListener()
}

func (collection *Collection) EventListener() {
	for event := range collection.EventChannel {

		if event.Type == utils.EVENT_CREATE {
			collection.Create(event.EventData)
		}
		if event.Type == utils.EVENT_UPDATE {
			collection.Update(event.Id, event.EventData)
		}
		if event.Type == utils.EVENT_DELETE {
			collection.Delete(event.Id)
		}
		if event.Type == utils.EVENT_SAVE_TO_DISK {
			if collection.IsChanged {
				collection.SaveCollectionToFile()
			} else {
				println("No changes occured in collection : ", collection.CollectionName)
			}
		}

	}
	fmt.Printf("\n %v Event channel closed. Exiting the goroutine. ", collection.CollectionName)
}
