package lib

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	. "github.com/bitly/go-simplejson"
)

type RespHandle func(js *Json) (interface{}, error)

type ExchangeKey struct {
	AccessKeyId string `json:"AccessKeyId"`
	SecretKeyId string `json:"SecretKeyId"`
}

var keys map[string]ExchangeKey
var initialized = false
var privkey = ""

func Init(bcexkey string) {
	length := len(bcexkey)
	if length == 0 {
		fmt.Println("BCEX Key is needed, use -k or ENV to set it")
		os.Exit(1)
	}

	if length != 16 && length != 24 && length != 32 {
		fmt.Printf("BCEX Key must be 16, 24 or 32 bytes, current is %d\n", length)
		os.Exit(1)
	}

	privkey = bcexkey

	ReadConf()

	for name, e := range ex {
		ek := keys[name]
		e.SetKey(ek.AccessKeyId, ek.SecretKeyId)
	}
}

func WriteConf(name, accesskey, secretkey string) {
	var exKey ExchangeKey

	ex := GetEx(name)
	if ex == nil {
		fmt.Println("Exchange: (", name, ") is not found, use command list to show supported exchagnes name")
		return
	}

	if keys == nil {
		keys = make(map[string]ExchangeKey)
	}
	exKey.AccessKeyId = accesskey
	exKey.SecretKeyId = secretkey
	keys[name] = exKey

	plain, _ := json.Marshal(keys)
	raw := encrypt(string(plain), privkey)
	ioutil.WriteFile("config.json", []byte(raw), 0644)
}

func ReadConf() {
	raw, err := ioutil.ReadFile("config.json")
	if err != nil {
		return
	}

	plain := decrypt(string(raw), privkey)

	json.Unmarshal([]byte(plain), &keys)
}

type Balance struct {
	Currency string
	Balance  string
}

type Price struct {
	Price float64
}

type Exchange interface {
	SetKey(access, secret string)
	Alive() bool
	GetBalance() ([]Balance, error)
	GetPrice(cp *CurrencyPair) (Price, error)
}

var ex = map[string]Exchange{}

func RegisterEx(name string, e Exchange) {
	if e != nil {
		ex[name] = e
	}
}

func GetExs() map[string]Exchange {
	return ex
}

func GetEx(name string) Exchange {
	if e, ok := ex[name]; ok {
		return e
	}
	return nil
}

func ListEx() (exchanges []string) {
	for key, _ := range ex {
		exchanges = append(exchanges, key)
	}
	return
}

var isAlive = func(js *Json) (interface{}, error) {
	return nil, nil
}
var notAlive = func(js *Json) (interface{}, error) {
	return nil, errors.New("Failed")
}

func recvResp(req *http.Request) (int, []byte) {
	client := &http.Client{}
	resp, _ := client.Do(req)
	body, _ := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	return resp.StatusCode, body
}

func ProcessResp(status int, js *Json, respOk, respErr RespHandle) (interface{}, error) {
	if respOk == nil || respErr == nil {
		return nil, errors.New("No proper handler")
	}
	if status == http.StatusOK {
		return respOk(js)
	} else {
		return respErr(js)
	}
}
