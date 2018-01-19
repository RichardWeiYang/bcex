package lib

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"

	. "github.com/bitly/go-simplejson"
)

/*
 * Reference page: https://docs.bitfinex.com/v1/docs
 */

type Bitfinex struct {
	name                     string
	accesskeyid, secretkeyid string
}

var bitfinex = Bitfinex{name: "bitfinex"}

func (bf *Bitfinex) createReq(method, path string, sign bool) *http.Request {
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
		req.Header.Add("X-BFX-APIKEY", bitfinex.accesskeyid)
		req.Header.Add("X-BFX-PAYLOAD", payload_enc)
		req.Header.Add("X-BFX-SIGNATURE", GetParamHmacSha384Sign(bitfinex.secretkeyid, payload_enc))
	}
	return req
}

func (bf *Bitfinex) getResp(req *http.Request) (int, []byte) {
	client := &http.Client{}
	resp, _ := client.Do(req)
	body, _ := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	return resp.StatusCode, body
}

func (bf *Bitfinex) GetBalance() (balances []Balance, err error) {
	req := bf.createReq("POST", "/v1/balances", true)
	status, body := bf.getResp(req)
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
	req := bf.createReq("GET", "/v1/symbols", false)
	status, _ := bf.getResp(req)
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

func init() {
	RegisterEx(bitfinex.name, &bitfinex)
}
