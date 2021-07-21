package filters

import (
	"log"
	"math/rand"
	"sync"

	"drhyu.com/indexer/models"
)

type SearchGroup struct {
	Filters          []Filter
	SearchMode       int
	searchGroupIndex int

	nSearchGroups int
}

const (
	SEARCH_MODE_OR  = iota
	SEARCH_MODE_AND = iota
)

/**
* Apply all filters in the SearchGroup to an item
* If the item passes the filters:
* 		If it has passed through ALL other SearchGroups in a Search it is passed to outMatchedItems
*		If it has NOT passed through all SearchGroups it is forwarded to the next SearchGroup through outMatchedItems
* If the item doesn't pass the filters it is discarded an not forwarded
 */
func (searchGroup *SearchGroup) ApplyTo(
	newItem *TrackedItem,
	outForwardedItems chan *TrackedItem,
	outMatchedItems chan *models.Item) error {

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
			select {
			case outMatchedItems <- newItem.item:
			default:
				log.Printf("[Error] SG %d: outMatchedItems (%d/%d) is full", searchGroup.searchGroupIndex, len(outMatchedItems), cap(outMatchedItems))
				outMatchedItems <- newItem.item
			}
		} else {
			// Still needs to go though other SearchGroups
			select {
			case outForwardedItems <- newItem:
			default:
				log.Printf("[Error] SG %d: outForwardedItems (%d/%d) is full\n", searchGroup.searchGroupIndex, len(outForwardedItems), cap(outForwardedItems))
				outForwardedItems <- newItem
			}
		}
	}
	// else {
	// Item is not passed anywhere so it is essentially discarded
	//}

	return nil
}

/**
* Process incoming items
* They can be taken from inForwardedItems or inNewItems
* inForwardedItems are items already in the "system" passed from SearchGroup to SearchGroup
* inNewItems are new items fresh from the tradeAPI
* Processing "old" items takes priority before fetching new ones
 */
func (searchGroup *SearchGroup) StartSearchGroup(
	inNewItems chan models.Item, // New items passed from the fetchAPI
	inForwardedItems chan *TrackedItem, // Items passed from another SearchGroup
	outForwardedItems chan *TrackedItem,
	outMatchedItems chan *models.Item,
	exitSignal chan bool,
	failureSignal chan error,
	wg *sync.WaitGroup) {

	defer wg.Done()

	for {

		// Mechanism to give items in inForwardedItems preference vs items in inNewItems
		select {
		case <-exitSignal:
			return
		case trackedItem := <-inForwardedItems:
			if err := searchGroup.ApplyTo(trackedItem, outForwardedItems, outMatchedItems); err != nil {
				failureSignal <- err
				return
			}
		default:
			// Makes it non-blocking
		}

		select {
		case <-exitSignal:
			return
		case trackedItem := <-inForwardedItems:
			if err := searchGroup.ApplyTo(trackedItem, outForwardedItems, outMatchedItems); err != nil {
				failureSignal <- err
				return
			}
		case newItem := <-inNewItems:
			// Item passed in from trade API is of type models.Item not TrackedItem
			//(which includes info on which SearchGroup has already processed this item)
			trackedItem := &TrackedItem{track: make([]bool, searchGroup.nSearchGroups), item: &newItem, id: rand.Int31()}
			if err := searchGroup.ApplyTo(trackedItem, outForwardedItems, outMatchedItems); err != nil {
				failureSignal <- err
				return
			}
		}
	}
}
