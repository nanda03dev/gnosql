package in_memory_database

import (
	"fmt"
	"gnosql/src/utils"
	"os"
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

func (collection *Collection) DeleteCollection(ToBeDeleted bool) {
	if ToBeDeleted {
		utils.DeleteFolder(utils.GetCollectionFolderPath(collection.ParentDBName, collection.CollectionName))
	}
	AddIncomingRequest(collection.ParentDBName, collection.CollectionName, Event{Type: utils.EVENT_STOP_GO_ROUTINE})
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
