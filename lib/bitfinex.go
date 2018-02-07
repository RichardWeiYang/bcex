package lib

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	. "github.com/bitly/go-simplejson"
)

/*
 * Reference page: https://docs.bitfinex.com/v1/docs
 */

type Bitfinex struct {
	accesskeyid, secretkeyid string
}

func (bf *Bitfinex) respErr(js *Json) (interface{}, error) {
	reason, _ := js.Get("message").String()
	err := errors.New(reason)
	return nil, err
}

func (bf *Bitfinex) ToSymbol(cp *CurrencyPair) string {
	return strings.ToUpper(cp.ToSymbol(""))
}

func (bf *Bitfinex) NormSymbol(cp *string) string {
	tmp := *cp
	return tmp[:3] + "_" + tmp[3:]
}

func (bf *Bitfinex) sendReq(method, path string,
	params map[string]interface{}, sign bool) (int, []byte) {
	header := map[string][]string{
		"Content-Type": {`application/json`},
		"Accept":       {`application/json`},
	}

	req := &http.Request{
		Method: method,
		Header: header,
	}

	req.URL, _ = url.Parse("https://api.bitfinex.com" + path)

	if sign {
		payload := map[string]interface{}{
			"request": path,
			"nonce":   fmt.Sprintf("%v", time.Now().Unix()*10000),
		}

		for k, v := range params {
			payload[k] = v
		}

		payload_json, _ := json.Marshal(payload)
		payload_enc := base64.StdEncoding.EncodeToString(payload_json)

		req.Header.Add("Content-Type", "application/json")
		req.Header.Add("Accept", "application/json")
		req.Header.Add("X-BFX-APIKEY", bf.accesskeyid)
		req.Header.Add("X-BFX-PAYLOAD", payload_enc)
		req.Header.Add("X-BFX-SIGNATURE", GetParamHmacSha384Sign(bf.secretkeyid, payload_enc))
	}
	return recvResp(req)
}

func (bf *Bitfinex) GetBalance() (balances []Balance, err error) {
	status, body := bf.sendReq("POST", "/v1/balances", nil, true)
	js, _ := NewJson(body)

	respOk := func(js *Json) (interface{}, error) {
		bs, _ := js.Array()
		for _, b := range bs {
			bt := b.(map[string]interface{})
			balances = append(balances,
				Balance{Currency: bt["currency"].(string),
					Balance: bt["amount"].(string)})
		}
		return balances, nil
	}

	b, err := ProcessResp(status, js, respOk, bf.respErr)
	if err == nil {
		balances = b.([]Balance)
	}
	return
}
func (bf *Bitfinex) Alive() bool {
	status, _ := bf.sendReq("GET", "/v1/symbols", nil, false)
	_, err := ProcessResp(status, nil, isAlive, notAlive)

	if err != nil {
		return true
	} else {
		return false
	}
}

func (bf *Bitfinex) SetKey(access, secret string) {
	bf.accesskeyid = access
	bf.secretkeyid = secret
}

func (bf *Bitfinex) GetPrice(cp *CurrencyPair) (price Price, err error) {
	status, body := bf.sendReq("GET", "/v1/pubticker/"+bf.ToSymbol(cp), nil, false)
	js, _ := NewJson(body)

	respOk := func(js *Json) (interface{}, error) {
		price_s, _ := js.Get("last_price").String()
		price, _ := strconv.ParseFloat(price_s, 64)
		return Price{price}, nil
	}

	p, err := ProcessResp(status, js, respOk, bf.respErr)
	if err == nil {
		price = p.(Price)
	}
	return
}

func (bf *Bitfinex) GetSymbols() (symbols []string, err error) {
	status, body := bf.sendReq("GET", "/v1/symbols", nil, false)
	js, _ := NewJson(body)

	respOk := func(js *Json) (interface{}, error) {
		var s []string
		data, _ := js.Array()
		for _, d := range data {
			base := d.(string)[0:3]
			quote := d.(string)[3:]
			s = append(s, base+"_"+quote)
		}
		return s, nil
	}

	s, err := ProcessResp(status, js, respOk, bf.respErr)
	if err == nil {
		symbols = s.([]string)
	}
	return
}

func (bf *Bitfinex) GetDepth(cp *CurrencyPair) (depth Depth, err error) {
	status, body := bf.sendReq("GET", "/v1/book/"+bf.ToSymbol(cp), nil, false)
	js, _ := NewJson(body)

	respOk := func(js *Json) (interface{}, error) {
		var depth Depth
		asks, _ := js.Get("asks").Array()
		for _, a := range asks {
			uu := a.(map[string]interface{})
			price, _ := strconv.ParseFloat(uu["price"].(string), 64)
			amount, _ := strconv.ParseFloat(uu["amount"].(string), 64)
			depth.Asks = append([]Unit{Unit{price, amount}}, depth.Asks...)
		}
		bids, _ := js.Get("bids").Array()
		for _, b := range bids {
			uu := b.(map[string]interface{})
			price, _ := strconv.ParseFloat(uu["price"].(string), 64)
			amount, _ := strconv.ParseFloat(uu["amount"].(string), 64)
			depth.Bids = append(depth.Bids, Unit{price, amount})
		}
		return depth, nil
	}

	d, err := ProcessResp(status, js, respOk, bf.respErr)
	if err == nil {
		depth = d.(Depth)
	}
	return
}

func (bf *Bitfinex) OrderState(s interface{}) string {
	if s.(bool) {
		return Cancelled
	}
	return Alive
}

func (bf *Bitfinex) OrderSide(s string) string {
	return s
}

func (bf *Bitfinex) NewOrder(o *Order) (id string, err error) {
	params := map[string]interface{}{
		"symbol":   bf.ToSymbol(&o.CP),
		"side":     o.Side,
		"amount":   strconv.FormatFloat(o.Amount, 'f', -1, 64),
		"price":    strconv.FormatFloat(o.Price, 'f', -1, 64),
		"type":     "exchange limit",
		"exchange": "bitfinex",
	}

	status, body := bf.sendReq("POST", "/v1/order/new", params, true)
	js, _ := NewJson(body)

	respOk := func(js *Json) (interface{}, error) {
		id, _ := js.Get("id").Int64()
		return strconv.FormatInt(id, 10), nil
	}

	oid, err := ProcessResp(status, js, respOk, bf.respErr)
	if err == nil {
		id = oid.(string)
	}
	return
}

func (bf *Bitfinex) CancelOrder(o *Order) (err error) {
	id, _ := strconv.ParseInt(o.Id, 10, 64)
	params := map[string]interface{}{
		"order_id": id,
	}

	status, body := bf.sendReq("POST", "/v1/order/cancel", params, true)
	js, _ := NewJson(body)

	respOk := func(js *Json) (interface{}, error) {
		return nil, nil
	}

	_, err = ProcessResp(status, js, respOk, bf.respErr)
	return
}

func NewBitfinex() Exchange {
	return new(Bitfinex)
}

func init() {
	RegisterEx("bitfinex", NewBitfinex)
}
