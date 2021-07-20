package filters

import (
	"log"
	"math/rand"
	"regexp"
	"strconv"

	"drhyu.com/indexer/models"
)

const (
	COMP_INT_EQ             = iota
	COMP_INT_GT             = iota
	COMP_INT_LT             = iota
	COMP_INT_GTE            = iota
	COMP_INT_LTE            = iota
	COMP_INT_BETWEEN        = iota
	COMP_STR_EQ             = iota
	COMP_REGEX_MATCH        = iota // Check if regex matches the target field
	COMP_REGEX_FIND_COMPARE = iota // Find a match in a string and compare the match vs another number. Regex -> Value1, CMP 1 -> Value2, CMP 2 -> Value3
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
	Value1 interface{}
	Value2 interface{}
	Value3 interface{}

	PropertyID  int
	SubProperty string

	// How to compare the match of the regex
	RegexMatchComparisonMethod int

	ComparisonMethod int
	InverseMatch     bool
}

func (filter *Filter) GetItemProperty(item *models.Item) interface{} {
	return models.GetItemProperty(filter.PropertyID, filter.SubProperty, item)
}

func (filter *Filter) GetItemPropertyName(item *models.Item) string {
	return models.GetItemPropertyName(filter.PropertyID, filter.SubProperty, item)
}

type TrackedItem struct {
	item *models.Item

	id int32

	// 1 item x number of SearchGroups in a Search
	// At the end of the search process, if it is all true the item is a match
	track []bool
}

type FilterError struct{}

func (f *FilterError) Error() string {
	return "Filter error !"
}

func (filter *Filter) ApplyTo(item *models.Item) (bool, error) {

	var result bool
	value := filter.GetItemProperty(item)

	if value == nil {
		return false, nil
	}

	switch filter.ComparisonMethod {

	case COMP_STR_EQ:
		result = value.(string) == filter.Value1.(string)
	case COMP_INT_EQ:
		result = value.(int) == filter.Value1.(int)
	case COMP_INT_GT:
		result = value.(int) > filter.Value1.(int)
	case COMP_INT_GTE:
		result = value.(int) >= filter.Value1.(int)
	case COMP_INT_LT:
		result = value.(int) < filter.Value1.(int)
	case COMP_INT_LTE:
		result = value.(int) <= filter.Value1.(int)
	case COMP_INT_BETWEEN:
		result = value.(int) >= filter.Value1.(int) && value.(int) <= filter.Value2.(int)
	case COMP_REGEX_MATCH:
		match, err := regexp.Match(filter.Value1.(string), []byte(value.(string)))
		if err != nil {
			return false, err
		}
		result = match
	case COMP_REGEX_FIND_COMPARE:
		r, err := regexp.Compile(filter.Value1.(string))
		if err != nil {
			return false, err
		}
		// Take underlying string in interface{} and cast it to byte[] for regex
		match := r.Find([]byte(value.(string)))
		// Regex didn't match
		if match == nil {
			result = false
			break
		}

		foundInt, err := strconv.Atoi(string(match))
		if err != nil {
			return false, err
		}

		switch filter.RegexMatchComparisonMethod {
		case COMP_STR_EQ:
			return false, &FilterError{}
		case COMP_INT_EQ:
			result = foundInt == filter.Value2.(int)
		case COMP_INT_GT:
			result = foundInt > filter.Value2.(int)
		case COMP_INT_GTE:
			result = foundInt >= filter.Value2.(int)
		case COMP_INT_LT:
			result = foundInt < filter.Value2.(int)
		case COMP_INT_LTE:
			result = foundInt <= filter.Value2.(int)
		case COMP_INT_BETWEEN:
			result = foundInt >= filter.Value2.(int) && value.(int) <= filter.Value3.(int)
		default:
			return false, nil
		}

	default:
		return false, nil
	}

	if filter.InverseMatch {
		return !result, nil
	} else {
		return result, nil
	}
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
			log.Print("[Error]: ", err)
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
				log.Printf("[Error] SG %d: outMatchedItems is full", searchGroup.searchGroupIndex)
			}
		} else {
			// Still needs to go though other SearchGroups
			select {
			case outForwardedItems <- newItem:
			default:
				log.Printf("[Error] SG %d: outForwardedItems is full", searchGroup.searchGroupIndex)
			}
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
	inNewItems <-chan models.Item, // New items passed from the fetchAPI
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
			searchGroup.ApplyTo(trackedItem, outForwardedItems, outMatchedItems)
			continue
		default:
			// Makes it non-blocking
		}

		select {
		case <-exitSignal:
			return
		case trackedItem = <-inForwardedItems:
			searchGroup.ApplyTo(trackedItem, outForwardedItems, outMatchedItems)
		case newItem := <-inNewItems:
			// Item passed in from trade API is of type models.Item not TrackedItem
			//(which includes info on which SearchGroup has already processed this item)
			trackedItem = &TrackedItem{track: make([]bool, searchGroup.nSearchGroups), item: &newItem, id: rand.Int31()}
			searchGroup.ApplyTo(trackedItem, outForwardedItems, outMatchedItems)
		}
	}
}

func (search *Search) StartSearch(
	itemsIn <-chan models.Item,
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

	// if len(search.SearchGroups) > 1 {

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
			childFailureSignal)
	}
	// } else if len(search.SearchGroups) == 1 {
	// 	search.SearchGroups[0].searchGroupIndex = 0
	// 	search.SearchGroups[0].nSearchGroups = 1
	// 	go search.SearchGroups[0].StartSearchGroup(
	// 		itemsIn,
	// 		nil,
	// 		nil,
	// 		itemsOut,
	// 		childExitSignal,
	// 		childFailureSignal)
	// }

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

}
