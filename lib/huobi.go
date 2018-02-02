package lib

import (
	"encoding/json"
	"errors"
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
	accesskeyid, secretkeyid string
	account_id               string
}

func (hb *Huobi) respErr(js *Json) (interface{}, error) {
	return nil, errors.New("Unknow")
}

func (hb *Huobi) sendReq(method, path string,
	params map[string][]string, sign bool) (int, []byte) {
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

	if sign {
		sign_params := map[string][]string{
			"AccessKeyId":      {hb.accesskeyid},
			"SignatureVersion": {`2`},
			"SignatureMethod":  {`HmacSHA256`},
			"Timestamp":        {time.Now().UTC().Format("2006-01-02T15:04:05")},
		}

		for k, v := range params {
			sign_params[k] = v
		}

		q := req.URL.Query()
		q = sign_params
		data := "GET\napi.huobi.pro\n" + path + "\n" + q.Encode()
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
	status, body := hb.sendReq("GET", "/v1/account/accounts", nil, true)
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

	status, body := hb.sendReq("GET", "/v1/account/accounts/"+hb.account_id+"/balance", nil, true)
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

	b, err := ProcessResp(status, js, respOk, hb.respErr)
	if err == nil {
		balances = b.([]Balance)
	}

	return
}

func (hb *Huobi) Alive() bool {
	status, _ := hb.sendReq("GET", "/v1/common/timestamp", nil, false)
	_, err := ProcessResp(status, nil, isAlive, notAlive)

	if err != nil {
		return true
	} else {
		return false
	}
}

func (hb *Huobi) SetKey(access, secret string) {
	hb.accesskeyid = access
	hb.secretkeyid = secret
}

func (hb *Huobi) GetPrice(cp *CurrencyPair) (price Price, err error) {
	params := map[string][]string{
		"symbol": {cp.ToSymbol("")},
	}

	status, body := hb.sendReq("GET", "/market/trade", params, false)
	js, _ := NewJson(body)

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
	status, body := hb.sendReq("GET", "/v1/common/symbols", nil, false)
	js, _ := NewJson(body)

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

func NewHuobi() Exchange {
	return new(Huobi)
}

func init() {
	RegisterEx("huobi", NewHuobi)
}
