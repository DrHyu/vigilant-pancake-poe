package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"drhyu.com/indexer/filters"
	"drhyu.com/indexer/models"
	"drhyu.com/indexer/tradeApi"
)

func setupLogs() {

	file, err := os.OpenFile("logs.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatal(err)
	}

	log.SetFlags(log.LstdFlags | log.Lmicroseconds)
	log.SetOutput(file)
}

func doTheThing() {

	setupLogs()
	exit := make(chan bool)
	failure := make(chan error)

	mFetcher := &tradeApi.ApiFetcher{}
	mFetcher.Init()
	go mFetcher.Start("1224902497-1229569941-1187442478-1328449025-1276721127", exit, failure)

	mSearch := &filters.Search{
		SearchGroups: []filters.SearchGroup{
			{
				SearchMode: filters.SEARCH_MODE_AND,
				Filters: []filters.Filter{
					{
						Field:            "test",
						Value:            "test",
						ComparisonMethod: filters.COMP_STR_EQ,
					},
					{
						Field:            "test2",
						Value:            "test3",
						ComparisonMethod: filters.COMP_STR_EQ,
					},
				},
			},
			{
				SearchMode: filters.SEARCH_MODE_AND,
				Filters: []filters.Filter{
					{
						Field:            "test",
						Value:            "test",
						ComparisonMethod: filters.COMP_STR_EQ,
					},
					{
						Field:            "test2",
						Value:            "test3",
						ComparisonMethod: filters.COMP_STR_EQ,
					},
				},
			},
		},
	}
	itemsOut := make(chan *models.Item, 100)

	go mSearch.StartSearch(mFetcher.NewItems, itemsOut, exit, failure)

	// _ = mSearch
	// for {
	// 	select {
	// 	case item := <-mFetcher.NewItems:
	// 		// case <-mFetcher.NewItems:
	// 		fmt.Print(item.BaseType)
	// 		continue

	// 	case childErr := <-failure:
	// 		_ = childErr
	// 		panic("Something bad happened ")
	// 	}
	// }

	for {
		item := <-itemsOut

		fmt.Printf("Got item %s \n", item.BaseType)

	}

	time.Sleep(time.Second * 40)

}
func main() {

	doTheThing()

}
