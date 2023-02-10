package gorequests

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"sync"
	"time"
)

const (
	DefaultApiConfigFile = "config/requests.json"
)

type (
	ApiStore struct {
		Config   *ApiConfig `json:"config"`
		Requests []*Request `json:"requests"`
		client   *http.Client

		expName   string
		wg        *sync.WaitGroup
		closeCh   chan bool
		numOfReqs int
	}
	ApiConfig struct {
		BaseUrl       string            `json:"base_url"`
		StatsFolder   string            `json:"stats_folder"`
		AuthMechanism string            `json:"auth_mechanism"`
		ApiKey        string            `json:"api_key"`
		Cookies       map[string]string `json:"cookies"`
	}
	Request struct {
		Api         string       `json:"api"`
		Method      string       `json:"method"`
		Payload     interface{}  `json:"payload"`
		QueryParams []QueryParam `json:"query_params"`
		Stats       Stats        `json:"-"`
		RawStats    []string     `json:"-"`
	}
	ApiResp struct {
		Status int
		Body   []byte
	}
)

func NewApiStore() (ast *ApiStore, err error) {
	ast = &ApiStore{
		client:  &http.Client{},
		closeCh: make(chan bool),
		wg:      &sync.WaitGroup{},
	}
	if err = ast.readConfig(); err != nil {
		return
	}
	return
}

func (ast *ApiStore) readConfig() (err error) {
	var (
		fileB []byte
	)
	if fileB, err = ioutil.ReadFile(DefaultApiConfigFile); err != nil {
		return
	}
	if err = json.Unmarshal(fileB, ast); err != nil {
		return
	}
	return
}

func (ast *ApiStore) setHeaders(req *http.Request) (err error) {
	var (
		cookieHeader string
	)
	for cookieKey, cookieVal := range ast.Config.Cookies {
		if cookieHeader != "" {
			cookieHeader += "; "
		}
		cookieHeader += fmt.Sprintf("%s=%s", cookieKey, cookieVal)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", ast.Config.ApiKey))
	req.Header.Set("Cookie", cookieHeader)

	return
}

func (ast *ApiStore) preparePayload(payload interface{}) (reqPayload io.Reader, err error) {
	var (
		payloadB []byte
	)
	if payload != nil {
		if payloadB, err = json.Marshal(&payload); err != nil {
			log.Println(err)
			return
		}
		reqPayload = bytes.NewBuffer(payloadB)
	}
	return
}

func (req *Request) GetName() (name string) {
	name = fmt.Sprintf("%s-%s", req.Method, req.Api)
	return
}

func (ast *ApiStore) GetUrl(req *Request) (uri string) {
	var (
		queryParamsStr string
		err            error
	)
	uri = fmt.Sprintf("%s/%s", ast.Config.BaseUrl, req.Api)
	if queryParamsStr, err = ast.FormQueryParams(req); err != nil {
		log.Println(err)
		return
	}
	uri += queryParamsStr
	return
}

func (ast *ApiStore) handleApiErrors(apiReq *Request, resp *http.Response) (apiResp *ApiResp, err error) {
	apiResp = &ApiResp{
		Status: resp.StatusCode,
	}
	if apiResp.Body, err = ioutil.ReadAll(resp.Body); err != nil {
		return
	}
	if apiResp.Status == http.StatusOK {
		return
	}
	// Entered Error Path
	err = errors.New(fmt.Sprintf(
		"Errored in API %s, Status Code - %d, Method - %s, Body - %s",
		apiReq.Api,
		apiResp.Status,
		apiReq.Method,
		string(apiResp.Body),
	))
	return
}

func (ast *ApiStore) fireApi(apiReq *Request) (apiResp *ApiResp, err error) {
	var (
		req        *http.Request
		resp       *http.Response
		reqPayload io.Reader
		startTime  time.Time

		latencyMs uint64
	)

	ast.wg.Add(1)
	ast.numOfReqs += 1
	defer func() {
		ast.wg.Done()
	}()

	startTime = time.Now()

	if reqPayload, err = ast.preparePayload(apiReq.Payload); err != nil {
		return
	}
	if req, err = http.NewRequest(apiReq.Method, ast.GetUrl(apiReq), reqPayload); err != nil {
		return
	}
	if err = ast.setHeaders(req); err != nil {
		return
	}
	if resp, err = ast.client.Do(req); err != nil {
		return
	}
	defer resp.Body.Close()
	latencyMs = uint64(time.Now().Sub(startTime).Milliseconds())
	apiResp, err = ast.handleApiErrors(apiReq, resp)
	ast.trackStats(apiReq, apiResp, latencyMs)
	return
}

func (ast *ApiStore) Close() (err error) {
	ast.closeCh <- true
	return
}

func (ast *ApiStore) waitForPrevIteration() (err error) {
	if ast.wg == nil {
		ast.wg = &sync.WaitGroup{}
		return
	}
	ast.wg.Wait()
	ast.wg = &sync.WaitGroup{}
	return
}

func (ast *ApiStore) runForRequests() (err error) {
	for _, req := range ast.Requests {
		go ast.fireApi(req)
	}
	return
}

func (ast *ApiStore) Run(expName string) (err error) {
	ast.waitForPrevIteration()
	ast.expName = expName
	ticker := time.NewTicker(2 * time.Second)
	for {
		select {
		case <-ast.closeCh:
			return
		case <-ticker.C:
			if err = ast.runForRequests(); err != nil {
				return
			}
		}
	}
	return
}
