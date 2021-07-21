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
	NewItems chan models.Item
}

type NextChangeIDEmptyError struct{}

func (t *NextChangeIDEmptyError) Error() string {
	return "NextChangeID is empty"
}

func (fetcher *ApiFetcher) Fetch(changeID string) (*[]byte, error) {

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

	val := resp.Header["X-Rate-Limit-Ip-State"]
	log.Printf("[INFO] Processing ID: %s %v\n", changeID, val)

	// Get the maximum allowed fetching frequency
	val, ok := resp.Header["X-Rate-Limit-Ip"]

	if !ok {
		log.Println("[ERROR] Failed to get X-Rate-Limit-Ip from response", resp.Header)
		return nil, nil
	}

	temp := strings.Split(val[0], ":")

	n_fetches, _ := strconv.Atoi(temp[0])
	per_n_seconds, _ := strconv.Atoi(temp[1])
	timeout_duration_s, _ := strconv.Atoi(temp[2])

	fetcher.fetchDelayMSec = int64(1000 * float32(per_n_seconds) / float32(n_fetches))
	fetcher.timeoutPolicy = int64(timeout_duration_s)

	if err != nil {
		log.Println("[ERROR] Failed to fetch ID: ", changeID, "\n-- ", err)
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

func (fetcher *ApiFetcher) StartFetcher(exitSignal chan bool, failureSignal chan error) {

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
