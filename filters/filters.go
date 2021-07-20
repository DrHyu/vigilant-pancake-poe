package filters

import (
	"fmt"
	"log"
	"math/rand"
	"reflect"

	"drhyu.com/indexer/models"
)

const (
	COMP_INT_EQ    = iota
	COMP_INT_GT    = iota
	COMP_INT_LT    = iota
	COMP_INT_GTE   = iota
	COMP_INT_LTE   = iota
	COMP_STR_REGEX = iota
	COMP_STR_EQ    = iota
)
const (
	SEARCH_MODE_OR  = iota
	SEARCH_MODE_AND = iota
)

type Search struct {
	SearchGroups []SearchGroup
}

type SearchGroup struct {
	Filters          []Filter
	SearchMode       int
	searchGroupIndex int

	nSearchGroups int
}

type Filter struct {
	Field string
	Value string

	ComparisonMethod int
	InverseMatch     bool
}

type TrackedItem struct {
	item *models.Item

	id int32

	// 1 item x number of SearchGroups in a Search
	// At the end of the search process, if it is all true the item is a match
	track []bool
}

func getAttr(obj interface{}, fieldName string) reflect.Value {
	pointToStruct := reflect.ValueOf(obj) // addressable
	curStruct := pointToStruct.Elem()
	if curStruct.Kind() != reflect.Struct {
		panic("not struct")
	}
	curField := curStruct.FieldByName(fieldName) // type: reflect.Value
	if !curField.IsValid() {
		panic("not found:" + fieldName)
	}
	return curField
}

func (filter *Filter) ApplyTo(item *models.Item) (bool, error) {
	return true, nil
}

// Apply all filters in the SearchGroup to an item
// If the item passes the filters:
// 		If it has passed through ALL other SearchGroups in a Search it is passed to outMatchedItems
//		If it has NOT passed through all SearchGroups it is forwarded to the next SearchGroup through outMatchedItems
// If the item doesn't pass the filters it is discarded an not forwarded
func (searchGroup *SearchGroup) ApplyTo(
	newItem *TrackedItem,
	outForwardedItems chan<- *TrackedItem,
	outMatchedItems chan<- *models.Item) error {

	log.Printf("SG %d got item %v %v", searchGroup.searchGroupIndex, newItem.track, newItem.id)

	var pass bool

	if searchGroup.SearchMode == SEARCH_MODE_AND {
		// Assume it passes, if one filter fails => failed
		pass = true
	} else {
		// Assume it fails, if one filter succeeds => success
		pass = false
	}

	// Pass this item through all Filters in this searchGroup
	for i := range searchGroup.Filters {

		success, err := searchGroup.Filters[i].ApplyTo(newItem.item)

		if err != nil {
			log.Print("ERROR: ", err)
			return err
		}

		if searchGroup.SearchMode == SEARCH_MODE_AND && !success {
			// All filters must match in AND mode
			// Stop processing and don't forward item
			pass = false
		}
		if searchGroup.SearchMode == SEARCH_MODE_OR && success {
			// Only one filter is required to match in OR mode
			// Stop processing and forward item to other SearchGroups (if any)
			pass = true
		}
	}

	if pass {
		// Mark the item as having passed this SearchGroup
		newItem.track[searchGroup.searchGroupIndex] = true

		// Check if this item has already been processed by all other searchgroups
		allPassed := true
		for _, p := range newItem.track {
			allPassed = allPassed && p
		}

		if allPassed {
			// Item match !
			// log.Printf("SG %d -> outMatchedItems TRY item %v %v - %v", searchGroup.searchGroupIndex, newItem.track, newItem.id, outMatchedItems)
			select {
			case outMatchedItems <- newItem.item:
				log.Printf("SG %d: put in outMatchedItems", searchGroup.searchGroupIndex)
			default:
				log.Printf("SG %d: outMatchedItems is full", searchGroup.searchGroupIndex)
			}
			// log.Printf("SG %d -> outMatchedItems DONE item %v %v - %v", searchGroup.searchGroupIndex, newItem.track, newItem.id, outMatchedItems)
		} else {
			select {
			case outForwardedItems <- newItem:
				log.Printf("SG %d: put in outForwardedItems", searchGroup.searchGroupIndex)
			default:
				log.Printf("SG %d: outForwardedItems is full", searchGroup.searchGroupIndex)
			}
			// Still needs to go though other SearchGroups
			// log.Printf("SG %d -> outForwardedItems TRY item %v %v - %v", searchGroup.searchGroupIndex, newItem.track, newItem.id, outForwardedItems)

			// log.Printf("SG %d -> outForwardedItems DONE item %v %v - %v", searchGroup.searchGroupIndex, newItem.track, newItem.id, outForwardedItems)
		}
	}
	// else {
	// Item is not passed anywhere so it is essentially discarded
	//}

	return nil
}

// Process incoming items
// They can be taken from inForwardedItems or inNewItems
// inForwardedItems are items already in the "system" passed from SearchGroup to SearchGroup
// inNewItems are new items fresh from the tradeAPI
// Processing "old" items takes priority before fetching new ones
func (searchGroup *SearchGroup) StartSearchGroup(
	inNewItems <-chan *models.Item, // New items passed from the fetchAPI
	inForwardedItems <-chan *TrackedItem, // Items passed from another SearchGroup
	outForwardedItems chan<- *TrackedItem,
	outMatchedItems chan<- *models.Item,
	exitSignal chan bool,
	failureSignal chan error) {

	for {
		var trackedItem *TrackedItem = nil

		// Mechanism to give items in inForwardedItems preference vs items in inNewItems
		select {
		case <-exitSignal:
			return
		case trackedItem = <-inForwardedItems:
			// log.Printf("Fetch from froward %d id: %d ", searchGroup.searchGroupIndex, trackedItem.id)
			searchGroup.ApplyTo(trackedItem, outForwardedItems, outMatchedItems)
			continue
		default:
			// Makes it non-blocking
		}

		select {
		case <-exitSignal:
			return
		case trackedItem = <-inForwardedItems:
			// log.Printf("Fetch from froward %d id: %d ", searchGroup.searchGroupIndex, trackedItem.id)
			searchGroup.ApplyTo(trackedItem, outForwardedItems, outMatchedItems)
		case item := <-inNewItems:
			// Item passed in from trade API is of type models.Item not TrackedItem
			//(which includes info on which SearchGroup has already processed this item)
			trackedItem = &TrackedItem{track: make([]bool, searchGroup.nSearchGroups), item: item, id: rand.Int31()}
			// log.Printf("Fetch from new %d id: %d ", searchGroup.searchGroupIndex, trackedItem.id)
			searchGroup.ApplyTo(trackedItem, outForwardedItems, outMatchedItems)
		}
	}
}

func (search *Search) StartSearch(
	itemsIn <-chan *models.Item,
	itemsOut chan<- *models.Item,
	exitSignal chan bool,
	failureSignal chan error) {

	childExitSignal := make(chan bool)
	childFailureSignal := make(chan error)

	// Make the inter-searchgroup channels to pass items
	channels := make([]chan *TrackedItem, len(search.SearchGroups))

	for i := range channels {
		channels[i] = make(chan *TrackedItem, 50)
	}

	for i := range search.SearchGroups {
		// Assign an index to each SearchGroup
		search.SearchGroups[i].searchGroupIndex = i
		search.SearchGroups[i].nSearchGroups = len(search.SearchGroups)

		// Launch each SearchGroup in parallel
		// eg:
		// SG 1 -> SG 2
		// SG 2 -> SG 3
		// ...
		// SG N -> SG 0
		go search.SearchGroups[i].StartSearchGroup(
			itemsIn,
			channels[i],
			channels[(i+1)%(len(channels)-1)],
			itemsOut,
			childExitSignal,
			childFailureSignal)
	}

	// Wait for kill signal or error
	select {
	// Signal to exit
	case <-exitSignal:
		for range search.SearchGroups {
			childExitSignal <- true
		}
		return
		// TODO
	case err := <-childFailureSignal:
		failureSignal <- err
		for range search.SearchGroups {
			childExitSignal <- true
		}
		return
	}

}

func Test() {
	f := Filter{Field: "test", Value: "23"}

	fmt.Print(getAttr(&f, "Field"))
}
