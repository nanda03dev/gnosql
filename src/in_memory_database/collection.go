package in_memory_database

import (
	"cmp"
	"errors"
	"fmt"
	"gnosql/src/utils"
	"os"
	"slices"
	"sort"
	"sync"
	"time"
)

type Document map[string]interface{}

type BatchDocuments map[string]Document

type DocumentsMap map[string]BatchDocuments // Ex: { file1: {id1: {...Document1}, id2: {...Document2}}, file2: {id1: {...Document1}, id2: {...Document2}} }

type DocumentIds []string

type IndexMap map[string]IndexIdsmap //  Ex: { city :{ chennai: {id1: ok , ids2: ok}}}

type IndexIdsmap map[string]MapString // Ex: { chennai: {id1: ok , ids2: ok}}

type Event struct {
	Type      string
	Id        string
	EventData Document
}

type CollectionStats struct {
	CollectionName string   `json:"collectionName"`
	IndexKeys      []string `json:"IndexKeys"`
	Documents      int      `json:"Documents"`
}

type BatchUpdateStatus map[string]bool

const EventChannelSize = 1 * 10 * 100 * 1000

// const TimerToSaveToDisk = 1 * time.Minute
const TimerToSaveToDisk = 30 * time.Second

type Collection struct {
	CollectionName     string            `json:"CollectionName"`
	ParentDBName       string            `json:"ParentDBName"`
	IndexMap           IndexMap          `json:"IndexMap"`  // Ex: { city :{ chennai: {id1: ok , ids2: ok}}}
	IndexKeys          []string          `json:"IndexKeys"` // Ex: [ "city", "pincode"]
	DocumentsMap       DocumentsMap      `json:"DocumentsMap"`
	CollectionFileName string            `json:"CollectionFileName"`
	CollectionFullPath string            `json:"CollectionFullPath"`
	LastIndex          int               `json:"LastIndex"`
	CurrentBatchId     string            `json:"CurrentBatchId"`
	CurrentBatchCount  int               `json:"CurrentBatchCount"`
	BatchUpdateStatus  BatchUpdateStatus `json:"BatchUpdateStatus"`
	mu                 sync.RWMutex
	IsChanged          bool
}

type CollectionFileStruct struct {
	CollectionName     string            `json:"CollectionName"`
	ParentDBName       string            `json:"ParentDBName"`
	IndexMap           IndexMap          `json:"IndexMap"`  // Ex: { city :{ chennai: {id1: ok , ids2: ok}}}
	IndexKeys          []string          `json:"IndexKeys"` // Ex: [ "city", "pincode"]
	DocumentsMap       DocumentsMap      `json:"DocumentsMap"`
	CollectionFileName string            `json:"CollectionFileName"`
	CollectionFullPath string            `json:"CollectionFullPath"`
	LastIndex          int               `json:"LastIndex"`
	CurrentBatchId     string            `json:"CurrentBatchId"`
	CurrentBatchCount  int               `json:"CurrentBatchCount"`
	BatchUpdateStatus  BatchUpdateStatus `json:"BatchUpdateStatus"`
}

type CollectionInput struct {
	// Example: collectionName
	CollectionName string

	// Example: indexKeys
	IndexKeys []string
}

func CreateCollection(collectionInput CollectionInput, db *Database) *Collection {

	fileName := utils.GetCollectionFileName(collectionInput.CollectionName)
	fullPath := utils.GetCollectionFilePath(db.DatabaseName, collectionInput.CollectionName, fileName)
	currentBatchId := utils.GetCollectionBatchIdFileName()

	collection :=
		&Collection{
			CollectionName:     collectionInput.CollectionName,
			ParentDBName:       db.DatabaseName,
			IndexKeys:          collectionInput.IndexKeys,
			DocumentsMap:       make(DocumentsMap),
			IndexMap:           make(IndexMap),
			CollectionFileName: fileName,
			CollectionFullPath: fullPath,
			mu:                 sync.RWMutex{},
			IsChanged:          true,
			LastIndex:          0,
			CurrentBatchId:     currentBatchId,
			BatchUpdateStatus:  BatchUpdateStatus{currentBatchId: true},
			CurrentBatchCount:  0,
		}

	collection.SaveCollectionToFile()
	collection.StartInternalFunctions()

	return collection
}

func LoadCollections(collectionsGob []CollectionFileStruct) []*Collection {
	var collections = make([]*Collection, 0)

	for _, collectionGob := range collectionsGob {
		collection := &Collection{
			CollectionName:     collectionGob.CollectionName,
			ParentDBName:       collectionGob.ParentDBName,
			IndexKeys:          collectionGob.IndexKeys,
			DocumentsMap:       collectionGob.DocumentsMap,
			IndexMap:           collectionGob.IndexMap,
			CollectionFileName: collectionGob.CollectionFileName,
			CollectionFullPath: collectionGob.CollectionFullPath,
			LastIndex:          collectionGob.LastIndex,
			CurrentBatchId:     collectionGob.CurrentBatchId,
			CurrentBatchCount:  collectionGob.CurrentBatchCount,
			BatchUpdateStatus:  collectionGob.BatchUpdateStatus,
			mu:                 sync.RWMutex{},
			IsChanged:          false,
		}

		collection.StartInternalFunctions()

		collections = append(collections, collection)
	}
	return collections
}

func (collection *Collection) DeleteCollection(IsDbDeleted bool) {
	if !IsDbDeleted {
		utils.DeleteFile(collection.CollectionFullPath)
	}
	AddIncomingRequest(collection.ParentDBName, collection.CollectionName, Event{Type: utils.EVENT_STOP_GO_ROUTINE})
}

func (collection *Collection) Create(document Document) Document {
	collection.mu.Lock()
	defer collection.mu.Unlock()

	if document["docId"] == nil {
		document["docId"] = utils.Generate16DigitUUID()
	}

	var uniqueUuid = document["docId"].(string)
	documentIndex := collection.LastIndex + 1
	document["created"] = utils.UuidStringToTimeString(uniqueUuid)
	document["docIndex"] = documentIndex

	var batchId = collection.CurrentBatchId
	var batchCount = collection.CurrentBatchCount + 1

	if batchCount > utils.MaximumLengthNoOfDocuments {
		batchId = utils.GetCollectionBatchIdFileName()
		collection.CurrentBatchId = batchId
		batchCount = 0
	}

	if _, exists := collection.DocumentsMap[batchId]; !exists {
		collection.DocumentsMap[batchId] = make(BatchDocuments)
	}

	collection.DocumentsMap[batchId][uniqueUuid] = document

	collection.createIndex(document)

	collection.IsChanged = true
	collection.BatchUpdateStatus[batchId] = true
	collection.LastIndex = documentIndex
	collection.CurrentBatchCount = batchCount
	return document
}

func (collection *Collection) Read(id string) Document {
	collection.mu.RLock()
	defer collection.mu.RUnlock()

	var _, _, document = collection.isDocumentExists(id)
	return document
}

func (collection *Collection) Filter(reqFilter MapInterface) []Document {
	collection.mu.RLock()
	defer collection.mu.RUnlock()

	filters := make([]MapInterface, 0)
	var limit int = 1000

	for key, value := range reqFilter {
		temp := make(MapInterface)
		if key != "limit" {
			temp["key"] = key
			temp["value"] = value
			filters = append(filters, temp)
		} else {
			limit = value.(int)
		}
	}
	fmt.Printf("\n filters %+v \n", filters)

	filtersWithoutIndex := make([]MapInterface, 0)
	filtersWithIndex := make([]MapInterface, 0)

outerLoop:
	// constructing IndexFilterKeys and Non-IndexFilterKeys
	for _, filter := range filters {
		for _, indexKey := range collection.IndexKeys {
			if indexKey == filter["key"] {
				filtersWithIndex = append(filtersWithIndex, filter)
				continue outerLoop
			}
		}
		filtersWithoutIndex = append(filtersWithoutIndex, filter)
	}

	var filteredDocIds = make(DocumentIds, 0)

	// if filter have index keys, first filter ids based on
	if len(filtersWithIndex) > 0 {
		filteredDocIds = collection.GetfilteredIdsWithIndexkeys(filtersWithIndex)
	}

	// fmt.Printf("\n Indexing filters count: %d ", len(filtersWithIndex))
	// fmt.Printf("\n Non-indexing filters count: %d ", len(filtersWithoutIndex))
	// fmt.Printf("\n Scanning %d documents \n", len(filteredDocIds))

	filteredDocIdsLength := len(filteredDocIds)

	workerCount := 4
	// Use a WaitGroup to wait for the goroutine to finish
	var wg sync.WaitGroup
	wg.Add(workerCount)

	// Create a channel to communicate results
	resultChannel := make(chan Document)
	var isIndexQuery = len(filtersWithIndex) > 0

	// filter document with index query
	if isIndexQuery {
		for i := 0; i < workerCount; i++ {
			start := i * filteredDocIdsLength / workerCount
			end := (i + 1) * filteredDocIdsLength / workerCount
			go collection.filterWithIndex(&wg, resultChannel, filtersWithoutIndex, start, end, filteredDocIds)
		}
	} else {
		var allBatchIds = make([]string, 0)

		for batchId := range collection.DocumentsMap {
			allBatchIds = append(allBatchIds, batchId)
		}
		var allBatchIdsLength = len(allBatchIds)

		for i := 0; i < workerCount; i++ {
			start := i * allBatchIdsLength / workerCount
			end := (i + 1) * allBatchIdsLength / workerCount
			go collection.filterWithoutIndex(&wg, resultChannel, filtersWithoutIndex, start, end, allBatchIds)
		}

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

	sortDocuments(results)

	var limitedResult = make([]Document, 0)

	var lengthOfResult = len(results)

	if len(results) < limit {
		limit = lengthOfResult
	}

	for i := 0; i < limit; i++ {
		limitedResult = append(limitedResult, results[i])
	}

	return limitedResult
}

func (collection *Collection) filterWithIndex(wg *sync.WaitGroup, resultChannel chan Document, filters []MapInterface, start int, end int, filteredDocIds DocumentIds) {
	defer wg.Done()

	for i := start; i < end; i++ {
		id := filteredDocIds[i]

		var exists, _, document = collection.isDocumentExists(id)

		// skip checking filters if document not found
		if !exists {
			continue
		}

		var isMatch bool = IsMatchWithDocument(filters, document)

		if isMatch {
			resultChannel <- document
		}
	}
}

func (collection *Collection) filterWithoutIndex(wg *sync.WaitGroup, resultChannel chan Document, filters []MapInterface, start int, end int, allBatchIds []string) {
	defer wg.Done()

	for i := start; i < end; i++ {
		var batchDocuments = collection.DocumentsMap[allBatchIds[i]]

		for _, document := range batchDocuments {
			var isMatch bool = IsMatchWithDocument(filters, document)

			if isMatch {
				resultChannel <- document
			}
		}
	}
}

func IsMatchWithDocument(filters []MapInterface, document Document) bool {
	for _, filter := range filters {
		if value, ok := document[filter["key"].(string)]; ok {
			// Convert both value and filter["value"] to strings for comparison
			documentValueStr := fmt.Sprintf("%v", value)
			filterValueStr := fmt.Sprintf("%v", filter["value"])

			if documentValueStr != filterValueStr {
				return false
			}
		} else {
			return false
		}
	}

	return true
}

func (collection *Collection) GetfilteredIdsWithIndexkeys(filters []MapInterface) DocumentIds {
	filteredIndexMap := make(IndexMap)

	for _, filter := range filters {
		for index, indexIds := range collection.IndexMap {
			if filter["key"].(string) == index {
				filteredIndexMap[index] = indexIds
			}
		}
	}

	// Sorting index filters, using this it will fetch and query small no of records filters
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

func (collection *Collection) Update(id string, updatedDocument Document) error {
	collection.mu.Lock()
	defer collection.mu.Unlock()

	var exists, batchId, document = collection.isDocumentExists(id)

	if !exists {
		return errors.New(utils.DOCUMENT_NOT_FOUND_MSG)
	}

	collection.updateIndex(document, updatedDocument)

	collection.DocumentsMap[batchId][id] = updatedDocument

	collection.IsChanged = true
	collection.BatchUpdateStatus[batchId] = true

	return nil
}

func (collection *Collection) Delete(id string) error {
	collection.mu.Lock()
	defer collection.mu.Unlock()

	var exists, batchId, document = collection.isDocumentExists(id)

	if !exists {
		return errors.New(utils.DOCUMENT_NOT_FOUND_MSG)
	}

	delete(collection.DocumentsMap[batchId], id)
	collection.deleteIndex(document)

	collection.IsChanged = true
	collection.BatchUpdateStatus[batchId] = true

	return nil
}
func (collection *Collection) isDocumentExists(id string) (bool, string, Document) {
	var document Document
	var batchId string

	for eachBatchId, documents := range collection.DocumentsMap {

		if value, exists := documents[id]; exists {
			document = value
			batchId = eachBatchId
		}
	}

	if _, exists := document["docId"]; !exists {
		return false, batchId, document
	}

	return true, batchId, document
}

func (collection *Collection) GetAllData() []Document {
	collection.mu.RLock()
	defer collection.mu.RUnlock()

	resultDocuments := make([]Document, 0, len(collection.DocumentsMap))

	for _, documents := range collection.DocumentsMap {
		for _, document := range documents {
			resultDocuments = append(resultDocuments, document)
		}
	}

	return resultDocuments
}

func (collection *Collection) Clear() {
	collection.mu.RLock()
	defer collection.mu.RUnlock()
	collection.DocumentsMap = make(DocumentsMap)
	collection.IndexMap = make(IndexMap)
	collection.IsChanged = true
}

func (collection *Collection) Stats() CollectionStats {
	collection.mu.RLock()
	defer collection.mu.RUnlock()

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
			if id, ok := document["docId"]; ok {
				collection.changeIndex(eachIndex, indexName.(string), id.(string), false)
			}
		}
	}
}

func (collection *Collection) updateIndex(oldDocument Document, updatedDocument Document) {
	for _, eachIndex := range collection.IndexKeys {
		if oldIndexValue, ok := oldDocument[eachIndex]; ok {
			if newIndexValue, ok := updatedDocument[eachIndex]; ok {
				var id string = oldDocument["docId"].(string)
				collection.changeIndex(eachIndex, oldIndexValue.(string), id, true)
				collection.changeIndex(eachIndex, newIndexValue.(string), id, false)

			}
		}
	}
}

func (collection *Collection) deleteIndex(document Document) {
	for _, eachIndex := range collection.IndexKeys {
		if indexName, ok := document[eachIndex]; ok {
			if id, ok := document["docId"]; ok {
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
	collection.mu.Lock()
	defer collection.mu.Unlock()

	if !collection.IsChanged {
		return
	}

	temp := CollectionFileStruct{
		CollectionName:     collection.CollectionName,
		ParentDBName:       collection.ParentDBName,
		IndexKeys:          collection.IndexKeys,
		IndexMap:           collection.IndexMap,
		CollectionFileName: collection.CollectionFileName,
		CollectionFullPath: collection.CollectionFullPath,
		LastIndex:          collection.LastIndex,
		CurrentBatchId:     collection.CurrentBatchId,
		CurrentBatchCount:  collection.CurrentBatchCount,
		BatchUpdateStatus:  collection.BatchUpdateStatus,
	}

	collection.IsChanged = false

	for fileName, isUpdated := range collection.BatchUpdateStatus {
		if documents, exists := collection.DocumentsMap[fileName]; exists && isUpdated {
			gobData, err := utils.EncodeGob(documents)

			if err == nil {
				go writeGobDataToDisk(utils.GetCollectionFilePath(collection.ParentDBName, collection.CollectionName, fileName), gobData)
			} else {
				fmt.Printf("\n collection: %v \t batch filename: %v \t GOB encoding error: %v ", collection.CollectionName, fileName, err)
			}

		} else {
			fmt.Printf("\n batchid: %v does not exists in DocumentsMap ", fileName)
		}
	}

	collectionGobData, err := utils.EncodeGob(temp)
	if err == nil {
		go writeGobDataToDisk(utils.GetCollectionFilePath(collection.ParentDBName, collection.CollectionName, collection.CollectionFileName), collectionGobData)
	} else {
		fmt.Printf("\n collection: %v \t GOB encoding error: %v ", collection.CollectionName, err)
	}
}

func writeGobDataToDisk(filePath string, data []byte) {
	err := utils.SaveToFile(filePath, data)

	if err != nil {
		fmt.Printf("\n filePath: %v Error saving collection GOB to file:", err)
	}
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

func ReadAndDecodeFile[T any](filePath string) (T, error) {
	var data T

	fileData, err := os.ReadFile(filePath)

	if err != nil {
		fmt.Printf("\n Reading file %s, Error %v", filePath, err)
		return data, err
	}

	err = utils.DecodeGob(fileData, &data)

	if err != nil {
		fmt.Printf("\n Decoding file %s, Error %v", filePath, err)

		return data, err
	}

	return data, nil
}

func (collection *Collection) StartInternalFunctions() {
	go collection.EventListener()
}

func (collection *Collection) EventListener() {
	var collectionChannelName = collection.ParentDBName + collection.CollectionName
	var collectionChannel = CollectionChannelInstance.GetCollectionChannelWithLock(collection.ParentDBName, collection.CollectionName)
	for {

		event := <-collectionChannel

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
			collection.SaveCollectionToFile()
			fmt.Printf("\n EVENT_SAVE_TO_DISK : %v done\n", collectionChannelName)
		}
		if event.Type == utils.EVENT_STOP_GO_ROUTINE {
			fmt.Printf("\n %v Event channel closed. Exiting the goroutine. ", collection.CollectionName)
			return
		}

	}
}

type SortByDocIndex []Document

func (a SortByDocIndex) Len() int      { return len(a) }
func (a SortByDocIndex) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a SortByDocIndex) Less(i, j int) bool {

	iDocIndex := a[i]["docIndex"].(int)
	jDocIndex := a[j]["docIndex"].(int)

	return iDocIndex < jDocIndex
}

func sortDocuments(documents []Document) {
	sort.Sort(SortByDocIndex(documents))
}
