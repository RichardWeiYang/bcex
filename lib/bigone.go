package lib

import (
	"errors"
	"io/ioutil"
	"net/http"
	"net/url"

	. "github.com/bitly/go-simplejson"
)

/*
 * Reference page: https://developer.big.one
 */

type BigOne struct {
	name                     string
	accesskeyid, secretkeyid string
}

var bigone = BigOne{name: "bigone"}

func (bo *BigOne) createReq(method, path string) *http.Request {
	header := map[string][]string{
		"Authorization": {"Bearer " + bigone.accesskeyid},
		"User-Agent":    {`standard browser user agent format`},
		"Big-Device-Id": {bigone.secretkeyid},
	}

	req := &http.Request{
		Method: method,
		Header: header,
	}

	req.URL, _ = url.Parse("https://api.big.one" + path)
	return req
}

func (bo *BigOne) getResp(req *http.Request) (int, []byte) {
	client := &http.Client{}
	resp, _ := client.Do(req)
	body, _ := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	return resp.StatusCode, body
}

func (bo *BigOne) GetBalance() (balances []Balance, err error) {
	req := bo.createReq("GET", "/accounts")
	status, body := bo.getResp(req)
	js, _ := NewJson(body)

	respOk := func(js *Json) (interface{}, error) {
		bs, _ := js.Get("data").Array()
		for _, b := range bs {
			bt := b.(map[string]interface{})
			if bt["active_balance"].(string) != "0.00000000" {
				balances = append(balances,
					Balance{Currency: bt["account_type"].(string),
						Balance: bt["active_balance"].(string)})
			}
		}
		return balances, nil
	}

	respErr := func(js *Json) (interface{}, error) {
		reason, _ := js.Get("error").Get("description").String()
		err = errors.New(reason)
		return nil, err
	}

	b, err := ProcessResp(status, js, respOk, respErr)
	if err == nil {
		balances = b.([]Balance)
	}

	return
}

func (bo *BigOne) Alive() bool {
	req := bo.createReq("GET", "/accounts")
	status, _ := bo.getResp(req)

	_, err := ProcessResp(status, nil, isAlive, notAlive)

	if err != nil {
		return true
	} else {
		return false
	}
}

func init() {
	readConf()
	bigone.accesskeyid = keys[bigone.name].AccessKeyId
	bigone.secretkeyid = GetUUID()
	RegisterEx(bigone.name, bigone.accesskeyid, bigone.secretkeyid, &bigone)
}
