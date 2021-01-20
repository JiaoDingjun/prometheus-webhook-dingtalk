package notifier

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type ChainMonitor struct {
	ServerEndpoint string
}

const (
	block_height_pql = `block_height{chainID="%s"}`
)

func NewChainMonitor(svr string) *ChainMonitor {
	return &ChainMonitor{ServerEndpoint: svr}
}

func (o *ChainMonitor) genPQL(query string, t int64) string {
	v := url.Values{}
	v.Add("query", query)
	v.Add("time", fmt.Sprintf("%d", t))
	return v.Encode()
}

func (o *ChainMonitor) genRangePQL(query string, start, end int64, step string) string {
	v := url.Values{}
	v.Add("query", query)
	v.Add("start", fmt.Sprintf("%d", start))
	v.Add("end", fmt.Sprintf("%d", end))
	v.Add("step", step)
	return v.Encode()
}

type RangePQLResp struct {
	Status    string                 `json:"status"`
	Data      map[string]interface{} `json:"data"`
	ErrorType string                 `json:"errorType"`
	ErrMsg    string                 `json:"error"`
}

func (o *ChainMonitor) query(q string, isRange bool) ([]byte, error) {
	qm := "query"
	if isRange {
		qm = "query_range"
	}
	u := fmt.Sprintf("%s/api/v1/%s", o.ServerEndpoint, qm)

    fmt.Printf("ChainMonitor,query,url:%s\n",u)

	resp, err := http.Post(u,
		"application/x-www-form-urlencoded",
		strings.NewReader(q))
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	fmt.Printf("chain monitor query, err: %v,url:%s, pql:%s, resp:%s\n", err, u, q, string(body))

	res := &RangePQLResp{}
	err = json.Unmarshal(body, &res)
	if err != nil {
		return nil, err
	}

	if res.Status != "success" {
		return nil, fmt.Errorf("status:%s,errorType:%s,errMsg:%s", res.Status, res.ErrorType, res.ErrMsg)
	}

	return body, nil
}

// GetChainBlockHeight 取得给节点的最近的区块高度
func (o *ChainMonitor) GetChainBlockHeight(chainID string) (string, error) {
	query := fmt.Sprintf(block_height_pql, chainID)
	pql := o.genPQL(query, time.Now().Unix())
	resp, err := o.query(pql, false)
	if err != nil {
		return "", err
	}

	result := &RangePQLResp{}
	err = json.Unmarshal(resp, &result)
	if err != nil {
        fmt.Printf("json.Unmarshal failed,err:%v\n",err)
		return "", err
	}

	resultDataResult, ok := result.Data["result"].([]interface{})
	if !ok {
		return "", nil
	}

	info := ""

	for _, v := range resultDataResult {
		item, ok := v.(map[string]interface{})
		if !ok {
			return "", nil
		}

		metric, ok := item["metric"].(map[string]interface{})
		if !ok {
			return "", nil
		}
		value, ok := item["value"].([]interface{})
		if !ok {
			return "", nil
		}

        itemInfo := fmt.Sprintf("\npeer:%s(nodeID:%s): %s;", metric["kubernetes_name"], metric["nodeID"], value[1].(string))
		info = info + itemInfo
	}

	return info, nil
}
