package main

import (
	"fmt"
	"log"
	"os"
	"time"

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
	mFetcher := &tradeApi.ApiFetcher{}
	exit := make(chan bool)
	failure := make(chan error)
	mFetcher.Init()
	go mFetcher.Start("1224902497-1229569941-1187442478-1328449025-1276721127", exit, failure)

	for {
		select {
		case item := <-mFetcher.NewItems:
			// case <-mFetcher.NewItems:
			fmt.Print(item.BaseType)
			continue

		case childErr := <-failure:
			_ = childErr
			panic("Something bad happened ")
		}
	}

	time.Sleep(time.Second * 20)

}
func main() {

	doTheThing()

}
