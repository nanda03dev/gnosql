package in_memory_database

import (
	"cmp"
	"fmt"
	"gnosql/src/global_constants"
	"slices"
	"sort"
	"sync"
)

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
	var limit int = global_constants.FILTER_DEFAULT_LIMIT

	for key, value := range reqFilter {
		temp := make(MapInterface)
		if key != global_constants.FILTER_LIMIT {
			temp[global_constants.FILTER_KEY] = key
			temp[global_constants.FILTER_VALUE] = value
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
			if indexKey == filter[global_constants.FILTER_KEY] {
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

	workerCount := global_constants.FILTER_DEFAULT_WORKER_COUNT

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
		if valueFromDoc, ok := document[filter[global_constants.FILTER_KEY].(string)]; ok {
			// Convert both value and filter["key"] to strings for comparison
			documentValueStr := fmt.Sprintf("%v", valueFromDoc)
			filterValueStr := fmt.Sprintf("%v", filter[global_constants.FILTER_VALUE])

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
			if filter[global_constants.FILTER_KEY].(string) == index {
				filteredIndexMap[index] = indexIds
			}
		}
	}

	// Sorting index filters, using this it will fetch and query small no of records filters
	slices.SortFunc(filters,
		func(a, b MapInterface) int {
			keyToSearchA := a[global_constants.FILTER_KEY].(string)
			valueToSearchA := a[global_constants.FILTER_VALUE].(string)

			keyToSearchB := b[global_constants.FILTER_KEY].(string)
			valueToSearchB := b[global_constants.FILTER_VALUE].(string)

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

		keyToSearch := eachIndexMap[global_constants.FILTER_KEY].(string)
		valueToSearch := eachIndexMap[global_constants.FILTER_VALUE].(string)

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

func (collection *Collection) isDocumentExists(id string) (bool, string, Document) {
	var document Document
	var batchId string

	for eachBatchId, documents := range collection.DocumentsMap {

		if value, exists := documents[id]; exists {
			document = value
			batchId = eachBatchId
		}
	}

	if _, exists := document[global_constants.DOC_ID]; !exists {
		return false, batchId, document
	}

	return true, batchId, document
}

type SortByDocIndex []Document

func (a SortByDocIndex) Len() int      { return len(a) }
func (a SortByDocIndex) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a SortByDocIndex) Less(i, j int) bool {

	iDocIndex := a[i][global_constants.DOC_INDEX].(int)
	jDocIndex := a[j][global_constants.DOC_INDEX].(int)

	return iDocIndex < jDocIndex
}

func sortDocuments(documents []Document) {
	sort.Sort(SortByDocIndex(documents))
}
