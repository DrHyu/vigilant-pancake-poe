package tradeApi

import (
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"drhyu.com/indexer/models"
)

const ENDPOINT_URL string = "https://www.pathofexile.com/api/public-stash-tabs?id="

type ApiFetcher struct {
	PrevChangeID string

	//Internal
	// Pass pending IDs to the fetcher
	pendingChangeID chan string
	dataToProcess   chan *[]byte

	lastFetchTime  time.Time
	fetchDelayMSec int64
	timeoutPolicy  int64

	client http.Client

	ChageIDsSeen map[string]struct{}

	// External
	NewItems chan *models.Item
}

type NextChangeIDEmptyError struct{}

func (t *NextChangeIDEmptyError) Error() string {
	return "NextChangeID is empty"
}

func (fetcher *ApiFetcher) Fetch(changeID string) (*[]byte, error) {

	log.Println("Processing ID: ", changeID)
	// Ensure we have a valid Change ID to fetch
	if changeID == "" {
		log.Println("Next ID is empty !")
		return nil, &NextChangeIDEmptyError{}
	}

	fetcher.lastFetchTime = time.Now()
	// Build request
	req, _ := http.NewRequest("GET", ENDPOINT_URL+changeID, nil)
	req.Header.Set("User-Agent", "Mozilla/5.0")
	resp, err := fetcher.client.Do(req)

	// Get the maximum allowed fetching frequency
	val, ok := resp.Header["X-Rate-Limit-Ip"]

	if !ok {
		log.Println("Failed to get X-Rate-Limit-Ip from response", resp.Header)
		return nil, nil
	}

	temp := strings.Split(val[0], ":")

	n_fetches, _ := strconv.Atoi(temp[0])
	per_n_seconds, _ := strconv.Atoi(temp[1])
	timeout_duration_s, _ := strconv.Atoi(temp[2])

	fetcher.fetchDelayMSec = int64(1000 * float32(per_n_seconds) / float32(n_fetches))
	fetcher.timeoutPolicy = int64(timeout_duration_s)

	if err != nil {
		log.Println("Failed to fetch ID: ", changeID, "\n-- ", err)
		return nil, err
	}

	defer resp.Body.Close()
	d, _ := io.ReadAll(resp.Body)
	return &d, nil
}

func (fetcher *ApiFetcher) SleepUntilNextFetch() {
	sleepDuration := fetcher.fetchDelayMSec - time.Since(fetcher.lastFetchTime).Milliseconds()
	time.Sleep(time.Millisecond * time.Duration(sleepDuration+200))
}

func (fetcher *ApiFetcher) EndlessFetch(exitSignal chan bool, failureSignal chan error) {

	for {
		select {
		// Signal to exit
		case <-exitSignal:
			return
		case FetchID := <-fetcher.pendingChangeID:
			data, err := fetcher.Fetch(FetchID)
			if err != nil {
				failureSignal <- err
				return
			} else {
				fetcher.dataToProcess <- data
				fetcher.SleepUntilNextFetch()
			}
		}
	}
}

func (fetcher *ApiFetcher) ProcessItems(data *[]byte) (result *models.RespStruct, err error) {

	// err = json.Unmarshal(data, &result)
	result = &models.RespStruct{}
	result.UnmarshalJSON(*data)

	// Load in the next change ID for fetching
	// fetcher.pendingChangeID <- result.NextChangeID
	fetcher.pendingChangeID <- result.NextChangeID

	// Check if this ID was processed before ... skip it then
	_, seen := fetcher.ChageIDsSeen[result.NextChangeID]

	if !seen {
		fetcher.ChageIDsSeen[result.NextChangeID] = struct{}{}
		for _, stash := range result.Stashes {
			if stash.League == "Ultimatum" {
				for _, item := range stash.Items {
					select {
					case fetcher.NewItems <- &item:
						continue
						// default:
						// 	log.Printf("Channel full. Discarding %v \n", item.BaseType)
					}
				}
			}
		}
	}

	return result, nil
}

// Take the JSON response and put individual fetched items in the newItemsChannel
func (fetcher *ApiFetcher) EndlessProcessItems(exitSignal chan bool, failureSignal chan error) {
	for {
		select {
		// Signal to exit
		case <-exitSignal:
			return
		case data := <-fetcher.dataToProcess:
			_, err := fetcher.ProcessItems(data)
			if err != nil {
				failureSignal <- err
				return
			} else {

			}
		}
	}

}

func (fetcher *ApiFetcher) Init() {
	// Initialize the channels
	fetcher.pendingChangeID = make(chan string, 10)
	fetcher.dataToProcess = make(chan *[]byte, 10)
	fetcher.NewItems = make(chan *models.Item, 500)

	fetcher.ChageIDsSeen = make(map[string]struct{})
}

// Main loop for this thread
func (fetcher *ApiFetcher) Start(initialFetchID string, exitSignal chan bool, failureSignal chan error) {

	childFailureSignal := make(chan error)
	fetchKillSignal := make(chan bool)
	go fetcher.EndlessFetch(fetchKillSignal, childFailureSignal)
	processKillSignal := make(chan bool)
	go fetcher.EndlessProcessItems(processKillSignal, childFailureSignal)

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
