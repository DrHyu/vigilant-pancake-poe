package filters

import (
	"sync"

	"drhyu.com/indexer/models"
)

type Search struct {
	SearchGroups []SearchGroup
}

type TrackedItem struct {
	item *models.Item

	id int32

	// 1 item x number of SearchGroups in a Search
	// At the end of the search process, if it is all true the item is a match
	track []bool
}

func (search *Search) StartSearch(
	itemsIn chan models.Item,
	itemsOut chan *models.Item,
	exitSignal chan bool,
	failureSignal chan error) {

	var wg sync.WaitGroup
	var childExitSignal = make(chan bool)
	var childFailureSignal = make(chan error)

	// Make the inter-searchgroup channels to pass items
	var channels = make([]chan *TrackedItem, len(search.SearchGroups))

	for i := range channels {
		channels[i] = make(chan *TrackedItem, 200)
	}

	for i := range search.SearchGroups {

		wg.Add(1)

		// Assign an index to each SearchGroup
		search.SearchGroups[i].searchGroupIndex = i
		search.SearchGroups[i].nSearchGroups = len(search.SearchGroups)

		// Launch each SearchGroup in parallel
		// eg:
		// SG 1 -> SG 2
		// SG 2 -> SG 3
		// ...
		// SG N -> SG 0
		// This approach WOULD deadlock
		// For a given number of inputs (N_INPUTS) and number of matched items (N_MATCHES)
		//  N_MATCHES approaches N_INPUTS/N_SEARCH_GROUPS
		// Escape are in place to prevent an actual deadlock but matched items will be dropped

		go search.SearchGroups[i].StartSearchGroup(
			itemsIn,
			channels[i],
			channels[(i+1)%len(channels)],
			itemsOut,
			childExitSignal,
			childFailureSignal,
			&wg)
	}

	// go debugChanLvlReporter(channels, itemsIn, itemsOut)

	// Wait for kill signal or error
	for {
		select {
		// Signal to exit
		case <-exitSignal:
			for range search.SearchGroups {
				childExitSignal <- true
			}
			return
		case err := <-childFailureSignal:
			// Propagate error
			failureSignal <- err
			// Shut down all childs
			for range search.SearchGroups {
				select {
				case childExitSignal <- true:
				default:
				}
			}
			// Wait for a clean end
			wg.Wait()
			return
		}
	}
}
