package tradeApi

import (
	"drhyu.com/indexer/models"
)

const LEAGUE string = "Standard"

func (fetcher *ApiFetcher) Init() {
	// Initialize the channels
	fetcher.pendingChangeID = make(chan string, 10)
	fetcher.dataToProcess = make(chan *[]byte, 10)
	fetcher.NewItems = make(chan models.Item, 10_000) // Usually 5000 items are fetched in 1 go
	fetcher.ChageIDsSeen = make(map[string]struct{})
}

// Main loop for this thread
func (fetcher *ApiFetcher) Start(initialFetchID string, exitSignal chan bool, failureSignal chan error) {

	childFailureSignal := make(chan error)
	fetchKillSignal := make(chan bool)
	go fetcher.StartFetcher(fetchKillSignal, childFailureSignal)
	processKillSignal := make(chan bool)
	go fetcher.StartItemProcessing(processKillSignal, childFailureSignal)

	// Give the initial fetch ID so that the fetching process may begin
	fetcher.pendingChangeID <- initialFetchID

	// wait to be killed or error happened
	select {
	case <-exitSignal:
		fetchKillSignal <- true
		processKillSignal <- true
	case childErr := <-childFailureSignal:
		// Kill all children
		fetchKillSignal <- true
		processKillSignal <- true
		// propagate error down
		failureSignal <- childErr
	}

}
