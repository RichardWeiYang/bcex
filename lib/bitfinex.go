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

func (bf *Bitfinex) sendReq(method, path string, sign bool) (int, []byte) {
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
	status, body := bf.sendReq("POST", "/v1/balances", true)
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

	respErr := func(js *Json) (interface{}, error) {
		reason, _ := js.Get("message").String()
		err = errors.New(reason)
		return nil, err
	}

	b, err := ProcessResp(status, js, respOk, respErr)
	if err == nil {
		balances = b.([]Balance)
	}
	return
}
func (bf *Bitfinex) Alive() bool {
	status, _ := bf.sendReq("GET", "/v1/symbols", false)
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
	status, body := bf.sendReq("GET", "/v1/pubticker/"+strings.ToUpper(cp.ToSymbol("")), false)
	js, _ := NewJson(body)

	respOk := func(js *Json) (interface{}, error) {
		price_s, _ := js.Get("last_price").String()
		price, _ := strconv.ParseFloat(price_s, 64)
		return Price{price}, nil
	}

	respErr := func(js *Json) (interface{}, error) {
		reason, _ := js.Get("message").String()
		err = errors.New(reason)
		return nil, err
	}

	p, err := ProcessResp(status, js, respOk, respErr)
	if err == nil {
		price = p.(Price)
	}
	return
}

func NewBitfinex() Exchange {
	return new(Bitfinex)
}

func init() {
	RegisterEx("bitfinex", NewBitfinex)
}
