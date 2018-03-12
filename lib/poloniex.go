package lib

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	. "github.com/bitly/go-simplejson"
)

type Poloniex struct {
	accesskeyid, secretkeyid string
}

func (p *Poloniex) respErr(js *Json) (interface{}, error) {
	return nil, nil
}

func (p *Poloniex) ToSymbol(cp *CurrencyPair) string {
	return strings.ToUpper(cp.ToSymbol2("_"))
}

func (p *Poloniex) NormSymbol(cp *string) string {
	return *cp
}

func (p *Poloniex) sendReq(method, path string,
	params map[string][]string, sign bool) (int, *Json, error) {
	header := map[string][]string{
		"Content-Type": {`application/x-www-form-urlencoded`},
	}

	req := &http.Request{
		Method: method,
		Header: header,
	}

	req.URL, _ = url.Parse("https://poloniex.com" + path)
	if sign {
		sign_params := map[string][]string{
			"ex_key": {p.accesskeyid},
		}

		for k, v := range params {
			sign_params[k] = v
		}

		q := req.URL.Query()
		q = sign_params
		data := q.Encode() + "&secret_key=" + p.secretkeyid
		q.Add("sign", strings.ToUpper(GetMD5Hash(data)))
		req.URL.RawQuery = q.Encode()
	} else {
		q := req.URL.Query()
		q = params
		req.URL.RawQuery = q.Encode()
	}
	return recvResp(req)
}

func (p *Poloniex) SetKey(access, secret string) {
	p.accesskeyid = access
	p.secretkeyid = secret
}

func (p *Poloniex) GetBalance() (balances []Balance, err error) {
	return
}

func (p *Poloniex) GetPrice(cp *CurrencyPair) (price Price, err error) {
	return
}

func (p *Poloniex) GetSymbols() (symbols []string, err error) {
	params := map[string][]string{
		"command": {"returnTicker"},
	}
	status, js, err := p.sendReq("GET", "/public", params, false)
	if err != nil {
		return
	}

	respOk := func(js *Json) (interface{}, error) {
		var s []string
		reason, err := js.Get("error").String()
		if err == nil {
			return nil, errors.New(reason)
		}

		data, _ := js.Map()
		for symbol, _ := range data {
			currency := strings.Split(symbol, "_")
			s = append(s, strings.ToLower(currency[1]+"_"+currency[0]))
		}
		return s, nil
	}

	s, err := ProcessResp(status, js, respOk, p.respErr)
	if err == nil {
		symbols = s.([]string)
	}
	return
}

func (p *Poloniex) GetDepth(cp *CurrencyPair) (depth Depth, err error) {
	params := map[string][]string{
		"command":      {"returnOrderBook"},
		"currencyPair": {p.ToSymbol(cp)},
		//"depth":        {"5"},
	}

	status, js, err := p.sendReq("GET", "/public", params, false)
	if err != nil {
		return
	}

	respOk := func(js *Json) (interface{}, error) {
		var depth Depth
		reason, err := js.Get("error").String()
		if err == nil {
			return nil, errors.New(reason)
		}

		asks, _ := js.Get("asks").Array()
		for _, a := range asks {
			uu := a.([]interface{})
			price, _ := strconv.ParseFloat(uu[0].(string), 64)
			amount, _ := uu[1].(json.Number).Float64()
			depth.Asks = append([]Unit{Unit{price, amount}}, depth.Asks...)
		}
		bids, _ := js.Get("bids").Array()
		for _, b := range bids {
			uu := b.([]interface{})
			price, _ := strconv.ParseFloat(uu[0].(string), 64)
			amount, _ := uu[1].(json.Number).Float64()
			depth.Bids = append(depth.Bids, Unit{price, amount})
		}
		return depth, nil
	}

	d, err := ProcessResp(status, js, respOk, p.respErr)
	if err == nil {
		depth = d.(Depth)
	}
	return
}

func (p *Poloniex) OrderState(s interface{}) string {
	return s.(string)
}

func (p *Poloniex) OrderSide(s string) string {
	return s
}

func (p *Poloniex) NewOrder(o *Order) (id string, err error) {
	return
}

func (p *Poloniex) CancelOrder(o *Order) (err error) {
	return
}

func (p *Poloniex) QueryOrder(o *Order) (order Order, err error) {
	return
}

func NewPoloniex() Exchange {
	return new(Poloniex)
}

func init() {
	RegisterEx("poloniex", NewPoloniex)
}
