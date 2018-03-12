package lib

import (
	"errors"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	. "github.com/bitly/go-simplejson"
)

type Exx struct {
	accesskeyid, secretkeyid string
}

func (exx *Exx) respErr(js *Json) (interface{}, error) {
	return nil, nil
}

func (exx *Exx) ToSymbol(cp *CurrencyPair) string {
	return cp.ToSymbol("_")
}

func (exx *Exx) NormSymbol(cp *string) string {
	return *cp
}

func (exx *Exx) sendReq(method, path string,
	params map[string][]string, sign bool) (int, *Json, error) {
	header := map[string][]string{
		"Content-Type": {`application/x-www-form-urlencoded`},
	}

	req := &http.Request{
		Method: method,
		Header: header,
	}

	req.URL, _ = url.Parse("https://api.exx.com" + path)
	if sign {
		sign_params := map[string][]string{
			"ex_key": {exx.accesskeyid},
		}

		for k, v := range params {
			sign_params[k] = v
		}

		q := req.URL.Query()
		q = sign_params
		data := q.Encode() + "&secret_key=" + exx.secretkeyid
		q.Add("sign", strings.ToUpper(GetMD5Hash(data)))
		req.URL.RawQuery = q.Encode()
	} else {
		q := req.URL.Query()
		q = params
		req.URL.RawQuery = q.Encode()
	}
	return recvResp(req)
}

func (exx *Exx) SetKey(access, secret string) {
	exx.accesskeyid = access
	exx.secretkeyid = secret
}

func (exx *Exx) GetBalance() (balances []Balance, err error) {
	return
}

func (exx *Exx) GetPrice(cp *CurrencyPair) (price Price, err error) {
	return
}

func (exx *Exx) GetSymbols() (symbols []string, err error) {
	status, js, err := exx.sendReq("GET", "/data/v1/markets", nil, false)
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

	s, err := ProcessResp(status, js, respOk, exx.respErr)
	if err == nil {
		symbols = s.([]string)
	}
	return
}

func (exx *Exx) GetDepth(cp *CurrencyPair) (depth Depth, err error) {
	params := map[string][]string{
		"currency": {exx.ToSymbol(cp)},
	}

	status, js, err := exx.sendReq("GET", "/data/v1/depth", params, false)
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
			amount, _ := strconv.ParseFloat(uu[1].(string), 64)
			depth.Asks = append(depth.Asks, Unit{price, amount})
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

	d, err := ProcessResp(status, js, respOk, exx.respErr)
	if err == nil {
		depth = d.(Depth)
	}
	return
}

func (exx *Exx) OrderState(s interface{}) string {
	return s.(string)
}

func (exx *Exx) OrderSide(s string) string {
	return s
}

func (exx *Exx) NewOrder(o *Order) (id string, err error) {
	return
}

func (exx *Exx) CancelOrder(o *Order) (err error) {
	return
}

func (exx *Exx) QueryOrder(o *Order) (order Order, err error) {
	return
}

func NewExx() Exchange {
	return new(Exx)
}

func init() {
	RegisterEx("exx", NewExx)
}
