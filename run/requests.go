package main

import (
	"log"
	"os"
	"time"

	gorequests "github.com/gauravsarma1992/gorequests"
)

func main() {
	var (
		req           *gorequests.ApiStore
		stats         interface{}
		stages        []string
		timeIntervals []int
		err           error
	)

	stages = []string{"firstStage", "secondStage", "thirdStage"}
	timeIntervals = []int{10, 5, 8}

	if req, err = gorequests.NewApiStore(); err != nil {
		log.Println(err)
		os.Exit(-1)
	}

	for idx, stageName := range stages {
		go req.Run(stageName)
		time.Sleep(time.Duration(timeIntervals[idx]) * time.Second)

		if err = req.Close(); err != nil {
			log.Println(err)
			os.Exit(-1)
		}
		if stats, err = req.FlushStats(); err != nil {
			log.Println(err)
			os.Exit(-1)
		}
		log.Println(stats)
	}
}
