package lib

import (
	"errors"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	. "github.com/bitly/go-simplejson"
)

/*
 * Reference page: https://www.okex.com/rest_api.html
 */

type Okex struct {
	accesskeyid, secretkeyid string
}

func (ok *Okex) respErr(js *Json) (interface{}, error) {
	return nil, errors.New("Unknow")
}

func (ok *Okex) sendReq(method, path string,
	params map[string][]string, sign bool) (int, []byte) {
	header := map[string][]string{
		"Content-Type": {`application/x-www-form-urlencoded`},
	}

	req := &http.Request{
		Method: method,
		Header: header,
	}

	req.URL, _ = url.Parse("https://www.okex.com" + path)
	if sign {
		sign_params := map[string][]string{
			"api_key": {ok.accesskeyid},
		}

		for k, v := range params {
			sign_params[k] = v
		}

		q := req.URL.Query()
		q = sign_params
		//q2 := q
		//q2.Add("secret_key", okex.secretkeyid)
		data := q.Encode() + "&secret_key=" + ok.secretkeyid
		q.Add("sign", strings.ToUpper(GetMD5Hash(data)))
		req.URL.RawQuery = q.Encode()
	} else {
		q := req.URL.Query()
		q = params
		req.URL.RawQuery = q.Encode()
	}
	return recvResp(req)
}

func (ok *Okex) GetBalance() (balances []Balance, err error) {
	status, body := ok.sendReq("POST", "/api/v1/userinfo.do", nil, true)
	js, _ := NewJson(body)

	respOk := func(js *Json) (interface{}, error) {
		result, _ := js.Get("result").Bool()
		if result {
			free, _ := js.Get("info").Get("funds").Get("free").Map()
			for cur, b := range free {
				if b.(string) != "0" {
					balances = append(balances,
						Balance{cur, b.(string)})
				}
			}
			return balances, nil
		} else {
			reason, _ := js.Get("error_code").Int64()
			err = errors.New(strconv.FormatInt(reason, 10))
			return nil, err
		}
		return nil, errors.New("Unknow")
	}

	b, err := ProcessResp(status, js, respOk, ok.respErr)
	if err == nil {
		balances = b.([]Balance)
	}

	return
}

func (ok *Okex) Alive() bool {
	status, _ := ok.sendReq("GET", "/api/v1/exchange_rate.do", nil, false)
	_, err := ProcessResp(status, nil, isAlive, notAlive)

	if err != nil {
		return true
	} else {
		return false
	}
}

func (ok *Okex) SetKey(access, secret string) {
	ok.accesskeyid = access
	ok.secretkeyid = secret
}

func (ok *Okex) GetPrice(cp *CurrencyPair) (price Price, err error) {
	params := map[string][]string{
		"symbol": {cp.ToSymbol("_")},
	}

	status, body := ok.sendReq("GET", "/api/v1/ticker.do", params, false)
	js, _ := NewJson(body)

	respOk := func(js *Json) (interface{}, error) {
		reason, e := js.Get("error_code").Int64()
		if e == nil {
			err = errors.New(strconv.FormatInt(reason, 10))
			return nil, err
		}

		last_s, _ := js.Get("ticker").Get("last").String()
		last, _ := strconv.ParseFloat(last_s, 64)
		return Price{last}, nil
	}

	p, err := ProcessResp(status, js, respOk, ok.respErr)
	if err == nil {
		price = p.(Price)
	}
	return
}

func (ok *Okex) GetSymbols() (symbols []string, err error) {
	status, body := ok.sendReq("GET", "/v2/markets/products", nil, false)
	js, _ := NewJson(body)

	respOk := func(js *Json) (interface{}, error) {
		reason, e := js.Get("error_code").Int64()
		if e == nil {
			err = errors.New(strconv.FormatInt(reason, 10))
			return nil, err
		}

		var s []string
		data, _ := js.Get("data").Array()
		for _, d := range data {
			dd := d.(map[string]interface{})
			symbol := dd["symbol"].(string)
			s = append(s, symbol)
		}
		return s, nil
	}

	s, err := ProcessResp(status, js, respOk, ok.respErr)
	if err == nil {
		symbols = s.([]string)
	}
	return
}

func NewOkex() Exchange {
	return new(Okex)
}

func init() {
	RegisterEx("okex", NewOkex)
}
