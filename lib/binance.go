package lib

import (
	"errors"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"time"

	. "github.com/bitly/go-simplejson"
)

/*
 * Reference page: https://github.com/binance-exchange/binance-official-api-docs/blob/master/rest-api.md
 */

type Binance struct {
	name                     string
	accesskeyid, secretkeyid string
}

var binance = Binance{name: "binance"}

func (bn *Binance) createReq(method, path string,
	params map[string][]string, sign bool) *http.Request {
	header := map[string][]string{
		"Content-Type": {`application/x-www-form-urlencoded`},
	}

	req := &http.Request{
		Method: method,
		Header: header,
	}

	req.URL, _ = url.Parse("https://api.binance.com" + path)

	if sign {
		q := req.URL.Query()
		q = params
		q.Add("signature", ComputeHmac256(q.Encode(), binance.secretkeyid))
		req.URL.RawQuery = q.Encode()
		req.Header.Add("X-MBX-APIKEY", binance.accesskeyid)
	}
	return req
}

func (bn *Binance) getResp(req *http.Request) (int, []byte) {
	client := &http.Client{}
	resp, _ := client.Do(req)
	body, _ := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	return resp.StatusCode, body
}

func (bn *Binance) GetBalance() (balances []Balance, err error) {
	params := map[string][]string{
		"recvWindow": {`5000`},
		"timestamp":  {strconv.FormatInt(time.Now().UnixNano(), 10)[0:13]},
	}
	req := bn.createReq("GET", "/api/v3/account", params, true)
	status, body := bn.getResp(req)
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

	respErr := func(js *Json) (interface{}, error) {
		reason, _ := js.Get("msg").String()
		err = errors.New(reason)
		return nil, err
	}

	b, err := ProcessResp(status, js, respOk, respErr)
	if err == nil {
		balances = b.([]Balance)
	}

	return
}
func (bn *Binance) Alive() bool {
	req := bn.createReq("GET", "/api/v1/time", nil, false)
	status, _ := bn.getResp(req)
	_, err := ProcessResp(status, nil, isAlive, notAlive)

	if err != nil {
		return true
	} else {
		return false
	}
}
func init() {
	readConf()
	binance.accesskeyid = keys[binance.name].AccessKeyId
	binance.secretkeyid = keys[binance.name].SecretKeyId
	RegisterEx(binance.name, binance.accesskeyid, binance.secretkeyid, &binance)
}
