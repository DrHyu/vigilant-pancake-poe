package tradeApi

import (
	"log"

	"drhyu.com/indexer/models"
)

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

		itemCount := 0
		stashCount := 0
		for i := range result.Stashes {
			if result.Stashes[i].League == LEAGUE {
				itemCount += len(result.Stashes[i].Items)
				stashCount++
			}
		}
		log.Printf("[INFO] Processing %d stashes %d items in %s\n", stashCount, itemCount, LEAGUE)

		fetcher.ChageIDsSeen[result.NextChangeID] = struct{}{}
		for i := range result.Stashes {
			if result.Stashes[i].League == LEAGUE {
				for it := range result.Stashes[i].Items {
					select {
					case fetcher.NewItems <- result.Stashes[i].Items[it]:
						continue
					default:
						log.Printf("[ERROR] Channel full (%d/%d). Fetcher is waiting\n", len(fetcher.NewItems), cap(fetcher.NewItems))
						// Do it anyways
						fetcher.NewItems <- result.Stashes[i].Items[it]
					}
				}
			}
		}
	}

	return result, nil
}

// Take the JSON response and put individual fetched items in the newItemsChannel
func (fetcher *ApiFetcher) StartItemProcessing(exitSignal chan bool, failureSignal chan error) {
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
			}
		}
	}
}
