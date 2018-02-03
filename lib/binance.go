package lib

import (
	"errors"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	. "github.com/bitly/go-simplejson"
)

/*
 * Reference page: https://github.com/binance-exchange/binance-official-api-docs/blob/master/rest-api.md
 */

type Binance struct {
	accesskeyid, secretkeyid string
}

func (bn *Binance) respErr(js *Json) (interface{}, error) {
	reason, _ := js.Get("msg").String()
	err := errors.New(reason)
	return nil, err
}

func (bn *Binance) ToSymbol(cp *CurrencyPair) string {
	return strings.ToUpper(cp.ToSymbol(""))
}

func (bn *Binance) sendReq(method, path string,
	params map[string][]string, sign bool) (int, []byte) {
	header := map[string][]string{
		"Content-Type": {`application/x-www-form-urlencoded`},
	}

	req := &http.Request{
		Method: method,
		Header: header,
	}

	req.URL, _ = url.Parse("https://api.binance.com" + path)

	q := req.URL.Query()
	q = params
	if sign {
		q.Add("signature", ComputeHmac256(q.Encode(), bn.secretkeyid))
		req.Header.Add("X-MBX-APIKEY", bn.accesskeyid)
	}
	req.URL.RawQuery = q.Encode()
	return recvResp(req)
}

func (bn *Binance) GetBalance() (balances []Balance, err error) {
	params := map[string][]string{
		"recvWindow": {`5000`},
		"timestamp":  {strconv.FormatInt(time.Now().UnixNano(), 10)[0:13]},
	}
	status, body := bn.sendReq("GET", "/api/v3/account", params, true)
	js, _ := NewJson(body)

	respOk := func(js *Json) (interface{}, error) {
		bs, _ := js.Get("balances").Array()
		for _, b := range bs {
			bt := b.(map[string]interface{})
			if bt["free"].(string) != "0.00000000" {
				balances = append(balances,
					Balance{Currency: bt["asset"].(string),
						Balance: bt["free"].(string)})
			}
		}
		return balances, nil
	}

	b, err := ProcessResp(status, js, respOk, bn.respErr)
	if err == nil {
		balances = b.([]Balance)
	}

	return
}

func (bn *Binance) Alive() bool {
	status, _ := bn.sendReq("GET", "/api/v1/time", nil, false)
	_, err := ProcessResp(status, nil, isAlive, notAlive)

	if err != nil {
		return true
	} else {
		return false
	}
}

func (bn *Binance) SetKey(access, secret string) {
	bn.accesskeyid = access
	bn.secretkeyid = secret
}

func (bn *Binance) GetPrice(cp *CurrencyPair) (price Price, err error) {
	params := map[string][]string{
		"symbol": {bn.ToSymbol(cp)},
	}
	status, body := bn.sendReq("GET", "/api/v3/ticker/price", params, false)
	js, _ := NewJson(body)

	respOk := func(js *Json) (interface{}, error) {
		price_s, _ := js.Get("price").String()
		price, _ := strconv.ParseFloat(price_s, 64)

		return Price{price}, nil
	}

	p, err := ProcessResp(status, js, respOk, bn.respErr)
	if err == nil {
		price = p.(Price)
	}
	return
}

func (bn *Binance) GetSymbols() (symbols []string, err error) {
	status, body := bn.sendReq("GET", "/api/v1/exchangeInfo", nil, false)
	js, _ := NewJson(body)

	respOk := func(js *Json) (interface{}, error) {
		var s []string
		data, _ := js.Get("symbols").Array()
		for _, d := range data {
			dd := d.(map[string]interface{})
			base := strings.ToLower(dd["baseAsset"].(string))
			quote := strings.ToLower(dd["quoteAsset"].(string))
			s = append(s, base+"_"+quote)
		}
		return s, nil
	}

	s, err := ProcessResp(status, js, respOk, bn.respErr)
	if err == nil {
		symbols = s.([]string)
	}
	return
}

func (bn *Binance) GetDepth(cp *CurrencyPair) (depth Depth, err error) {
	params := map[string][]string{
		"symbol": {bn.ToSymbol(cp)},
	}
	status, body := bn.sendReq("GET", "/api/v1/depth", params, false)
	js, _ := NewJson(body)

	respOk := func(js *Json) (interface{}, error) {
		var depth Depth
		asks, _ := js.Get("asks").Array()
		for _, a := range asks {
			uu := a.([]interface{})
			price, _ := strconv.ParseFloat(uu[0].(string), 64)
			amount, _ := strconv.ParseFloat(uu[1].(string), 64)
			depth.Asks = append([]Unit{Unit{price, amount}}, depth.Asks...)
		}
		bids, _ := js.Get("bids").Array()
		for _, b := range bids {
			uu := b.([]interface{})
			price, _ := strconv.ParseFloat(uu[0].(string), 64)
			amount, _ := strconv.ParseFloat(uu[1].(string), 64)
			depth.Bids = append(depth.Bids, Unit{price, amount})
		}
		return depth, nil
	}

	d, err := ProcessResp(status, js, respOk, bn.respErr)
	if err == nil {
		depth = d.(Depth)
	}
	return
}

func NewBinance() Exchange {
	return new(Binance)
}

func init() {
	RegisterEx("binance", NewBinance)
}
