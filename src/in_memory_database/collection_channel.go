package in_memory_database

import (
	"fmt"
	"gnosql/src/utils"
	"sync"
	"time"
)

type ChannelMap map[string]chan Event

type CollectionChannel struct {
	channels ChannelMap
	mu       sync.RWMutex
}

var (
	CollectionChannelInstance *CollectionChannel
)

func NewCollectionChannel() *CollectionChannel {
	return &CollectionChannel{channels: make(ChannelMap)}
}

func init() {
	CollectionChannelInstance = NewCollectionChannel()
	fmt.Printf("\n CollectionChannelInstance initialzed successfully %v ", CollectionChannelInstance)
	go CollectionChannelInstance.StartTimerToSaveFile()
}

func (cc *CollectionChannel) AddCollectionEvent(databaseName string, collectionName string, event Event) {
	cc.mu.Lock()
	defer cc.mu.Unlock()

	var channel = GetCollectionChannel(databaseName, collectionName)
	channel <- event
}

func GetCollectionChannel(databaseName string, collectionName string) chan Event {
	var channelName = databaseName + collectionName

	if _, isExists := CollectionChannelInstance.channels[channelName]; !isExists {
		CollectionChannelInstance.channels[channelName] = make(chan Event, 10000)
	}

	var channel = CollectionChannelInstance.channels[channelName]

	return channel
}

func DeleteCollectionChannel(databaseName string, collectionName string) {
	var channelName = databaseName + collectionName

	if _, isExists := CollectionChannelInstance.channels[channelName]; isExists {
		delete(CollectionChannelInstance.channels, channelName)
	}
}

func (cc *CollectionChannel) StartTimerToSaveFile() {
	for range time.Tick(TimerToSaveToDisk) {
		cc.mu.Lock()
		defer cc.mu.Unlock()

		for _, channel := range cc.channels {
			channel <- Event{Type: utils.EVENT_SAVE_TO_DISK}
		}
	}
}
