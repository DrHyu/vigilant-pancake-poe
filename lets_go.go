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

	os.Remove("logs.txt")
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
						PropertyID:       models.P_BASETYPE,
						Value1:           "Lapis Amulet",
						ComparisonMethod: filters.COMP_STR_EQ,
					},
					// {
					// 	PropertyID:       filters.P_ILVL,
					// 	Value:            75,
					// 	ComparisonMethod: filters.COMP_INT_EQ,
					// },
				},
			},
			{
				SearchMode: filters.SEARCH_MODE_AND,
				Filters: []filters.Filter{
					// {
					// 	PropertyID:       filters.P_BASETYPE,
					// 	Value:            "Lapis Amulet",
					// 	ComparisonMethod: filters.COMP_STR_EQ,
					// },
					{
						PropertyID:       models.P_ILVL,
						Value1:           75,
						ComparisonMethod: filters.COMP_INT_GT,
					},
				},
			},
			{
				SearchMode: filters.SEARCH_MODE_AND,
				Filters: []filters.Filter{
					// {
					// 	PropertyID:       filters.P_BASETYPE,
					// 	Value:            "Lapis Amulet",
					// 	ComparisonMethod: filters.COMP_STR_EQ,
					// },
					{
						PropertyID:       models.P_FRAMETYPE,
						Value1:           3,
						ComparisonMethod: filters.COMP_INT_EQ,
					},
				},
			},
		},
	}
	itemsOut := make(chan *models.Item, 100)

	go mSearch.StartSearch(mFetcher.NewItems, itemsOut, exit, failure)

	for {
		item := <-itemsOut

		fmt.Print(item.Describe(models.P_NAME, models.P_DESCRTEXT))
		fmt.Println("Found Match: ")
		for i := range mSearch.SearchGroups {
			for _, f := range mSearch.SearchGroups[i].Filters {

				label := f.GetItemPropertyName(item)
				prop := f.GetItemProperty(item)
				fmt.Printf("\t%v: %v\n", label, prop)
			}
		}

		log.Printf("Got item %s \n", item.ID)
	}

	time.Sleep(time.Second * 40)

}
func main() {

	doTheThing()

}
