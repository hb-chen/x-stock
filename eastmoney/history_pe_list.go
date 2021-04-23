// 获取历史市盈率

package eastmoney

import (
	"context"
	"errors"
	"sort"
	"strconv"
	"time"

	"github.com/axiaoxin-com/goutils"
	"github.com/axiaoxin-com/logging"
	"go.uber.org/zap"
)

// RespHistoryPE 历史市盈率接口返回结构
type RespHistoryPE struct {
	Data [][]struct {
		Securitycode string `json:"SECURITYCODE"`
		Datetype     string `json:"DATETYPE"`
		Sl           string `json:"SL"`
		Endate       string `json:"ENDATE"`
		Value        string `json:"VALUE"`
	} `json:"data"`
	Pe [][]struct {
		Securitycode string `json:"SECURITYCODE"`
		Pe30         string `json:"PE30"`
		Pe50         string `json:"PE50"`
		Pe70         string `json:"PE70"`
		Total        string `json:"TOTAL"`
		Rn1          string `json:"RN1"`
		Rn2          string `json:"RN2"`
		Rn3          string `json:"RN3"`
	} `json:"pe"`
}

// HistoryPE 历史 pe
type HistoryPE struct {
	Value float64
	Date  string
}

// HistoryPEList 历史 pe 列表
type HistoryPEList []HistoryPE

// GetMidValue 获取历史 pe 中位数
func (h HistoryPEList) GetMidValue() (float64, error) {
	vlen := len(h)
	if vlen == 0 {
		return 0, errors.New("no data")
	}
	values := []float64{}
	for _, i := range h {
		values = append(values, i.Value)
	}
	sort.Float64s(values)
	mid := vlen / 2
	if vlen%2 == 0 {
		return (values[mid-1] + values[mid]) / 2.0, nil
	}
	return values[mid], nil
}

// QueryHistoryPEList 获取历史市盈率
func (e EastMoney) QueryHistoryPEList(ctx context.Context, secuCode string) (HistoryPEList, error) {
	apiurl := "https://emfront.eastmoney.com/APP_HSF10/CPBD/GZFX"
	params := map[string]string{
		"code": e.GetFC(secuCode),
		"year": "4", // 10 年
		"type": "1", // 市盈率
	}
	logging.Debug(ctx, "EastMoney QueryHistoryPE "+apiurl+" begin", zap.Any("params", params))
	beginTime := time.Now()
	apiurl, err := goutils.NewHTTPGetURLWithQueryString(ctx, apiurl, params)
	if err != nil {
		return nil, err
	}
	resp := RespHistoryPE{}
	if err := goutils.HTTPGET(ctx, e.HTTPClient, apiurl, &resp); err != nil {
		return nil, err
	}
	latency := time.Now().Sub(beginTime).Milliseconds()
	logging.Debug(ctx, "EastMoney QueryHistoryPE "+apiurl+" end", zap.Int64("latency(ms)", latency), zap.Any("resp", resp))
	result := HistoryPEList{}
	if len(resp.Data) == 0 {
		return nil, errors.New("no history pe data")
	}
	for _, i := range resp.Data[0] {
		value, err := strconv.ParseFloat(i.Value, 64)
		if err != nil {
			logging.Error(ctx, "HistoryPE ParseFloat error:"+err.Error())
			continue
		}
		pe := HistoryPE{
			Date:  i.Endate,
			Value: value,
		}
		result = append(result, pe)
	}
	return result, nil
}
