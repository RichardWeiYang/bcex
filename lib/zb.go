package lib

import (
	"encoding/json"
	"net/http"
	"net/url"
	"strings"

	. "github.com/bitly/go-simplejson"
)

type ZB struct {
	accesskeyid, secretkeyid string
}

func (zb *ZB) respErr(js *Json) (interface{}, error) {
	return nil, nil
}

func (zb *ZB) ToSymbol(cp *CurrencyPair) string {
	return cp.ToSymbol("_")
}

func (zb *ZB) NormSymbol(cp *string) string {
	return *cp
}

func (zb *ZB) sendReq(method, path string,
	params map[string][]string, sign bool) (int, *Json, error) {
	header := map[string][]string{
		"Content-Type": {`application/x-www-form-urlencoded`},
	}

	req := &http.Request{
		Method: method,
		Header: header,
	}

	req.URL, _ = url.Parse("http://api.zb.com" + path)
	if sign {
		sign_params := map[string][]string{
			"ex_key": {zb.accesskeyid},
		}

		for k, v := range params {
			sign_params[k] = v
		}

		q := req.URL.Query()
		q = sign_params
		data := q.Encode() + "&secret_key=" + zb.secretkeyid
		q.Add("sign", strings.ToUpper(GetMD5Hash(data)))
		req.URL.RawQuery = q.Encode()
	} else {
		q := req.URL.Query()
		q = params
		req.URL.RawQuery = q.Encode()
	}
	return recvResp(req)
}

func (zb *ZB) SetKey(access, secret string) {
	zb.accesskeyid = access
	zb.secretkeyid = secret
}

func (zb *ZB) GetBalance() (balances []Balance, err error) {
	return
}

func (zb *ZB) GetPrice(cp *CurrencyPair) (price Price, err error) {
	return
}

func (zb *ZB) GetSymbols() (symbols []string, err error) {
	status, js, err := zb.sendReq("GET", "/data/v1/markets", nil, false)
	if err != nil {
		return
	}

	respOk := func(js *Json) (interface{}, error) {
		var s []string
		data, _ := js.Map()
		for symbol, _ := range data {
			s = append(s, symbol)
		}
		return s, nil
	}

	s, err := ProcessResp(status, js, respOk, zb.respErr)
	if err == nil {
		symbols = s.([]string)
	}
	return
}

func (zb *ZB) GetDepth(cp *CurrencyPair) (depth Depth, err error) {
	params := map[string][]string{
		"market": {zb.ToSymbol(cp)},
		"size":   {"10"},
	}

	status, js, err := zb.sendReq("GET", "/data/v1/depth", params, false)
	if err != nil {
		return
	}

	respOk := func(js *Json) (interface{}, error) {
		var depth Depth
		asks, _ := js.Get("asks").Array()
		for _, a := range asks {
			uu := a.([]interface{})
			price, _ := uu[0].(json.Number).Float64()
			amount, _ := uu[1].(json.Number).Float64()
			depth.Asks = append(depth.Asks, Unit{price, amount})
		}
		bids, _ := js.Get("bids").Array()
		for _, b := range bids {
			uu := b.([]interface{})
			price, _ := uu[0].(json.Number).Float64()
			amount, _ := uu[1].(json.Number).Float64()
			depth.Bids = append(depth.Bids, Unit{price, amount})
		}
		return depth, nil
	}

	d, err := ProcessResp(status, js, respOk, zb.respErr)
	if err == nil {
		depth = d.(Depth)
	}
	return
}

func (zb *ZB) OrderState(s interface{}) string {
	return s.(string)
}

func (zb *ZB) OrderSide(s string) string {
	return s
}

func (zb *ZB) NewOrder(o *Order) (id string, err error) {
	return
}

func (zb *ZB) CancelOrder(o *Order) (err error) {
	return
}

func (zb *ZB) QueryOrder(o *Order) (order Order, err error) {
	return
}

func NewZB() Exchange {
	return new(ZB)
}

func init() {
	RegisterEx("zb", NewZB)
}
