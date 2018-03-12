package lib

import (
	"net/http"
	"net/url"
	"strconv"
	"strings"

	. "github.com/bitly/go-simplejson"
)

type BitStamp struct {
	accesskeyid, secretkeyid string
}

func (bs *BitStamp) respErr(js *Json) (interface{}, error) {
	return nil, nil
}

func (bs *BitStamp) ToSymbol(cp *CurrencyPair) string {
	return cp.ToSymbol("")
}

func (bs *BitStamp) NormSymbol(cp *string) string {
	return *cp
}

func (bs *BitStamp) sendReq(method, path string,
	params map[string][]string, sign bool) (int, *Json, error) {
	header := map[string][]string{
		"Content-Type": {`application/x-www-form-urlencoded`},
	}

	req := &http.Request{
		Method: method,
		Header: header,
	}

	req.URL, _ = url.Parse("https://www.bitstamp.net" + path)
	if sign {
		sign_params := map[string][]string{
			"ex_key": {bs.accesskeyid},
		}

		for k, v := range params {
			sign_params[k] = v
		}

		q := req.URL.Query()
		q = sign_params
		data := q.Encode() + "&secret_key=" + bs.secretkeyid
		q.Add("sign", strings.ToUpper(GetMD5Hash(data)))
		req.URL.RawQuery = q.Encode()
	} else {
		q := req.URL.Query()
		q = params
		req.URL.RawQuery = q.Encode()
	}
	return recvResp(req)
}

func (bs *BitStamp) SetKey(access, secret string) {
	bs.accesskeyid = access
	bs.secretkeyid = secret
}

func (bs *BitStamp) GetBalance() (balances []Balance, err error) {
	return
}

func (bs *BitStamp) GetPrice(cp *CurrencyPair) (price Price, err error) {
	return
}

func (bs *BitStamp) GetSymbols() (symbols []string, err error) {
	status, js, err := bs.sendReq("GET", "/api/v2/trading-pairs-info/", nil, false)
	if err != nil {
		return
	}

	respOk := func(js *Json) (interface{}, error) {
		var s []string
		data, _ := js.Array()
		for _, d := range data {
			dd := d.(map[string]interface{})
			raw := strings.ToLower(dd["name"].(string))
			s = append(s, strings.Replace(raw, "/", "_", 1))
		}
		return s, nil
	}

	s, err := ProcessResp(status, js, respOk, bs.respErr)
	if err == nil {
		symbols = s.([]string)
	}
	return
}

func (bs *BitStamp) GetDepth(cp *CurrencyPair) (depth Depth, err error) {
	status, js, err := bs.sendReq("GET", "/api/v2/order_book/"+bs.ToSymbol(cp), nil, false)
	if err != nil {
		return
	}

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

	d, err := ProcessResp(status, js, respOk, bs.respErr)
	if err == nil {
		depth = d.(Depth)
	}
	return
}

func (bs *BitStamp) OrderState(s interface{}) string {
	return s.(string)
}

func (bs *BitStamp) OrderSide(s string) string {
	return s
}

func (bs *BitStamp) NewOrder(o *Order) (id string, err error) {
	return
}

func (bs *BitStamp) CancelOrder(o *Order) (err error) {
	return
}

func (bs *BitStamp) QueryOrder(o *Order) (order Order, err error) {
	return
}

func NewBitStamp() Exchange {
	return new(BitStamp)
}

func init() {
	RegisterEx("bitstamp", NewBitStamp)
}
