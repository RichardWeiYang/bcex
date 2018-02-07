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

/*
 * Reference page: https://www.okex.com/rest_api.html
 */

type Okex struct {
	accesskeyid, secretkeyid string
}

func (ok *Okex) respErr(js *Json) (interface{}, error) {
	return nil, errors.New("Unknow")
}

func (ok *Okex) ToSymbol(cp *CurrencyPair) string {
	return cp.ToSymbol("_")
}

func (ok *Okex) NormSymbol(cp *string) string {
	return *cp
}

func (ok *Okex) sendReq(method, path string,
	params map[string][]string, sign bool) (int, []byte) {
	header := map[string][]string{
		"Content-Type": {`application/x-www-form-urlencoded`},
	}

	req := &http.Request{
		Method: method,
		Header: header,
	}

	req.URL, _ = url.Parse("https://www.okex.com" + path)
	if sign {
		sign_params := map[string][]string{
			"api_key": {ok.accesskeyid},
		}

		for k, v := range params {
			sign_params[k] = v
		}

		q := req.URL.Query()
		q = sign_params
		//q2 := q
		//q2.Add("secret_key", okex.secretkeyid)
		data := q.Encode() + "&secret_key=" + ok.secretkeyid
		q.Add("sign", strings.ToUpper(GetMD5Hash(data)))
		req.URL.RawQuery = q.Encode()
	} else {
		q := req.URL.Query()
		q = params
		req.URL.RawQuery = q.Encode()
	}
	return recvResp(req)
}

func (ok *Okex) GetBalance() (balances []Balance, err error) {
	status, body := ok.sendReq("POST", "/api/v1/userinfo.do", nil, true)
	js, _ := NewJson(body)

	respOk := func(js *Json) (interface{}, error) {
		result, _ := js.Get("result").Bool()
		if result {
			free, _ := js.Get("info").Get("funds").Get("free").Map()
			for cur, b := range free {
				if b.(string) != "0" {
					balances = append(balances,
						Balance{cur, b.(string)})
				}
			}
			return balances, nil
		} else {
			code, _ := js.Get("error_code").Int64()
			err = errors.New(ok.code2reason(code))
			return nil, err
		}
		return nil, errors.New("Unknow")
	}

	b, err := ProcessResp(status, js, respOk, ok.respErr)
	if err == nil {
		balances = b.([]Balance)
	}

	return
}

func (ok *Okex) Alive() bool {
	status, _ := ok.sendReq("GET", "/api/v1/exchange_rate.do", nil, false)
	_, err := ProcessResp(status, nil, isAlive, notAlive)

	if err != nil {
		return true
	} else {
		return false
	}
}

func (ok *Okex) SetKey(access, secret string) {
	ok.accesskeyid = access
	ok.secretkeyid = secret
}

func (ok *Okex) GetPrice(cp *CurrencyPair) (price Price, err error) {
	params := map[string][]string{
		"symbol": {ok.ToSymbol(cp)},
	}

	status, body := ok.sendReq("GET", "/api/v1/ticker.do", params, false)
	js, _ := NewJson(body)

	respOk := func(js *Json) (interface{}, error) {
		code, e := js.Get("error_code").Int64()
		if e == nil {
			err = errors.New(ok.code2reason(code))
			return nil, err
		}

		last_s, _ := js.Get("ticker").Get("last").String()
		last, _ := strconv.ParseFloat(last_s, 64)
		return Price{last}, nil
	}

	p, err := ProcessResp(status, js, respOk, ok.respErr)
	if err == nil {
		price = p.(Price)
	}
	return
}

func (ok *Okex) GetSymbols() (symbols []string, err error) {
	status, body := ok.sendReq("GET", "/v2/markets/products", nil, false)
	js, _ := NewJson(body)

	respOk := func(js *Json) (interface{}, error) {
		code, e := js.Get("error_code").Int64()
		if e == nil {
			err = errors.New(ok.code2reason(code))
			return nil, err
		}

		var s []string
		data, _ := js.Get("data").Array()
		for _, d := range data {
			dd := d.(map[string]interface{})
			symbol := dd["symbol"].(string)
			s = append(s, symbol)
		}
		return s, nil
	}

	s, err := ProcessResp(status, js, respOk, ok.respErr)
	if err == nil {
		symbols = s.([]string)
	}
	return
}

func (ok *Okex) GetDepth(cp *CurrencyPair) (depth Depth, err error) {
	params := map[string][]string{
		"symbol": {ok.ToSymbol(cp)},
	}

	status, body := ok.sendReq("GET", "/api/v1/depth.do", params, false)
	js, _ := NewJson(body)

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

	d, err := ProcessResp(status, js, respOk, ok.respErr)
	if err == nil {
		depth = d.(Depth)
	}
	return
}

func (ok *Okex) OrderState(s interface{}) string {
	return s.(string)
}

func (ok *Okex) OrderSide(s string) string {
	return s
}

func (ok *Okex) NewOrder(o *Order) (id string, err error) {
	params := map[string][]string{
		"symbol": {ok.ToSymbol(&o.CP)},
		"type":   {o.Side},
		"amount": {strconv.FormatFloat(o.Amount, 'f', -1, 64)},
		"price":  {strconv.FormatFloat(o.Price, 'f', -1, 64)},
	}

	status, body := ok.sendReq("POST", "/api/v1/trade.do", params, true)
	js, _ := NewJson(body)

	respOk := func(js *Json) (interface{}, error) {
		result, _ := js.Get("result").Bool()
		if result {
			id, _ := js.Get("order_id").Int64()
			return strconv.FormatInt(id, 10), nil
		} else {
			code, _ := js.Get("error_code").Int64()
			err = errors.New(ok.code2reason(code))
			return nil, err
		}
	}

	oid, err := ProcessResp(status, js, respOk, ok.respErr)
	if err == nil {
		id = oid.(string)
	}
	return
}

func (ok *Okex) CancelOrder(o *Order) (err error) {
	params := map[string][]string{
		"symbol":   {ok.ToSymbol(&o.CP)},
		"order_id": {o.Id},
	}

	status, body := ok.sendReq("POST", "/api/v1/cancel_order.do", params, true)
	js, _ := NewJson(body)

	respOk := func(js *Json) (interface{}, error) {
		result, _ := js.Get("result").Bool()
		if result {
			return nil, nil
		} else {
			code, _ := js.Get("error_code").Int64()
			err = errors.New(ok.code2reason(code))
			return nil, err
		}

	}

	_, err = ProcessResp(status, js, respOk, ok.respErr)
	return
}

func NewOkex() Exchange {
	return new(Okex)
}

func init() {
	RegisterEx("okex", NewOkex)
}
