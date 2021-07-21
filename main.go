package main

import (
	"fmt"
	"log"
	"os"
	"regexp"
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
	// go mFetcher.Start("1224902497-1229569941-1187442478-1328449025-1276721127", exit, failure)
	go mFetcher.Start("1227018323-1231611186-1189493786-1330691408-1278889110", exit, failure)

	mSearch := &filters.Search{
		SearchGroups: []filters.SearchGroup{
			// {
			// 	SearchMode: filters.SEARCH_MODE_AND,
			// 	Filters: []filters.Filter{
			// 		// {
			// 		// 	PropertyID:       models.P_BASETYPE,
			// 		// 	Value1:           "Lapis Amulet",
			// 		// 	ComparisonMethod: filters.COMP_STR_EQ,
			// 		// },
			// 		// {
			// 		// 	PropertyID:       filters.P_ILVL,
			// 		// 	Value:            75,
			// 		// 	ComparisonMethod: filters.COMP_INT_EQ,
			// 		// },
			// 	},
			// },
			// {
			// 	SearchMode: filters.SEARCH_MODE_AND,
			// 	Filters: []filters.Filter{
			// 		// {
			// 		// 	PropertyID:       filters.P_BASETYPE,
			// 		// 	Value:            "Lapis Amulet",
			// 		// 	ComparisonMethod: filters.COMP_STR_EQ,
			// 		// },
			// 		{
			// 			PropertyID:       models.P_ILVL,
			// 			Value1:           75,
			// 			ComparisonMethod: filters.COMP_INT_GT,
			// 		},
			// 	},
			// },
			{
				SearchMode: filters.SEARCH_MODE_OR,
				Filters: []filters.Filter{
					// {
					// 	PropertyID:       models.P_RARITY,
					// 	Value1:           "UNIQUE",
					// 	ComparisonMethod: filters.COMP_STR_EQ,
					// },
					{
						PropertyID: models.P_NOTE,
						Regex:      regexp.MustCompile(`~(?:price|b/o) (\d+) chaos`),

						Value1:                     66,
						Value2:                     68,
						ComparisonMethod:           filters.COMP_REGEX_FIND_COMPARE,
						RegexMatchComparisonMethod: filters.COMP_INT_BETWEEN,
					},
					{
						PropertyID: models.P_NOTE,
						Regex:      regexp.MustCompile(`~(?:price|b/o) (\d+) chaos`),

						Value1:                     "fladhjsfdhjsfkl", //interface{}
						Value2:                     20,
						ComparisonMethod:           filters.COMP_REGEX_FIND_COMPARE,
						RegexMatchComparisonMethod: filters.COMP_INT_BETWEEN,
					},
				},
			},
		},
	}
	itemsOut := make(chan *models.Item, 10_000)

	go mSearch.StartSearch(mFetcher.NewItems, itemsOut, exit, failure)

	for {
		item := <-itemsOut

		// fmt.Print(item.Describe(models.P_NAME, models.P_NOTE))
		fmt.Println("Found Match: ")
		for i := range mSearch.SearchGroups {
			for _, f := range mSearch.SearchGroups[i].Filters {

				label := f.GetFilteredPropertyName(item)
				prop := f.GetFilteredPropertyValue(item)
				fmt.Printf("\t%v: %v\n", label, prop)
			}
		}

		// log.Printf("Got item %s \n", item.ID)
	}

	time.Sleep(time.Second * 40)

}

type test struct {
	abc interface{}
}

func main() {

	doTheThing()
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
