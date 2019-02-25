package investing

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

const historyUrl = `https://tvc4.forexpros.com/7faf7c0428fe0cdb621c45d461268e4b/%v/6/6/28/history?symbol=%v&resolution=%v&from=%v&to=%v`

func genHistoryUrl(symbol, typ string, start, end time.Time) string {
	return fmt.Sprintf(historyUrl, time.Now().Unix(), symbol, typ, start.Unix(), end.Unix())
}

func GetHistory(symbol, typ string, start, end time.Time) (*Lines, error) {
	lines := new(Lines)
	url := genHistoryUrl(symbol, typ, start, end)
	resp, err := http.Get(url)
	if err != nil {
		return lines, err
	}
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	err = json.Unmarshal(body, lines)
	if err != nil {
		return lines, err
	}
	if lines.S == "ok" {
		return lines, err
	} else {
		return lines, fmt.Errorf("resp body :%v", string(body))
	}
}

type Lines struct {
	T  []int64       `json:"t"`  // 时间点
	C  []float64     `json:"c"`  // 收盘价
	O  []float64     `json:"o"`  // 开盘价
	L  []float64     `json:"l"`  // 最低
	H  []float64     `json:"h"`  // 最高
	V  []json.Number `json:"v"`  // 成交额
	Vo []json.Number `json:"vo"` // 成交量
	S  string        `json:"s"`
}
