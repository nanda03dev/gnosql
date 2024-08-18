package in_memory_database

import (
	"fmt"
	"time"
)

type IncomeRequest struct {
	DatabaseName   string
	CollectionName string
	Event          Event
}

var IncomeRequestChannel chan IncomeRequest = make(chan IncomeRequest, 100000)

func init() {
	go InitializeWorker()
}

func AddIncomingRequest(databaseName string, collectionName string, event Event) {
	incomingRequest := IncomeRequest{
		DatabaseName:   databaseName,
		CollectionName: collectionName,
		Event:          event,
	}
	IncomeRequestChannel <- incomingRequest
}

func InitializeWorker() {
	go startWorkerWithRecovery(StartIncomeRequestWorker)
}

func startWorkerWithRecovery(workerFunc func()) {
	for {
		func() {
			defer func() {
				if r := recover(); r != nil {
					fmt.Printf("Worker panic recovered: %v. Restarting worker...\n", r)
				}
			}()

			workerFunc()
		}()

		// Optional: Add a small delay before restarting the worker
		time.Sleep(2 * time.Second)
	}
}

func StartIncomeRequestWorker() {
	for {
		incomeRequest := <-IncomeRequestChannel
		CollectionChannelInstance.AddCollectionEvent(incomeRequest.DatabaseName, incomeRequest.CollectionName, incomeRequest.Event)
	}
}
