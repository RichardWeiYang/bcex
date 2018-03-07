package lib

import (
	"errors"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	. "github.com/bitly/go-simplejson"
)

type OCTBTC struct {
	accesskeyid, secretkeyid string
}

func (otc *OCTBTC) respErr(js *Json) (interface{}, error) {
	reason, _ := js.Get("error").Get("message").String()
	err := errors.New(reason)
	return nil, err
}

func (otc *OCTBTC) ToSymbol(cp *CurrencyPair) string {
	return cp.ToSymbol("")
}

func (otc *OCTBTC) NormSymbol(cp *string) string {
	tmp := *cp
	return strings.ToLower(tmp[:3] + "_" + tmp[3:])
}

func (otc *OCTBTC) sendReq(method, path string,
	params map[string][]string, sign bool) (int, *Json, error) {
	header := map[string][]string{
		"Content-Type": {`application/x-www-form-urlencoded`},
	}

	req := &http.Request{
		Method: method,
		Header: header,
	}

	req.URL, _ = url.Parse("https://bb.otcbtc.com" + path)
	if sign {
		sign_params := map[string][]string{
			"access_key": {otc.accesskeyid},
		}

		for k, v := range params {
			sign_params[k] = v
		}

		q := req.URL.Query()
		q = sign_params
		data := method + "|" + path + "|" + q.Encode()
		q.Add("signature", ComputeHmac256(data, otc.secretkeyid))
		req.URL.RawQuery = q.Encode()
	} else {
		q := req.URL.Query()
		q = params
		req.URL.RawQuery = q.Encode()
	}
	return recvResp(req)
}

func (otc *OCTBTC) SetKey(access, secret string) {
	otc.accesskeyid = access
	otc.secretkeyid = secret
}

func (otc *OCTBTC) GetBalance() (balances []Balance, err error) {
	status, js, err := otc.sendReq("GET", "/api/v2/users/me", nil, true)
	if err != nil {
		return
	}

	respOk := func(js *Json) (interface{}, error) {
		bs, _ := js.Get("accounts").Array()
		for _, b := range bs {
			bt := b.(map[string]interface{})
			if bt["balance"].(string) != "0.0" {
				balances = append(balances,
					Balance{Currency: bt["currency"].(string),
						Balance: bt["balance"].(string)})
			}
		}
		return balances, nil
	}

	b, err := ProcessResp(status, js, respOk, otc.respErr)
	if err == nil {
		balances = b.([]Balance)
	}

	return
}

func (otc *OCTBTC) GetPrice(cp *CurrencyPair) (price Price, err error) {
	return
}

func (otc *OCTBTC) GetSymbols() (symbols []string, err error) {
	status, js, err := otc.sendReq("GET", "/api/v2/markets", nil, false)
	if err != nil {
		return
	}

	respOk := func(js *Json) (interface{}, error) {
		var s []string
		data, _ := js.Array()
		for _, d := range data {
			dd := d.(map[string]interface{})
			symbol := strings.ToLower(dd["ticker_id"].(string))
			s = append(s, symbol)
		}
		return s, nil
	}

	s, err := ProcessResp(status, js, respOk, otc.respErr)
	if err == nil {
		symbols = s.([]string)
	}
	return
}

func (otc *OCTBTC) GetDepth(cp *CurrencyPair) (depth Depth, err error) {
	params := map[string][]string{
		"market": {otc.ToSymbol(cp)},
	}
	status, js, err := otc.sendReq("GET", "/api/v2/order_book", params, false)
	if err != nil {
		return
	}

	respOk := func(js *Json) (interface{}, error) {
		var depth Depth
		asks, _ := js.Get("asks").Array()
		for _, a := range asks {
			uu := a.(map[string]interface{})
			price, _ := strconv.ParseFloat(uu["price"].(string), 64)
			amount, _ := strconv.ParseFloat(uu["volume"].(string), 64)
			depth.Asks = append([]Unit{Unit{price, amount}}, depth.Asks...)
		}
		bids, _ := js.Get("bids").Array()
		for _, b := range bids {
			uu := b.(map[string]interface{})
			price, _ := strconv.ParseFloat(uu["price"].(string), 64)
			amount, _ := strconv.ParseFloat(uu["volume"].(string), 64)
			depth.Bids = append(depth.Bids, Unit{price, amount})
		}
		return depth, nil
	}

	d, err := ProcessResp(status, js, respOk, otc.respErr)
	if err == nil {
		depth = d.(Depth)
	}
	return
}

func (otc *OCTBTC) OrderState(s interface{}) string {
	if s.(string) == "wait" {
		return Alive
	} else if s.(string) == "cancel" {
		return Cancelled
	}
	return s.(string)
}

func (otc *OCTBTC) OrderSide(s string) string {
	return s
}

func (otc *OCTBTC) NewOrder(o *Order) (id string, err error) {
	params := map[string][]string{
		"market": {otc.ToSymbol(&o.CP)},
		"side":   {o.Side},
		"volume": {strconv.FormatFloat(o.Amount, 'f', -1, 64)},
		"price":  {strconv.FormatFloat(o.Price, 'f', -1, 64)},
		//"ord_type": {"limit"},
	}

	status, js, err := otc.sendReq("POST", "/api/v2/orders", params, true)
	if err != nil {
		return
	}

	respOk := func(js *Json) (interface{}, error) {
		id, _ := js.Get("id").Int64()
		return strconv.FormatInt(id, 10), nil
	}

	oid, err := ProcessResp(status, js, respOk, otc.respErr)
	if err == nil {
		id = oid.(string)
	}
	return
}

func (otc *OCTBTC) CancelOrder(o *Order) (err error) {
	params := map[string][]string{
		"id": {o.Id},
	}

	status, js, err := otc.sendReq("POST", "/api/v2/order/delete", params, true)
	if err != nil {
		return
	}

	respOk := func(js *Json) (interface{}, error) {
		return nil, nil
	}

	_, err = ProcessResp(status, js, respOk, otc.respErr)
	return
	return
}

func (otc *OCTBTC) QueryOrder(o *Order) (order Order, err error) {
	params := map[string][]string{
		"id": {o.Id},
	}

	status, js, err := otc.sendReq("GET", "/api/v2/order", params, true)
	if err != nil {
		return
	}

	respOk := func(js *Json) (interface{}, error) {
		var order Order
		id, _ := js.Get("id").Int64()
		order.Id = strconv.FormatInt(id, 10)
		market, _ := js.Get("market").String()
		order.CP = NewCurrencyPair2(otc.NormSymbol(&market))
		order.Side, _ = js.Get("side").String()
		price, _ := js.Get("price").String()
		order.Price, _ = strconv.ParseFloat(price, 64)
		amount, _ := js.Get("volume").String()
		order.Amount, _ = strconv.ParseFloat(amount, 64)
		executed, _ := js.Get("executed_volume").String()
		order.Executed, _ = strconv.ParseFloat(executed, 64)
		order.Remain = order.Amount - order.Executed
		status, _ := js.Get("state").String()
		order.State = otc.OrderState(status)
		return order, nil
	}

	od, err := ProcessResp(status, js, respOk, otc.respErr)
	if err == nil {
		order = od.(Order)
	}
	return
}

func NewOTCBTC() Exchange {
	return new(OCTBTC)
}

func init() {
	RegisterEx("otcbtc", NewOTCBTC)
}
