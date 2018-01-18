package lib

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"time"

	. "github.com/bitly/go-simplejson"
)

/*
 * Reference page: https://github.com/huobiapi/API_Docs/wiki
 */

type Huobi struct {
	name                     string
	accesskeyid, secretkeyid string
	account_id               string
}

var huobi = Huobi{name: "huobi", account_id: ""}

func (hb *Huobi) createReq(method, path string) *http.Request {
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
	params := map[string][]string{
		"AccessKeyId":      {huobi.accesskeyid},
		"SignatureVersion": {`2`},
		"SignatureMethod":  {`HmacSHA256`},
		"Timestamp":        {time.Now().UTC().Format("2006-01-02T15:04:05")},
	}

	q := req.URL.Query()
	q = params
	data := "GET\napi.huobi.pro\n" + path + "\n" + q.Encode()
	q.Add("Signature", ComputeHmac256Base64(data, huobi.secretkeyid))
	req.URL.RawQuery = q.Encode()
	return req
}

func (hb *Huobi) getResp(req *http.Request) (int, []byte) {
	client := &http.Client{}
	resp, _ := client.Do(req)
	body, _ := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	return resp.StatusCode, body
}

func (hb *Huobi) GetAccount() (account string, err error) {
	req := hb.createReq("GET", "/v1/account/accounts")
	status, body := hb.getResp(req)
	js, _ := NewJson(body)

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

	respErr := func(js *Json) (interface{}, error) {
		return nil, errors.New("Unknow")
	}

	acc, err := ProcessResp(status, js, respOk, respErr)
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

	req := hb.createReq("GET", "/v1/account/accounts/"+hb.account_id+"/balance")
	status, body := hb.getResp(req)
	js, _ := NewJson(body)

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

	respErr := func(js *Json) (interface{}, error) {
		return nil, errors.New("Unknow")
	}

	b, err := ProcessResp(status, js, respOk, respErr)
	if err == nil {
		balances = b.([]Balance)
	}

	return
}

func (hb *Huobi) Alive() bool {
	req := hb.createReq("GET", "/v1/common/timestamp")
	status, _ := hb.getResp(req)
	_, err := ProcessResp(status, nil, isAlive, notAlive)

	if err != nil {
		return true
	} else {
		return false
	}
}

func init() {
	readConf()
	huobi.accesskeyid = keys[huobi.name].AccessKeyId
	huobi.secretkeyid = keys[huobi.name].SecretKeyId
	RegisterEx(huobi.name, huobi.accesskeyid, huobi.secretkeyid, &huobi)
}
