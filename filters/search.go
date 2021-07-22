package filters

import (
	"fmt"
	"log"
	"sync"

	"drhyu.com/indexer/models"
	"github.com/google/uuid"
)

type Search struct {
	SearchGroups []SearchGroup

	input   chan models.Item
	matches chan *models.Item

	ID uint32

	stop chan bool
}

type TrackedItem struct {
	item *models.Item

	id int32

	// 1 item x number of SearchGroups in a Search
	// At the end of the search process, if it is all true the item is a match
	track []bool
}

func NewSearch(searchPrameters []SearchGroup, outMatches chan *models.Item) *Search {
	return &Search{
		SearchGroups: searchPrameters,
		input:        make(chan models.Item, 500),
		matches:      outMatches,
		ID:           uuid.New().ID(),
		stop:         make(chan bool),
	}
}

func (search *Search) Start(failureSignal chan error) {

	log.Printf("[INFO] Search %d started\n", search.ID)
	fmt.Printf("[INFO] Search %d started\n", search.ID)
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
			search.input,
			channels[i],
			channels[(i+1)%len(channels)],
			search.matches,
			childExitSignal,
			childFailureSignal,
			&wg)
	}

	// go debugChanLvlReporter(channels, itemsIn, itemsOut)

	// Wait for kill signal or error
	for {
		select {
		case <-search.stop:
			for range search.SearchGroups {
				childExitSignal <- true
			}
			log.Printf("[INFO] Search %d ended\n", search.ID)
			fmt.Printf("[INFO] Search %d ended\n", search.ID)
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

func (search *Search) Stop() {
	search.stop <- true
}
