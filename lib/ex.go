package lib

import (
	"net/http"
	"net/url"
	"strings"

	. "github.com/bitly/go-simplejson"
)

type Ex struct {
	accesskeyid, secretkeyid string
}

func (exe *Ex) respErr(js *Json) (interface{}, error) {
	return nil, nil
}

func (exe *Ex) ToSymbol(cp *CurrencyPair) string {
	return cp.ToSymbol("_")
}

func (exe *Ex) sendReq(method, path string,
	params map[string][]string, sign bool) (int, []byte) {
	header := map[string][]string{
		"Content-Type": {`application/x-www-form-urlencoded`},
	}

	req := &http.Request{
		Method: method,
		Header: header,
	}

	req.URL, _ = url.Parse("https://www.ex.com" + path)
	if sign {
		sign_params := map[string][]string{
			"ex_key": {exe.accesskeyid},
		}

		for k, v := range params {
			sign_params[k] = v
		}

		q := req.URL.Query()
		q = sign_params
		data := q.Encode() + "&secret_key=" + exe.secretkeyid
		q.Add("sign", strings.ToUpper(GetMD5Hash(data)))
		req.URL.RawQuery = q.Encode()
	} else {
		q := req.URL.Query()
		q = params
		req.URL.RawQuery = q.Encode()
	}
	return recvResp(req)
}

func (exe *Ex) SetKey(access, secret string) {
	exe.accesskeyid = access
	exe.secretkeyid = secret
}

func (exe *Ex) Alive() bool {
	return true
}

func (exe *Ex) GetBalance() (balances []Balance, err error) {
	return
}

func (exe *Ex) GetPrice(cp *CurrencyPair) (price Price, err error) {
	return
}

func (exe *Ex) GetSymbols() (symbols []string, err error) {
	return
}

func (exe *Ex) GetDepth(cp *CurrencyPair) (depth Depth, err error) {
	return
}

func NewEx() Exchange {
	return new(Ex)
}

func init() {
	RegisterEx("exe", nil)
}
