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
	fmt.Printf("\n CollectionChannelInstance initialzed successfully %v \n", CollectionChannelInstance)
	go CollectionChannelInstance.StartTimerToSaveFile()
}

func (cc *CollectionChannel) AddCollectionEvent(databaseName string, collectionName string, event Event) {
	cc.mu.Lock()
	defer cc.mu.Unlock()

	var channel = GetCollectionChannel(databaseName, collectionName)
	channel <- event
}

func (cc *CollectionChannel) GetCollectionChannelWithLock(databaseName string, collectionName string) chan Event {
	cc.mu.Lock()
	defer cc.mu.Unlock()

	return GetCollectionChannel(databaseName, collectionName)
}

func (cc *CollectionChannel) AddCollectionEventByChannelName(channelName string, event Event) {
	cc.mu.Lock()
	defer cc.mu.Unlock()

	var channel = CollectionChannelInstance.channels[channelName]
	channel <- event
}

func (cc *CollectionChannel) GetAllCollections() []string {
	cc.mu.Lock()
	defer cc.mu.Unlock()
	var channelNames []string

	for channelName := range cc.channels {
		channelNames = append(channelNames, channelName)
	}
	return channelNames

}

func (cc *CollectionChannel) StartTimerToSaveFile() {
	for range time.Tick(TimerToSaveToDisk) {
		fmt.Printf("\n ---------------------------------------------------------------- \n")
		fmt.Printf("\n Ticker started. getting all channels\n")
		for _, channelName := range cc.GetAllCollections() {
			cc.AddCollectionEventByChannelName(channelName, Event{Type: utils.EVENT_SAVE_TO_DISK})
		}
		fmt.Printf("\n EVENT_SAVE_TO_DISK event sent to all colelction channels \n")
		fmt.Printf("\n ---------------------------------------------------------------- \n")

	}
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
