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
 * Reference page: https://developer.big.one
 */

type BigOne struct {
	accesskeyid, secretkeyid string
}

func (bo *BigOne) sendReq(method, path string) (int, []byte) {
	header := map[string][]string{
		"Authorization": {"Bearer " + bo.accesskeyid},
		"User-Agent":    {`standard browser user agent format`},
		"Big-Device-Id": {bo.secretkeyid},
	}

	req := &http.Request{
		Method: method,
		Header: header,
	}

	req.URL, _ = url.Parse("https://api.big.one" + path)
	return recvResp(req)
}

func (bo *BigOne) GetBalance() (balances []Balance, err error) {
	status, body := bo.sendReq("GET", "/accounts")
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
	status, _ := bo.sendReq("GET", "/accounts")

	_, err := ProcessResp(status, nil, isAlive, notAlive)

	if err != nil {
		return true
	} else {
		return false
	}
}

func (bo *BigOne) SetKey(access, secret string) {
	bo.accesskeyid = access
	bo.secretkeyid = GetUUID()
}

func (bo *BigOne) GetPrice(cp *CurrencyPair) (price Price, err error) {
	status, body := bo.sendReq("GET", "/markets/"+strings.ToUpper(cp.ToSymbol("-")))
	js, _ := NewJson(body)

	respOk := func(js *Json) (interface{}, error) {
		price_s, _ := js.Get("data").Get("ticker").Get("price").String()
		price, _ := strconv.ParseFloat(price_s, 64)
		return Price{price}, nil
	}

	respErr := func(js *Json) (interface{}, error) {
		reason, _ := js.Get("error").Get("description").String()
		err = errors.New(reason)
		return nil, err
	}

	p, err := ProcessResp(status, js, respOk, respErr)
	if err == nil {
		price = p.(Price)
	}

	return
}

func NewBigOne() Exchange {
	return new(BigOne)
}

func init() {
	RegisterEx("bigone", NewBigOne)
}
