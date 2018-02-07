package lib

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	. "github.com/bitly/go-simplejson"
)

/*
 * Reference page: https://developer.big.one
 */

type BigOne struct {
	accesskeyid, secretkeyid string
}

func (bo *BigOne) respErr(js *Json) (interface{}, error) {
	reason, _ := js.Get("error").Get("description").String()
	err := errors.New(reason)
	return nil, err
}

func (bo *BigOne) ToSymbol(cp *CurrencyPair) string {
	return strings.ToUpper(cp.ToSymbol("-"))
}

func (bo *BigOne) NormSymbol(cp *string) string {
	return strings.ToLower(strings.Replace(*cp, "-", "_", 1))
}

func (bo *BigOne) sendReq(method, path string,
	body map[string]string, sign bool) (int, []byte) {

	var header map[string][]string

	req := &http.Request{
		Method: method,
	}

	req.URL, _ = url.Parse("https://api.big.one" + path)

	if sign {
		header = map[string][]string{
			"Authorization": {"Bearer " + bo.accesskeyid},
			"User-Agent":    {`standard browser user agent format`},
			"Big-Device-Id": {bo.secretkeyid},
			"Content-Type":  {`application/json`},
		}
		req.Header = header
	}

	if body != nil {
		jsonBody, _ := json.Marshal(body)
		req.Body = ioutil.NopCloser(bytes.NewBuffer(jsonBody))
	}

	return recvResp(req)
}

func (bo *BigOne) GetBalance() (balances []Balance, err error) {
	status, body := bo.sendReq("GET", "/accounts", nil, true)
	js, _ := NewJson(body)

	respOk := func(js *Json) (interface{}, error) {
		bs, _ := js.Get("data").Array()
		for _, b := range bs {
			bt := b.(map[string]interface{})
			if bt["active_balance"].(string) != "0.00000000" {
				balances = append(balances,
					Balance{Currency: bt["account_type"].(string),
						Balance: bt["active_balance"].(string)})
			}
		}
		return balances, nil
	}

	b, err := ProcessResp(status, js, respOk, bo.respErr)
	if err == nil {
		balances = b.([]Balance)
	}

	return
}

func (bo *BigOne) Alive() bool {
	status, _ := bo.sendReq("GET", "/accounts", nil, true)

	_, err := ProcessResp(status, nil, isAlive, notAlive)

	if err != nil {
		return true
	} else {
		return false
	}
}

func (bo *BigOne) SetKey(access, secret string) {
	bo.accesskeyid = access
	bo.secretkeyid = GetUUID()
}

func (bo *BigOne) GetPrice(cp *CurrencyPair) (price Price, err error) {
	status, body := bo.sendReq("GET", "/markets/"+bo.ToSymbol(cp), nil, false)
	js, _ := NewJson(body)

	respOk := func(js *Json) (interface{}, error) {
		price_s, _ := js.Get("data").Get("ticker").Get("price").String()
		price, _ := strconv.ParseFloat(price_s, 64)
		return Price{price}, nil
	}

	p, err := ProcessResp(status, js, respOk, bo.respErr)
	if err == nil {
		price = p.(Price)
	}

	return
}

func (bo *BigOne) GetSymbols() (symbols []string, err error) {
	status, body := bo.sendReq("GET", "/markets", nil, false)
	js, _ := NewJson(body)

	respOk := func(js *Json) (interface{}, error) {
		var s []string
		data, _ := js.Get("data").Array()
		for _, d := range data {
			dd := d.(map[string]interface{})
			base := strings.ToLower(dd["base"].(string))
			quote := strings.ToLower(dd["quote"].(string))
			s = append(s, quote+"_"+base)
		}
		return s, nil
	}

	s, err := ProcessResp(status, js, respOk, bo.respErr)
	if err == nil {
		symbols = s.([]string)
	}
	return
}

func (bo *BigOne) GetDepth(cp *CurrencyPair) (depth Depth, err error) {
	status, body := bo.sendReq("GET", "/markets/"+bo.ToSymbol(cp), nil, false)
	js, _ := NewJson(body)

	respOk := func(js *Json) (interface{}, error) {
		var depth Depth
		asks, _ := js.Get("data").Get("depth").Get("asks").Array()
		for _, a := range asks {
			uu := a.(map[string]interface{})
			price, _ := strconv.ParseFloat(uu["price"].(string), 64)
			amount, _ := strconv.ParseFloat(uu["amount"].(string), 64)
			depth.Asks = append([]Unit{Unit{price, amount}}, depth.Asks...)
		}
		bids, _ := js.Get("data").Get("depth").Get("bids").Array()
		for _, b := range bids {
			uu := b.(map[string]interface{})
			price, _ := strconv.ParseFloat(uu["price"].(string), 64)
			amount, _ := strconv.ParseFloat(uu["amount"].(string), 64)
			depth.Bids = append(depth.Bids, Unit{price, amount})
		}
		return depth, nil
	}

	d, err := ProcessResp(status, js, respOk, bo.respErr)
	if err == nil {
		depth = d.(Depth)
	}
	return
}

// different on bigone
func getSide(side string) string {
	if side == "sell" {
		return "ASK"
	}
	return "BID"
}

func (bo *BigOne) OrderState(s interface{}) string {
	if s.(string) == "open" {
		return Alive
	}
	return s.(string)
}

func (bo *BigOne) OrderSide(s string) string {
	if s == "ASK" {
		return "sell"
	} else {
		return "buy"
	}
}

func (bo *BigOne) NewOrder(o *Order) (id string, err error) {
	params := map[string]string{
		"order_market": bo.ToSymbol(&o.CP),
		"order_side":   getSide(o.Side),
		"amount":       strconv.FormatFloat(o.Amount, 'f', -1, 64),
		"price":        strconv.FormatFloat(o.Price, 'f', -1, 64),
	}

	status, body := bo.sendReq("POST", "/orders", params, true)
	js, _ := NewJson(body)

	respOk := func(js *Json) (interface{}, error) {
		id, _ = js.Get("data").Get("order_id").String()
		return id, nil
	}

	oid, err := ProcessResp(status, js, respOk, bo.respErr)
	if err == nil {
		id = oid.(string)
	}
	return
}

func (bo *BigOne) CancelOrder(o *Order) (err error) {
	status, body := bo.sendReq("DELETE", "/orders/"+o.Id, nil, true)
	js, _ := NewJson(body)

	respOk := func(js *Json) (interface{}, error) {
		return nil, nil
	}

	_, err = ProcessResp(status, js, respOk, bo.respErr)
	return
}

func NewBigOne() Exchange {
	return new(BigOne)
}

func init() {
	RegisterEx("bigone", NewBigOne)
}
