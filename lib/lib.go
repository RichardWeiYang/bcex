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

type Balance struct {
	Currency string
	Balance  string
}

var keys map[string]ExchangeKey
var initialized = false

func readConf() {
	if initialized {
		return
	}
	initialized = true
	raw, err := ioutil.ReadFile("config.json")
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	json.Unmarshal(raw, &keys)
}

type Exchange interface {
	Alive() bool
	GetBalance() ([]Balance, error)
}

var ex = map[string]Exchange{}

func RegisterEx(name, ak, sk string, e Exchange) {
	if (ak != "replaceme" || ak != "") && (sk != "replaceme" || sk != "") {
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
