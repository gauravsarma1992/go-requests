package main

import (
	"log"
	"os"
	"time"

	gorequests "github.com/gauravsarma1992/gorequests"
)

func main() {
	var (
		req *gorequests.ApiStore
		err error
	)

	if req, err = gorequests.NewApiStore(); err != nil {
		log.Println(err)
		os.Exit(-1)
	}

	go req.Run("continuous")
	time.Sleep(10 * time.Second)

	if err = req.Close(); err != nil {
		log.Println(err)
		os.Exit(-1)
	}
	if _, err = req.FlushStats(); err != nil {
		log.Println(err)
		os.Exit(-1)
	}
}
