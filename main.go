package main

import (
	"fmt"
	"log"
	"os"

	"drhyu.com/indexer/filters"
	"drhyu.com/indexer/models"
	"drhyu.com/indexer/tradeApi"
	"drhyu.com/indexer/wsocket"
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
	go mFetcher.Start("1239487625-1243885691-1201564317-1344202353-1291779188", exit, failure)

	mSearchMgr := filters.NewSearchManager(mFetcher.NewItems)
	go mSearchMgr.Start(failure)

	wsocket.StartWS(mSearchMgr)

	for {
	}

}

func dumpMatches(itemsOut chan *models.Item, mSearch *filters.Search) {
	for {
		item := <-itemsOut

		// fmt.Print(item.Describe(models.P_NAME, models.P_NOTE))
		fmt.Printf("Search %d Found Match: \n", mSearch.ID)
		for i := range mSearch.SearchGroups {
			for _, f := range mSearch.SearchGroups[i].Filters {

				label := f.GetFilteredPropertyName(item)
				prop := f.GetFilteredPropertyValue(item)
				fmt.Printf("  %v: %v\n", label, prop)
			}
		}
		// log.Printf("Got item %s \n", item.ID)
	}
}

func main() {

	doTheThing()
	// wsocket.NANI1()
	// testStr := "~price 60 chaos"
	// r, err := regexp.Compile(`~(?:price|b/o) (\d+) chaos`)
	// if err != nil {
	// 	panic(err)
	// }
	// // Take underlying string in interface{} and cast it to byte[] for regex
	// // match := r.Find([]byte(value.(string)))
	// match := r.FindSubmatch([]byte(testStr))
	// fmt.Printf("%q\n", match)

}
