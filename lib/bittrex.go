package lib

import (
	"encoding/json"
	"net/http"
	"net/url"
	"strings"

	. "github.com/bitly/go-simplejson"
)

type Bittrex struct {
	accesskeyid, secretkeyid string
}

func (bt *Bittrex) respErr(js *Json) (interface{}, error) {
	return nil, nil
}

func (bt *Bittrex) ToSymbol(cp *CurrencyPair) string {
	return strings.ToUpper(cp.ToSymbol2("-"))
}

func (bt *Bittrex) NormSymbol(cp *string) string {
	return *cp
}

func (bt *Bittrex) sendReq(method, path string,
	params map[string][]string, sign bool) (int, *Json, error) {
	header := map[string][]string{
		"Content-Type": {`application/x-www-form-urlencoded`},
	}

	req := &http.Request{
		Method: method,
		Header: header,
	}

	req.URL, _ = url.Parse("https://bittrex.com" + path)
	if sign {
		sign_params := map[string][]string{
			"ex_key": {bt.accesskeyid},
		}

		for k, v := range params {
			sign_params[k] = v
		}

		q := req.URL.Query()
		q = sign_params
		data := q.Encode() + "&secret_key=" + bt.secretkeyid
		q.Add("sign", strings.ToUpper(GetMD5Hash(data)))
		req.URL.RawQuery = q.Encode()
	} else {
		q := req.URL.Query()
		q = params
		req.URL.RawQuery = q.Encode()
	}
	return recvResp(req)
}

func (bt *Bittrex) SetKey(access, secret string) {
	bt.accesskeyid = access
	bt.secretkeyid = secret
}

func (bt *Bittrex) GetBalance() (balances []Balance, err error) {
	return
}

func (bt *Bittrex) GetPrice(cp *CurrencyPair) (price Price, err error) {
	return
}

func (bt *Bittrex) GetSymbols() (symbols []string, err error) {
	status, js, err := bt.sendReq("GET", "/api/v1.1/public/getmarkets", nil, false)
	if err != nil {
		return
	}

	respOk := func(js *Json) (interface{}, error) {
		var s []string
		data, _ := js.Get("result").Array()
		for _, d := range data {
			dd := d.(map[string]interface{})
			quote := strings.ToLower(dd["BaseCurrency"].(string))
			base := strings.ToLower(dd["MarketCurrency"].(string))
			s = append(s, base+"_"+quote)
		}
		return s, nil
	}

	s, err := ProcessResp(status, js, respOk, bt.respErr)
	if err == nil {
		symbols = s.([]string)
	}
	return
}

func (bt *Bittrex) GetDepth(cp *CurrencyPair) (depth Depth, err error) {
	params := map[string][]string{
		"market": {bt.ToSymbol(cp)},
		"type":   {"both"},
	}
	status, js, err := bt.sendReq("GET", "/api/v1.1/public/getorderbook", params, false)
	if err != nil {
		return
	}

	respOk := func(js *Json) (interface{}, error) {
		var depth Depth
		asks, _ := js.Get("result").Get("sell").Array()
		for _, a := range asks {
			uu := a.(map[string]interface{})
			price, _ := uu["Rate"].(json.Number).Float64()
			amount, _ := uu["Quantity"].(json.Number).Float64()
			depth.Asks = append([]Unit{Unit{price, amount}}, depth.Asks...)
		}
		bids, _ := js.Get("result").Get("buy").Array()
		for _, b := range bids {
			uu := b.(map[string]interface{})
			price, _ := uu["Rate"].(json.Number).Float64()
			amount, _ := uu["Quantity"].(json.Number).Float64()
			depth.Bids = append(depth.Bids, Unit{price, amount})
		}
		return depth, nil
	}

	d, err := ProcessResp(status, js, respOk, bt.respErr)
	if err == nil {
		depth = d.(Depth)
	}
	return
}

func (bt *Bittrex) OrderState(s interface{}) string {
	return s.(string)
}

func (bt *Bittrex) OrderSide(s string) string {
	return s
}

func (bt *Bittrex) NewOrder(o *Order) (id string, err error) {
	return
}

func (bt *Bittrex) CancelOrder(o *Order) (err error) {
	return
}

func (bt *Bittrex) QueryOrder(o *Order) (order Order, err error) {
	return
}

func NewBittrex() Exchange {
	return new(Bittrex)
}

func init() {
	RegisterEx("bittrex", NewBittrex)
}
