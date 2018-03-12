package lib

import (
	"net/http"
	"net/url"
	"strconv"
	"strings"

	. "github.com/bitly/go-simplejson"
)

type HitBTC struct {
	accesskeyid, secretkeyid string
}

func (hb *HitBTC) respErr(js *Json) (interface{}, error) {
	return nil, nil
}

func (hb *HitBTC) ToSymbol(cp *CurrencyPair) string {
	return strings.ToUpper(cp.ToSymbol(""))
}

func (hb *HitBTC) NormSymbol(cp *string) string {
	return *cp
}

func (hb *HitBTC) sendReq(method, path string,
	params map[string][]string, sign bool) (int, *Json, error) {
	header := map[string][]string{
		"Content-Type": {`application/x-www-form-urlencoded`},
	}

	req := &http.Request{
		Method: method,
		Header: header,
	}

	req.URL, _ = url.Parse("https://api.hitbtc.com" + path)
	if sign {
		sign_params := map[string][]string{
			"ex_key": {hb.accesskeyid},
		}

		for k, v := range params {
			sign_params[k] = v
		}

		q := req.URL.Query()
		q = sign_params
		data := q.Encode() + "&secret_key=" + hb.secretkeyid
		q.Add("sign", strings.ToUpper(GetMD5Hash(data)))
		req.URL.RawQuery = q.Encode()
	} else {
		q := req.URL.Query()
		q = params
		req.URL.RawQuery = q.Encode()
	}
	return recvResp(req)
}

func (hb *HitBTC) SetKey(access, secret string) {
	hb.accesskeyid = access
	hb.secretkeyid = secret
}

func (hb *HitBTC) GetBalance() (balances []Balance, err error) {
	return
}

func (hb *HitBTC) GetPrice(cp *CurrencyPair) (price Price, err error) {
	return
}

func (hb *HitBTC) GetSymbols() (symbols []string, err error) {
	status, js, err := hb.sendReq("GET", "/api/2/public/symbol", nil, false)
	if err != nil {
		return
	}

	respOk := func(js *Json) (interface{}, error) {
		var s []string
		data, _ := js.Array()
		for _, d := range data {
			dd := d.(map[string]interface{})
			symbol := strings.ToLower(dd["id"].(string))
			base := symbol[0:3]
			quote := symbol[3:]
			s = append(s, base+"_"+quote)
		}
		return s, nil
	}

	s, err := ProcessResp(status, js, respOk, hb.respErr)
	if err == nil {
		symbols = s.([]string)
	}
	return
}

func (hb *HitBTC) GetDepth(cp *CurrencyPair) (depth Depth, err error) {
	status, js, err := hb.sendReq("GET", "/api/2/public/orderbook/"+hb.ToSymbol(cp), nil, false)
	if err != nil {
		return
	}

	respOk := func(js *Json) (interface{}, error) {
		var depth Depth
		asks, _ := js.Get("ask").Array()
		for _, a := range asks {
			uu := a.(map[string]interface{})
			price, _ := strconv.ParseFloat(uu["price"].(string), 64)
			amount, _ := strconv.ParseFloat(uu["size"].(string), 64)
			depth.Asks = append([]Unit{Unit{price, amount}}, depth.Asks...)
		}
		bids, _ := js.Get("bid").Array()
		for _, b := range bids {
			uu := b.(map[string]interface{})
			price, _ := strconv.ParseFloat(uu["price"].(string), 64)
			amount, _ := strconv.ParseFloat(uu["size"].(string), 64)
			depth.Bids = append(depth.Bids, Unit{price, amount})
		}
		return depth, nil
	}

	d, err := ProcessResp(status, js, respOk, hb.respErr)
	if err == nil {
		depth = d.(Depth)
	}
	return
}

func (hb *HitBTC) OrderState(s interface{}) string {
	return s.(string)
}

func (hb *HitBTC) OrderSide(s string) string {
	return s
}

func (hb *HitBTC) NewOrder(o *Order) (id string, err error) {
	return
}

func (hb *HitBTC) CancelOrder(o *Order) (err error) {
	return
}

func (hb *HitBTC) QueryOrder(o *Order) (order Order, err error) {
	return
}

func NewHitBTC() Exchange {
	return new(HitBTC)
}

func init() {
	RegisterEx("hitbtc", NewHitBTC)
}
