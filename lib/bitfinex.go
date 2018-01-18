package lib

import (
	"crypto/hmac"
	"crypto/sha512"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"

	"github.com/bitly/go-simplejson"
)

/*
 * Reference page: https://docs.bitfinex.com/v1/docs
 */

type Bitfinex struct {
	name                     string
	accesskeyid, secretkeyid string
}

var bitfinex = Bitfinex{name: "bitfinex"}

func (bf *Bitfinex) createReq(method, path string, sign bool) *http.Request {
	header := map[string][]string{
		"Content-Type": {`application/json`},
		"Accept":       {`application/json`},
	}

	req := &http.Request{
		Method: method,
		Header: header,
	}

	req.URL, _ = url.Parse("https://api.bitfinex.com" + path)

	if sign {
		payload := map[string]interface{}{
			"request": path,
			"nonce":   fmt.Sprintf("%v", time.Now().Unix()*10000),
		}
		payload_json, _ := json.Marshal(payload)
		payload_enc := base64.StdEncoding.EncodeToString(payload_json)

		sig := hmac.New(sha512.New384, []byte(bitfinex.secretkeyid))
		sig.Write([]byte(payload_enc))
		req.Header.Add("Content-Type", "application/json")
		req.Header.Add("Accept", "application/json")
		req.Header.Add("X-BFX-APIKEY", bitfinex.accesskeyid)
		req.Header.Add("X-BFX-PAYLOAD", payload_enc)
		req.Header.Add("X-BFX-SIGNATURE", hex.EncodeToString(sig.Sum(nil)))
	}
	return req
}

func (bf *Bitfinex) getResp(req *http.Request) (int, []byte) {
	client := &http.Client{}
	resp, _ := client.Do(req)
	body, _ := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	return resp.StatusCode, body
}

func (bf *Bitfinex) GetBalance() (balances []Balance, err error) {
	req := bf.createReq("POST", "/v1/balances", true)
	status, body := bf.getResp(req)
	if status == http.StatusOK {
		js, _ := simplejson.NewJson(body)
		bs, _ := js.Array()
		for _, b := range bs {
			bt := b.(map[string]interface{})
			balances = append(balances,
				Balance{Currency: bt["currency"].(string),
					Balance: bt["amount"].(string)})
		}
		return
	}
	return
}
func (bf *Bitfinex) Alive() bool {
	req := bf.createReq("GET", "/v1/symbols", false)
	status, _ := bf.getResp(req)
	if status == http.StatusOK {
		return true
	}
	return false
}

func init() {
	readConf()
	bitfinex.accesskeyid = keys[bitfinex.name].AccessKeyId
	bitfinex.secretkeyid = keys[bitfinex.name].SecretKeyId
	if bitfinex.Alive() {
		RegisterEx(bitfinex.name, &bitfinex)
	}
}
