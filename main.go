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
	go mFetcher.Start("1227351031-1231948075-1189827612-1331056003-1279251828", exit, failure)

	mSearchMgr := filters.NewSearchManager(mFetcher.NewItems)
	go mSearchMgr.Start(failure)

	wsocket.StartWS(mSearchMgr)

	for {
	}

	//// triggere by web

	// time.Sleep(time.Second * 1)
	// var mSearchGrps1 = []filters.SearchGroup{
	// 	// {
	// 	// 	SearchMode: filters.SEARCH_MODE_AND,
	// 	// 	Filters: []filters.Filter{
	// 	// 		// {
	// 	// 		// 	PropertyID:       models.P_BASETYPE,
	// 	// 		// 	Value1:           "Lapis Amulet",
	// 	// 		// 	ComparisonMethod: filters.COMP_STR_EQ,
	// 	// 		// },
	// 	// 		// {
	// 	// 		// 	PropertyID:       filters.P_ILVL,
	// 	// 		// 	Value:            75,
	// 	// 		// 	ComparisonMethod: filters.COMP_INT_EQ,
	// 	// 		// },
	// 	// 	},
	// 	// },
	// 	// {
	// 	// 	SearchMode: filters.SEARCH_MODE_AND,
	// 	// 	Filters: []filters.Filter{
	// 	// 		// {
	// 	// 		// 	PropertyID:       filters.P_BASETYPE,
	// 	// 		// 	Value:            "Lapis Amulet",
	// 	// 		// 	ComparisonMethod: filters.COMP_STR_EQ,
	// 	// 		// },
	// 	// 		{
	// 	// 			PropertyID:       models.P_ILVL,
	// 	// 			Value1:           75,
	// 	// 			ComparisonMethod: filters.COMP_INT_GT,
	// 	// 		},
	// 	// 	},
	// 	// },
	// 	{
	// 		SearchMode: filters.SEARCH_MODE_OR,
	// 		Filters: []filters.Filter{
	// 			// {
	// 			// 	PropertyID:       models.P_RARITY,
	// 			// 	Value1:           "UNIQUE",
	// 			// 	ComparisonMethod: filters.COMP_STR_EQ,
	// 			// },
	// 			{
	// 				PropertyID: models.P_NOTE,
	// 				Regex:      regexp.MustCompile(`~(?:price|b/o) (\d+) chaos`),

	// 				Value1:                     66,
	// 				Value2:                     68,
	// 				ComparisonMethod:           filters.COMP_REGEX_FIND_COMPARE,
	// 				RegexMatchComparisonMethod: filters.COMP_INT_BETWEEN,
	// 			},
	// 			{
	// 				PropertyID: models.P_NOTE,
	// 				Regex:      regexp.MustCompile(`~(?:price|b/o) (\d+) chaos`),

	// 				Value1:                     "fladhjsfdhjsfkl", //interface{}
	// 				Value2:                     20,
	// 				ComparisonMethod:           filters.COMP_REGEX_FIND_COMPARE,
	// 				RegexMatchComparisonMethod: filters.COMP_INT_BETWEEN,
	// 			},
	// 		},
	// 	},
	// }

	// var mSearchGrps2 = []filters.SearchGroup{
	// 	{
	// 		SearchMode: filters.SEARCH_MODE_AND,
	// 		Filters: []filters.Filter{
	// 			// {
	// 			// 	PropertyID:       models.P_RARITY,
	// 			// 	Value1:           "UNIQUE",
	// 			// 	ComparisonMethod: filters.COMP_STR_EQ,
	// 			// },
	// 			{
	// 				PropertyID:       models.P_NAME,
	// 				Regex:            regexp.MustCompile("Tabula "),
	// 				ComparisonMethod: filters.COMP_REGEX_MATCH,
	// 			},
	// 		},
	// 	},
	// }

	// itemsOut1 := make(chan *models.Item, 10_000)
	// itemsOut2 := make(chan *models.Item, 10_000)

	// var newSearch1 = filters.NewSearch(mSearchGrps1, itemsOut1)
	// var newSearch2 = filters.NewSearch(mSearchGrps2, itemsOut2)

	// go dumpMatches(itemsOut1, newSearch1)
	// go dumpMatches(itemsOut2, newSearch2)

	// mSearchMgr.NewSearch(newSearch1)

	// time.Sleep(10 * time.Second)

	// mSearchMgr.NewSearch(newSearch2)

	// time.Sleep(20 * time.Second)

	// mSearchMgr.EndSearch(newSearch1.ID)
	// mSearchMgr.EndSearch(newSearch2.ID)

	// mSearchMgr.Stop()

	// go mSearch.StartSearch(mFetcher.NewItems, itemsOut, exit, failure)

	// go stream.NANI(itemsOut)

	// time.Sleep(time.Second * 200)

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
