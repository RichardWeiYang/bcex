package lib

import (
	"errors"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/bitly/go-simplejson"
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
	js, _ := simplejson.NewJson(body)
	if status == http.StatusOK {
		bs, _ := js.Get("balances").Array()
		for _, b := range bs {
			bt := b.(map[string]interface{})
			if bt["free"].(string) != "0.00000000" {
				balances = append(balances,
					Balance{Currency: bt["asset"].(string),
						Balance: bt["free"].(string)})
			}
		}
	} else {
		reason, _ := js.Get("msg").String()
		err = errors.New(reason)
	}
	return
}
func (bn *Binance) Alive() bool {
	req := bn.createReq("GET", "/api/v1/time", nil, false)
	status, _ := bn.getResp(req)
	if status == http.StatusOK {
		return true
	}
	return false
}
func init() {
	readConf()
	binance.accesskeyid = keys[binance.name].AccessKeyId
	binance.secretkeyid = keys[binance.name].SecretKeyId
	if binance.Alive() {
		RegisterEx(binance.name, &binance)
	}
}
