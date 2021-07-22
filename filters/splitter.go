package filters

import (
	"fmt"
	"log"

	"drhyu.com/indexer/models"
)

// 						API FETCHER
//							  |
//						   Spliter		-> DUMP (if no S)
//					 _________|_________
//					 /	|	|	|	|  \
//					/	|	|	|	|	\
//					|	|	|	|	|	|
//					S1	S2	S3	S4	S5	SN

type Splitter struct {
	inItems chan models.Item

	newSearch chan *Search
	endSearch chan *Search

	stop chan bool
}

func NewSplitter(inputItems chan models.Item) *Splitter {
	return &Splitter{
		inItems:   inputItems,
		newSearch: make(chan *Search),
		endSearch: make(chan *Search),
		stop:      make(chan bool),
	}
}
func (splitter Splitter) Start(failureSignal chan error) {

	log.Print("[INFO] Splitter started")
	var activeSplits = make(map[uint32]chan models.Item)

	defer func() {
		// Close all channels on exit
		for id := range activeSplits {
			close(activeSplits[id])
		}
	}()

	for {

		select {
		// Handle search addition/removal
		case nsearch := <-splitter.newSearch:
			_, ok := activeSplits[nsearch.ID]
			if ok {
				// error seach is already beeing splitted
				failureSignal <- fmt.Errorf("[Splitter] Search (%d) was already beeing splitted", nsearch.ID)
			}
			if nsearch.input == nil {
				failureSignal <- fmt.Errorf("[Splitter] Search (%d) has a nil input channel", nsearch.ID)
				return
			}
			activeSplits[nsearch.ID] = nsearch.input
		case endsearch := <-splitter.endSearch:
			_, ok := activeSplits[endsearch.ID]
			if !ok {
				failureSignal <- fmt.Errorf("[Splitter] Could not delete Search split (%d), not found", endsearch.ID)
			}
			delete(activeSplits, endsearch.ID)
		case inItem := <-splitter.inItems:
			// If there are no active splits it will essentially just throw away all incoming items
			for id := range activeSplits {
				select {
				// Pass by value
				case activeSplits[id] <- inItem:
				default:
					// log.Printf("[Error] Splitter -> Input channel for Search %d is full! (%d/%d)", id, len(activeSplits[id]), cap(activeSplits[id]))
				}
			}
			// if len(activeSplits) == 0 {
			// 	log.Printf("[INFO] Splitter dumping item %s", inItem.ID)
			// }
		case <-splitter.stop:
			log.Print("[INFO] Splitter stopped")
			return
		}
	}
}

func (splitter Splitter) Stop(newSearch *Search) {
	splitter.stop <- true
}
func (splitter Splitter) NewSearch(newSearch *Search) {
	splitter.newSearch <- newSearch
}

func (splitter Splitter) EndSearch(endSearch *Search) {
	splitter.endSearch <- endSearch
}
