package in_memory_database

import (
	"errors"
	"fmt"
	"gnosql/src/common"
	"gnosql/src/global_constants"
)

func (collection *Collection) StartMutationWorker() {
	var collectionChannelName = collection.DatabaseName + collection.CollectionName
	var collectionChannel = CollectionChannelInstance.GetCollectionChannelWithLock(collection.DatabaseName, collection.CollectionName)

	for {
		event := <-collectionChannel

		if event.Type == global_constants.EVENT_CREATE {
			collection.Create(event.EventData)
		}
		if event.Type == global_constants.EVENT_UPDATE {
			collection.Update(event.Id, event.EventData)
		}
		if event.Type == global_constants.EVENT_DELETE {
			collection.Delete(event.Id)
		}
		if event.Type == global_constants.EVENT_SAVE_TO_DISK {
			collection.SaveCollectionToFile()
			fmt.Printf("\n EVENT_SAVE_TO_DISK : %v done\n", collectionChannelName)
		}
		if event.Type == global_constants.EVENT_STOP_GO_ROUTINE {
			collection.Clear()
			fmt.Printf("\n %v Event channel closed. Exiting the goroutine. ", collection.CollectionName)
			return
		}

	}
}

func (collection *Collection) Create(document Document) Document {
	collection.mu.Lock()
	defer collection.mu.Unlock()

	if document[global_constants.DOC_ID] == nil {
		document[global_constants.DOC_ID] = common.Generate16DigitUUID()
	}

	var uniqueUuid = document[global_constants.DOC_ID].(string)
	documentIndex := collection.LastIndex + 1
	document[global_constants.DOC_CREATED_AT] = common.UuidStringToTimeString(uniqueUuid)
	document[global_constants.DOC_INDEX] = documentIndex

	var batchId = collection.CurrentBatchId
	var batchCount = collection.CurrentBatchCount + 1

	if batchCount > global_constants.BATCH_SIZE {
		batchId = common.GetCollectionBatchIdFileName()
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

func (collection *Collection) Update(id string, updatedDocument Document) error {
	collection.mu.Lock()
	defer collection.mu.Unlock()

	var exists, batchId, document = collection.isDocumentExists(id)

	if !exists {
		return errors.New(global_constants.DOCUMENT_NOT_FOUND_MSG)
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
		return errors.New(global_constants.DOCUMENT_NOT_FOUND_MSG)
	}

	delete(collection.DocumentsMap[batchId], id)
	collection.deleteIndex(document)

	collection.IsChanged = true
	collection.BatchUpdateStatus[batchId] = true

	return nil
}
