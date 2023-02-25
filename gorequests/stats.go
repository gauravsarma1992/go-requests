package gorequests

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

type (
	Stats struct {
		ReqCountDist           uint64
		SuccessfulReqCountDist uint64
		FailedReqCountDist     uint64
		LatencyMsData          []uint64
	}
)

func (ast *ApiStore) PrintStats() (err error) {
	log.Println("Total number of requests fired", ast.numOfReqs)
	for _, req := range ast.Requests {
		statsFmt := fmt.Sprintf("API %s :- Total - %d, Success - %d, Failed - %d",
			req.GetName(),
			req.Stats.ReqCountDist,
			req.Stats.SuccessfulReqCountDist,
			req.Stats.FailedReqCountDist)

		log.Println(statsFmt)
	}
	log.Println("=============")
	return
}

func (ast *ApiStore) trackStats(req *Request, apiResp *ApiResp, latencyMs uint64) (err error) {
	if apiResp.Status == http.StatusOK {
		req.Stats.SuccessfulReqCountDist += 1
	} else {
		req.Stats.FailedReqCountDist += 1
	}
	req.Stats.ReqCountDist += 1
	rawStat := fmt.Sprintf("%s,%s,%s,%d,%d", time.Now().String(), ast.expName, req.GetName(), apiResp.Status, latencyMs)
	req.RawStats = append(req.RawStats, rawStat)
	return
}

func (ast *ApiStore) FlushStats() (reqStats map[string]Stats, err error) {
	var (
		content []string
	)
	ast.waitForPrevIteration()
	ast.PrintStats()
	ast.numOfReqs = 0
	reqStats = make(map[string]Stats)

	content = append(content, "Time,Experiment Name,Request Name,Status Code,Latency\n")

	for _, req := range ast.Requests {
		reqStats[req.GetName()] = req.Stats
		req.Stats = Stats{}
		content = append(content, req.RawStats...)
		req.RawStats = []string{}
	}
	os.MkdirAll(ast.Config.StatsFolder, os.ModePerm)
	fileName := fmt.Sprintf("%s/%s.csv", ast.Config.StatsFolder, ast.expName)
	if err = ioutil.WriteFile(fileName, []byte(strings.Join(content, "\n")), 0755); err != nil {
		return
	}
	return
}
