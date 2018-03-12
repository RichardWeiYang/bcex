package lib

import (
	"encoding/json"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	. "github.com/bitly/go-simplejson"
)

type Gate struct {
	accesskeyid, secretkeyid string
}

func (gate *Gate) respErr(js *Json) (interface{}, error) {
	return nil, nil
}

func (gate *Gate) ToSymbol(cp *CurrencyPair) string {
	return cp.ToSymbol("_")
}

func (gate *Gate) NormSymbol(cp *string) string {
	return *cp
}

func (gate *Gate) sendReq(method, path string,
	params map[string][]string, sign bool) (int, *Json, error) {
	header := map[string][]string{
		"Content-Type": {`application/x-www-form-urlencoded`},
	}

	req := &http.Request{
		Method: method,
		Header: header,
	}

	req.URL, _ = url.Parse("http://data.gate.io" + path)
	if sign {
		sign_params := map[string][]string{
			"ex_key": {gate.accesskeyid},
		}

		for k, v := range params {
			sign_params[k] = v
		}

		q := req.URL.Query()
		q = sign_params
		data := q.Encode() + "&secret_key=" + gate.secretkeyid
		q.Add("sign", strings.ToUpper(GetMD5Hash(data)))
		req.URL.RawQuery = q.Encode()
	} else {
		q := req.URL.Query()
		q = params
		req.URL.RawQuery = q.Encode()
	}
	return recvResp(req)
}

func (gate *Gate) SetKey(access, secret string) {
	gate.accesskeyid = access
	gate.secretkeyid = secret
}

func (gate *Gate) GetBalance() (balances []Balance, err error) {
	return
}

func (gate *Gate) GetPrice(cp *CurrencyPair) (price Price, err error) {
	return
}

func (gate *Gate) GetSymbols() (symbols []string, err error) {
	status, js, err := gate.sendReq("GET", "/api2/1/pairs", nil, false)
	if err != nil {
		return
	}

	respOk := func(js *Json) (interface{}, error) {
		var s []string
		data, _ := js.Array()
		for _, d := range data {
			dd := d.(string)
			s = append(s, dd)
		}
		return s, nil
	}

	s, err := ProcessResp(status, js, respOk, gate.respErr)
	if err == nil {
		symbols = s.([]string)
	}
	return
}

func (gate *Gate) GetDepth(cp *CurrencyPair) (depth Depth, err error) {
	status, js, err := gate.sendReq("GET", "/api2/1/orderBook/"+gate.ToSymbol(cp), nil, false)
	if err != nil {
		return
	}

	respOk := func(js *Json) (interface{}, error) {
		var depth Depth
		asks, _ := js.Get("asks").Array()
		for _, a := range asks {
			var price, amount float64
			uu := a.([]interface{})
			if _, ok := uu[0].(string); ok {
				price, _ = strconv.ParseFloat(uu[0].(string), 64)
			} else if _, ok := uu[0].(json.Number); ok {
				price, _ = uu[0].(json.Number).Float64()
			}
			if _, ok := uu[1].(string); ok {
				amount, _ = strconv.ParseFloat(uu[1].(string), 64)
			} else if _, ok := uu[1].(json.Number); ok {
				amount, _ = uu[1].(json.Number).Float64()
			}
			depth.Asks = append(depth.Asks, Unit{price, amount})
		}
		bids, _ := js.Get("bids").Array()
		for _, b := range bids {
			var price, amount float64
			uu := b.([]interface{})
			if _, ok := uu[0].(string); ok {
				price, _ = strconv.ParseFloat(uu[0].(string), 64)
			} else if _, ok := uu[0].(json.Number); ok {
				price, _ = uu[0].(json.Number).Float64()
			}
			if _, ok := uu[1].(string); ok {
				amount, _ = strconv.ParseFloat(uu[1].(string), 64)
			} else if _, ok := uu[1].(json.Number); ok {
				amount, _ = uu[1].(json.Number).Float64()
			}
			depth.Bids = append(depth.Bids, Unit{price, amount})
		}
		return depth, nil
	}

	d, err := ProcessResp(status, js, respOk, gate.respErr)
	if err == nil {
		depth = d.(Depth)
	}
	return
}

func (gate *Gate) OrderState(s interface{}) string {
	return s.(string)
}

func (gate *Gate) OrderSide(s string) string {
	return s
}

func (gate *Gate) NewOrder(o *Order) (id string, err error) {
	return
}

func (gate *Gate) CancelOrder(o *Order) (err error) {
	return
}

func (gate *Gate) QueryOrder(o *Order) (order Order, err error) {
	return
}

func NewGate() Exchange {
	return new(Gate)
}

func init() {
	RegisterEx("gate", NewGate)
}
