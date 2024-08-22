package in_memory_database

import (
	"fmt"
	"gnosql/src/common"
	"gnosql/src/global_constants"
	"sync"
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

type Collection struct {
	CollectionName    string            `json:"CollectionName"`
	DatabaseName      string            `json:"DatabaseName"`
	IndexMap          IndexMap          `json:"IndexMap"`  // Ex: { city :{ chennai: {id1: ok , ids2: ok}}}
	IndexKeys         []string          `json:"IndexKeys"` // Ex: [ "city", "pincode"]
	DocumentsMap      DocumentsMap      `json:"DocumentsMap"`
	LastIndex         int               `json:"LastIndex"`
	CurrentBatchId    string            `json:"CurrentBatchId"`
	CurrentBatchCount int               `json:"CurrentBatchCount"`
	BatchUpdateStatus BatchUpdateStatus `json:"BatchUpdateStatus"`
	IsChanged         bool
	mu                sync.RWMutex
}

type CollectionFileStruct struct {
	CollectionName    string            `json:"CollectionName"`
	DatabaseName      string            `json:"DatabaseName"`
	IndexMap          IndexMap          `json:"IndexMap"`  // Ex: { city :{ chennai: {id1: ok , ids2: ok}}}
	IndexKeys         []string          `json:"IndexKeys"` // Ex: [ "city", "pincode"]
	DocumentsMap      DocumentsMap      `json:"DocumentsMap"`
	LastIndex         int               `json:"LastIndex"`
	CurrentBatchId    string            `json:"CurrentBatchId"`
	CurrentBatchCount int               `json:"CurrentBatchCount"`
	BatchUpdateStatus BatchUpdateStatus `json:"BatchUpdateStatus"`
}

type CollectionInput struct {
	// Example: collectionName
	CollectionName string

	// Example: indexKeys
	IndexKeys []string
}

func CreateCollection(collectionInput CollectionInput, db *Database) *Collection {
	currentBatchId := common.GetCollectionBatchIdFileName()

	collection :=
		&Collection{
			CollectionName:    collectionInput.CollectionName,
			DatabaseName:      db.DatabaseName,
			IndexKeys:         collectionInput.IndexKeys,
			DocumentsMap:      make(DocumentsMap),
			IndexMap:          make(IndexMap),
			IsChanged:         true,
			LastIndex:         0,
			CurrentBatchId:    currentBatchId,
			BatchUpdateStatus: BatchUpdateStatus{currentBatchId: true},
			CurrentBatchCount: 0,
			mu:                sync.RWMutex{},
		}

	collection.SaveCollectionToFile()
	collection.StartInternalFunctions()

	return collection
}

func LoadCollections(collectionsGob []CollectionFileStruct) []*Collection {
	var collections = make([]*Collection, 0)

	for _, collectionGob := range collectionsGob {
		collection := &Collection{
			CollectionName:    collectionGob.CollectionName,
			DatabaseName:      collectionGob.DatabaseName,
			IndexKeys:         collectionGob.IndexKeys,
			DocumentsMap:      collectionGob.DocumentsMap,
			IndexMap:          collectionGob.IndexMap,
			LastIndex:         collectionGob.LastIndex,
			CurrentBatchId:    collectionGob.CurrentBatchId,
			CurrentBatchCount: collectionGob.CurrentBatchCount,
			BatchUpdateStatus: collectionGob.BatchUpdateStatus,
			IsChanged:         false,
			mu:                sync.RWMutex{},
		}

		go collection.StartInternalFunctions()
		collections = append(collections, collection)
	}
	return collections
}

func (collection *Collection) DeleteCollection(ToBeDeleted bool) {
	if ToBeDeleted {
		common.DeleteFolder(common.GetCollectionFolderPath(collection.DatabaseName, collection.CollectionName))
	}
	AddIncomingRequest(collection.DatabaseName, collection.CollectionName, Event{Type: global_constants.EVENT_STOP_GO_ROUTINE})
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
	collection.mu.Lock()
	defer collection.mu.Unlock()

	collection.CollectionName = ""
	collection.DatabaseName = ""
	collection.IndexMap = make(IndexMap)         // Reset to an empty map
	collection.IndexKeys = nil                   // Reset to nil (or make([]string, 0) for an empty slice)
	collection.DocumentsMap = make(DocumentsMap) // Reset to an empty map
	collection.LastIndex = 0
	collection.CurrentBatchId = ""
	collection.CurrentBatchCount = 0
	collection.BatchUpdateStatus = make(BatchUpdateStatus) // Reset to an empty map
	collection.IsChanged = false
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
		if indexValue, ok := document[eachIndex]; ok {
			if id, ok := document[global_constants.DOC_ID]; ok {
				collection.changeIndex(eachIndex, indexValue.(string), id.(string), false)
			}
		}
	}
}

func (collection *Collection) updateIndex(oldDocument Document, updatedDocument Document) {
	for _, eachIndex := range collection.IndexKeys {
		if oldIndexValue, ok := oldDocument[eachIndex]; ok {
			if newIndexValue, ok := updatedDocument[eachIndex]; ok {
				var id string = oldDocument[global_constants.DOC_ID].(string)
				collection.changeIndex(eachIndex, oldIndexValue.(string), id, true)
				collection.changeIndex(eachIndex, newIndexValue.(string), id, false)

			}
		}
	}
}

func (collection *Collection) deleteIndex(document Document) {
	for _, eachIndex := range collection.IndexKeys {
		if indexValue, ok := document[eachIndex]; ok {
			if id, ok := document[global_constants.DOC_ID]; ok {
				collection.changeIndex(eachIndex, indexValue.(string), id.(string), true)
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

func ConvertToCollectionInputs(collectionsInterface []interface{}) []CollectionInput {
	var collectionsInput []CollectionInput

	for _, each := range collectionsInterface {
		if collectionName, ok := each.(map[string]interface{})[global_constants.COLLECTION_NAME].(string); ok {
			var indexKeys = make([]string, 0)

			for _, each := range each.(map[string]interface{})[global_constants.INDEX_KEYS_NAME].([]interface{}) {
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

func (collection *Collection) SaveCollectionToFile() {
	collection.mu.Lock()
	defer collection.mu.Unlock()

	if !collection.IsChanged {
		return
	}

	temp := CollectionFileStruct{
		CollectionName:    collection.CollectionName,
		DatabaseName:      collection.DatabaseName,
		IndexKeys:         collection.IndexKeys,
		IndexMap:          collection.IndexMap,
		LastIndex:         collection.LastIndex,
		CurrentBatchId:    collection.CurrentBatchId,
		CurrentBatchCount: collection.CurrentBatchCount,
		BatchUpdateStatus: collection.BatchUpdateStatus,
	}

	collection.IsChanged = false

	// Write collection file to disk
	collectionGobData, err := common.EncodeGob(temp)
	if err == nil {
		var collectionFileName = common.GetCollectionFileName(collection.CollectionName)
		go common.WriteGobDataToDisk(common.GetCollectionFilePath(collection.DatabaseName, collection.CollectionName, collectionFileName), collectionGobData)
	} else {
		fmt.Printf("\n collection: %v \t GOB encoding error: %v ", collection.CollectionName, err)
	}

	// Write Batch file to disk
	for fileName, isUpdated := range collection.BatchUpdateStatus {
		if documents, exists := collection.DocumentsMap[fileName]; exists && isUpdated {
			gobData, err := common.EncodeGob(documents)
			if err == nil {
				go common.WriteGobDataToDisk(common.GetCollectionFilePath(collection.DatabaseName, collection.CollectionName, fileName), gobData)
			} else {
				fmt.Printf("\n collection: %v \t batch filename: %v \t GOB encoding error: %v ", collection.CollectionName, fileName, err)
			}

		} else {
			fmt.Printf("\n batchid: %v does not exists in DocumentsMap ", fileName)
		}
		collection.BatchUpdateStatus[fileName] = false
	}

}

func (collection *Collection) StartInternalFunctions() {
	go collection.StartMutationWorker()
}
