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
	"time"

	. "github.com/bitly/go-simplejson"
)

/*
 * Reference page: https://github.com/huobiapi/API_Docs/wiki
 */

type Huobi struct {
	accesskeyid, secretkeyid string
	account_id               string
}

func (hb *Huobi) respErr(js *Json) (interface{}, error) {
	reason, err := js.Get("err-msg").String()
	if err == nil {
		return nil, errors.New(reason)
	} else {
		return nil, errors.New("Unknow")
	}
}

func (hb *Huobi) ToSymbol(cp *CurrencyPair) string {
	return cp.ToSymbol("")
}

func (hb *Huobi) NormSymbol(cp *string) string {
	tmp := *cp
	return tmp[:3] + "_" + tmp[3:]
}

func (hb *Huobi) sendReq(method, path string,
	params map[string][]string, body map[string]string, sign bool) (int, *Json, error) {
	header := map[string][]string{
		"Content-Type": {`application/json`},
		"Accept":       {`application/json`},
		//"Accept-Language": {`zh-CN`},
		"User-Agent": {`Mozilla/5.0 (Windows NT 6.1; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/39.0.2171.71 Safari/537.36`},
	}

	req := &http.Request{
		Method: method,
		Header: header,
	}

	req.URL, _ = url.Parse("https://api.huobi.pro" + path)
	if body != nil {
		jsonBody, _ := json.Marshal(body)
		req.Body = ioutil.NopCloser(bytes.NewBuffer(jsonBody))
	}

	if sign {
		sign_params := map[string][]string{
			"AccessKeyId":      {hb.accesskeyid},
			"SignatureVersion": {`2`},
			"SignatureMethod":  {`HmacSHA256`},
			"Timestamp":        {time.Now().UTC().Format("2006-01-02T15:04:05")},
		}

		q := req.URL.Query()
		q = sign_params
		data := method + "\napi.huobi.pro\n" + path + "\n" + q.Encode()
		q.Add("Signature", ComputeHmac256Base64(data, hb.secretkeyid))
		req.URL.RawQuery = q.Encode()
	} else {
		q := req.URL.Query()
		q = params
		req.URL.RawQuery = q.Encode()
	}

	return recvResp(req)
}

func (hb *Huobi) GetAccount() (account string, err error) {
	status, js, err := hb.sendReq("GET", "/v1/account/accounts", nil, nil, true)
	if err != nil {
		return
	}

	respOk := func(js *Json) (interface{}, error) {
		status, _ := js.Get("status").String()
		if status == "ok" {
			acc, _ := js.Get("data").Array()
			for _, a := range acc {
				at := a.(map[string]interface{})
				account, _ := at["id"].(json.Number).Int64()
				return strconv.FormatInt(account, 10), nil
			}
		} else {
			reason, _ := js.Get("err-msg").String()
			return nil, errors.New(reason)
		}
		return nil, errors.New("Unknow")
	}

	acc, err := ProcessResp(status, js, respOk, hb.respErr)
	if err == nil {
		account = acc.(string)
	}
	return
}

func (hb *Huobi) GetBalance() (balances []Balance, err error) {
	var acc string
	if hb.account_id == "" {
		acc, err = hb.GetAccount()
		if err != nil {
			return
		}
		hb.account_id = acc
	}

	status, js, err := hb.sendReq("GET", "/v1/account/accounts/"+hb.account_id+"/balance", nil, nil, true)
	if err != nil {
		return
	}

	respOk := func(js *Json) (interface{}, error) {
		status, _ := js.Get("status").String()
		if status == "ok" {
			list, _ := js.Get("data").Get("list").Array()

			for _, l := range list {
				b := l.(map[string]interface{})
				if b["balance"].(string) != "0.000000000000000000" {
					balances = append(balances,
						Balance{Currency: b["currency"].(string),
							Balance: b["balance"].(string)})
				}
			}
			return balances, nil
		} else {
			reason, _ := js.Get("err-msg").String()
			err = errors.New(reason)
			return nil, err
		}
		return nil, errors.New("Unknow")
	}

	b, err := ProcessResp(status, js, respOk, hb.respErr)
	if err == nil {
		balances = b.([]Balance)
	}

	return
}

func (hb *Huobi) SetKey(access, secret string) {
	hb.accesskeyid = access
	hb.secretkeyid = secret
}

func (hb *Huobi) GetPrice(cp *CurrencyPair) (price Price, err error) {
	params := map[string][]string{
		"symbol": {hb.ToSymbol(cp)},
	}

	status, js, err := hb.sendReq("GET", "/market/trade", params, nil, false)
	if err != nil {
		return
	}

	respOk := func(js *Json) (interface{}, error) {
		status, _ := js.Get("status").String()
		if status == "ok" {
			data, _ := js.Get("tick").Get("data").Array()
			for _, d := range data {
				dd := d.(map[string]interface{})
				ddd, _ := dd["price"].(json.Number).Float64()
				return Price{ddd}, nil
			}
			return nil, nil
		} else {
			reason, _ := js.Get("err-msg").String()
			err = errors.New(reason)
			return nil, err
		}
		return nil, errors.New("Unknow")
	}

	p, err := ProcessResp(status, js, respOk, hb.respErr)
	if err == nil {
		price = p.(Price)
	}
	return
}

func (hb *Huobi) GetSymbols() (symbols []string, err error) {
	status, js, err := hb.sendReq("GET", "/v1/common/symbols", nil, nil, false)
	if err != nil {
		return
	}

	respOk := func(js *Json) (interface{}, error) {
		status, _ := js.Get("status").String()
		if status == "ok" {
			var s []string
			data, _ := js.Get("data").Array()
			for _, d := range data {
				dd := d.(map[string]interface{})
				base := dd["base-currency"].(string)
				quote := dd["quote-currency"].(string)
				s = append(s, base+"_"+quote)
			}
			return s, nil
		} else {
			reason, _ := js.Get("err-msg").String()
			err = errors.New(reason)
			return nil, err
		}
		return nil, errors.New("Unknow")
	}

	s, err := ProcessResp(status, js, respOk, hb.respErr)
	if err == nil {
		symbols = s.([]string)
	}
	return
}

func (hb *Huobi) GetDepth(cp *CurrencyPair) (depth Depth, err error) {
	params := map[string][]string{
		"symbol": {hb.ToSymbol(cp)},
		"type":   {"step0"},
	}

	status, js, err := hb.sendReq("GET", "/market/depth", params, nil, false)
	if err != nil {
		return
	}

	respOk := func(js *Json) (interface{}, error) {
		status, _ := js.Get("status").String()
		if status == "ok" {
			var depth Depth
			asks, _ := js.Get("tick").Get("asks").Array()
			for _, a := range asks {
				uu := a.([]interface{})
				price, _ := uu[0].(json.Number).Float64()
				amount, _ := uu[1].(json.Number).Float64()
				depth.Asks = append([]Unit{Unit{price, amount}}, depth.Asks...)
			}
			bids, _ := js.Get("tick").Get("bids").Array()
			for _, b := range bids {
				uu := b.([]interface{})
				price, _ := uu[0].(json.Number).Float64()
				amount, _ := uu[1].(json.Number).Float64()
				depth.Bids = append(depth.Bids, Unit{price, amount})
			}
			return depth, nil
		} else {
			reason, _ := js.Get("err-msg").String()
			err = errors.New(reason)
			return nil, err
		}
		return nil, errors.New("Unknow")
	}

	d, err := ProcessResp(status, js, respOk, hb.respErr)
	if err == nil {
		depth = d.(Depth)
	}
	return
}

func (hb *Huobi) OrderState(s interface{}) string {
	if s.(string) == "submitted" {
		return Alive
	} else if s.(string) == "canceled" {
		return Cancelled
	}
	return s.(string)
}

func (hb *Huobi) OrderSide(s string) string {
	side := strings.Split(s, "-")
	return side[0]
}

func (hb *Huobi) NewOrder(o *Order) (id string, err error) {
	var acc string
	if hb.account_id == "" {
		acc, err = hb.GetAccount()
		if err != nil {
			return
		}
		hb.account_id = acc
	}

	pb := map[string]string{
		"account-id": hb.account_id,
		"symbol":     hb.ToSymbol(&o.CP),
		"type":       o.Side + "-limit",
		"amount":     strconv.FormatFloat(o.Amount, 'f', -1, 64),
		"price":      strconv.FormatFloat(o.Price, 'f', -1, 64),
	}

	status, js, err := hb.sendReq("POST", "/v1/order/orders/place", nil, pb, true)
	if err != nil {
		return
	}

	respOk := func(js *Json) (interface{}, error) {
		status, _ := js.Get("status").String()
		if status == "ok" {
			id, _ = js.Get("data").String()
			return id, nil
		} else {
			reason, _ := js.Get("err-msg").String()
			err = errors.New(reason)
			return nil, err
		}
		return nil, errors.New("Unknow")
	}

	oid, err := ProcessResp(status, js, respOk, hb.respErr)
	if err == nil {
		id = oid.(string)
	}
	return
}

func (hb *Huobi) CancelOrder(o *Order) (err error) {
	status, js, err := hb.sendReq("POST", "/v1/order/orders/"+o.Id+"/submitcancel", nil, nil, true)
	if err != nil {
		return
	}

	respOk := func(js *Json) (interface{}, error) {
		status, _ := js.Get("status").String()
		if status == "ok" {
			return nil, nil
		} else {
			reason, _ := js.Get("err-msg").String()
			err = errors.New(reason)
			return nil, err
		}
		return nil, errors.New("Unknow")
	}

	_, err = ProcessResp(status, js, respOk, hb.respErr)
	return
}

func (hb *Huobi) QueryOrder(o *Order) (order Order, err error) {
	status, js, err := hb.sendReq("GET", "/v1/order/orders/"+o.Id, nil, nil, true)
	if err != nil {
		return
	}

	respOk := func(js *Json) (interface{}, error) {
		status, _ := js.Get("status").String()
		if status == "ok" {
			var order Order
			id, _ := js.Get("data").Get("id").Int64()
			order.Id = strconv.FormatInt(id, 10)
			symbol, _ := js.Get("data").Get("symbol").String()
			order.CP = NewCurrencyPair2(hb.NormSymbol(&symbol))
			side, _ := js.Get("data").Get("type").String()
			order.Side = hb.OrderSide(side)
			price, _ := js.Get("data").Get("price").String()
			order.Price, _ = strconv.ParseFloat(price, 64)
			amount, _ := js.Get("data").Get("amount").String()
			order.Amount, _ = strconv.ParseFloat(amount, 64)
			status, _ := js.Get("data").Get("state").String()
			order.State = hb.OrderState(status)
			executed, _ := js.Get("data").Get("field-amount").String()
			order.Executed, _ = strconv.ParseFloat(executed, 64)
			order.Remain = order.Amount - order.Executed
			return order, nil
		} else {
			reason, _ := js.Get("err-msg").String()
			err = errors.New(reason)
			return nil, err
		}
		return nil, errors.New("Unknow")
	}

	od, err := ProcessResp(status, js, respOk, hb.respErr)
	if err == nil {
		order = od.(Order)
	}
	return
}

func NewHuobi() Exchange {
	return new(Huobi)
}

func init() {
	RegisterEx("huobi", NewHuobi)
}
