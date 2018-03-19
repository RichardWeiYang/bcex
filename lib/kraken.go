package lib

import (
	"net/http"
	"net/url"
	"strings"

	. "github.com/bitly/go-simplejson"
)

type Kraken struct {
	accesskeyid, secretkeyid string
}

func (kk *Kraken) respErr(js *Json) (interface{}, error) {
	return nil, nil
}

func (kk *Kraken) ToSymbol(cp *CurrencyPair) string {
	return cp.ToSymbol("_")
}

func (kk *Kraken) NormSymbol(cp *string) string {
	return *cp
}

func (kk *Kraken) sendReq(method, path string,
	params map[string][]string, sign bool) (int, *Json, error) {
	header := map[string][]string{
		"Content-Type": {`application/x-www-form-urlencoded`},
	}

	req := &http.Request{
		Method: method,
		Header: header,
	}

	req.URL, _ = url.Parse("https://api.kraken.com" + path)
	if sign {
		sign_params := map[string][]string{
			"ex_key": {kk.accesskeyid},
		}

		for k, v := range params {
			sign_params[k] = v
		}

		q := req.URL.Query()
		q = sign_params
		data := q.Encode() + "&secret_key=" + kk.secretkeyid
		q.Add("sign", strings.ToUpper(GetMD5Hash(data)))
		req.URL.RawQuery = q.Encode()
	} else {
		q := req.URL.Query()
		q = params
		req.URL.RawQuery = q.Encode()
	}
	return recvResp(req)
}

func (kk *Kraken) SetKey(access, secret string) {
	kk.accesskeyid = access
	kk.secretkeyid = secret
}

func (kk *Kraken) GetBalance() (balances []Balance, err error) {
	return
}

func (kk *Kraken) GetPrice(cp *CurrencyPair) (price Price, err error) {
	return
}

func (kk *Kraken) GetSymbols() (symbols []string, err error) {
	status, js, err := kk.sendReq("GET", "/0/public/AssetPairs", nil, false)
	if err != nil {
		return
	}

	respOk := func(js *Json) (interface{}, error) {
		var s []string
		data, _ := js.Get("result").Map()
		for _, d := range data {
			dd := d.(map[string]interface{})
			base := strings.ToLower(dd["base"].(string))
			quote := strings.ToLower(dd["quote"].(string))
			s = append(s, base+"_"+quote)
		}
		return s, nil
	}

	s, err := ProcessResp(status, js, respOk, kk.respErr)
	if err == nil {
		symbols = s.([]string)
	}
	return
}

func (kk *Kraken) GetDepth(cp *CurrencyPair) (depth Depth, err error) {
	return
}

func (kk *Kraken) OrderState(s interface{}) string {
	return s.(string)
}

func (kk *Kraken) OrderSide(s string) string {
	return s
}

func (kk *Kraken) NewOrder(o *Order) (id string, err error) {
	return
}

func (kk *Kraken) CancelOrder(o *Order) (err error) {
	return
}

func (kk *Kraken) QueryOrder(o *Order) (order Order, err error) {
	return
}

func NewKraken() Exchange {
	return new(Kraken)
}

func init() {
	RegisterEx("kraken", NewKraken)
}
