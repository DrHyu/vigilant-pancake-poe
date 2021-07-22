package filters

import (
	"fmt"
	"log"

	"drhyu.com/indexer/models"
)

type SearchManager struct {
	searches map[uint32]*Search

	inItems  chan models.Item
	splitter *Splitter

	newSearch chan *Search
	endSearch chan uint32

	stop chan bool
}

func NewSearchManager(in chan models.Item) *SearchManager {

	return &SearchManager{
		splitter:  NewSplitter(in),
		inItems:   in,
		searches:  make(map[uint32]*Search),
		stop:      make(chan bool),
		newSearch: make(chan *Search),
		endSearch: make(chan uint32),
	}
}

func (searchManager *SearchManager) Start(failureSignal chan error) {

	log.Print("[INFO] SearchManager started")
	var childFailureSignal = make(chan error)

	go searchManager.splitter.Start(childFailureSignal)

	for {
		select {
		case newSearch := <-searchManager.newSearch:
			searchManager.handleNewSearch(newSearch, childFailureSignal)
		case endSearchId := <-searchManager.endSearch:
			searchManager.handleEndSearch(endSearchId)
		case <-searchManager.stop:
			log.Print("[INFO] SearchManager stopping")
			// Stop all child searches
			for id := range searchManager.searches {
				searchManager.handleEndSearch(id)
			}
			log.Print("[INFO] SearchManager stopped")
			return
		case err := <-childFailureSignal:
			fmt.Println("error out")
			failureSignal <- err
			return
		}
	}

}
func (searchManager *SearchManager) Stop() {
	searchManager.stop <- true
}

func (searchManager *SearchManager) handleNewSearch(newSearch *Search, childFailureSignal chan error) {

	if _, ok := searchManager.searches[newSearch.ID]; ok {
		//Already running TODO error or no error ?
		return
	}
	// Register this search in the splitter so trafic is also splitted towards it
	searchManager.splitter.NewSearch(newSearch)
	searchManager.searches[newSearch.ID] = newSearch
	// Launch the Search
	go searchManager.searches[newSearch.ID].Start(childFailureSignal)
}

func (searchManager *SearchManager) handleEndSearch(endSearchId uint32) error {

	endSearch, ok := searchManager.searches[endSearchId]
	if !ok {
		return fmt.Errorf("[Error] SearchManager could not find %d in registered searches.")
	}
	// De-register this search in the splitter so trafic split is stopped
	searchManager.splitter.EndSearch(endSearch)

	endSearch.Stop()
	return nil
}

func (searchManager *SearchManager) NewSearch(nSearch *Search) {
	searchManager.newSearch <- nSearch
}

func (searchManager *SearchManager) EndSearch(id uint32) bool {
	_, ok := searchManager.searches[id]
	if ok {
		searchManager.endSearch <- id
		return true
	} else {
		return false
	}
}
