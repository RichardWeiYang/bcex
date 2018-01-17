package lib

import (
	"crypto/md5"
	"encoding/hex"
	"errors"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/bitly/go-simplejson"
)

/*
 * Reference page: https://www.okex.com/rest_api.html
 */

type Okex struct {
	name                     string
	accesskeyid, secretkeyid string
}

var okex = Okex{name: "okex"}

func GetMD5Hash(text string) string {
	hasher := md5.New()
	hasher.Write([]byte(text))
	return hex.EncodeToString(hasher.Sum(nil))
}

func (ok *Okex) createReq(method, path string, sign bool) *http.Request {
	header := map[string][]string{
		"Content-Type": {`application/x-www-form-urlencoded`},
	}

	req := &http.Request{
		Method: method,
		Header: header,
	}

	req.URL, _ = url.Parse("https://www.okex.com" + path)
	params := map[string][]string{
		"api_key": {okex.accesskeyid},
	}

	if sign {
		q := req.URL.Query()
		q = params
		//q2 := q
		//q2.Add("secret_key", okex.secretkeyid)
		data := q.Encode() + "&secret_key=" + okex.secretkeyid
		q.Add("sign", strings.ToUpper(GetMD5Hash(data)))
		req.URL.RawQuery = q.Encode()
	}
	return req
}

func (ok *Okex) getResp(req *http.Request) (int, []byte) {
	client := &http.Client{}
	resp, _ := client.Do(req)
	body, _ := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	return resp.StatusCode, body
}

func (ok *Okex) GetBalance() (balances []Balance, err error) {
	req := ok.createReq("POST", "/api/v1/userinfo.do", true)
	status, body := ok.getResp(req)
	err = nil
	if status == http.StatusOK {
		js, _ := simplejson.NewJson(body)
		result, _ := js.Get("result").Bool()
		if result {
			free, _ := js.Get("info").Get("funds").Get("free").Map()
			for cur, b := range free {
				if b.(string) != "0" {
					balances = append(balances,
						Balance{cur, b.(string)})
				}
			}
		} else {
			reason, _ := js.Get("error_code").Int64()
			err = errors.New(strconv.FormatInt(reason, 10))
		}
		return
	}
	err = errors.New("Http Error")
	return
}

func (ok *Okex) Alive() bool {
	req := ok.createReq("GET", "/api/v1/exchange_rate.do", false)
	status, _ := ok.getResp(req)
	if status == http.StatusOK {
		return true
	}
	return false
}

func init() {
	readConf()
	okex.accesskeyid = keys[okex.name].AccessKeyId
	okex.secretkeyid = keys[okex.name].SecretKeyId
	if okex.Alive() {
		RegisterEx(okex.name, &okex)
	}
}
