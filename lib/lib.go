package lib

import (
	"errors"
	"io/ioutil"
	"net/http"

	. "github.com/bitly/go-simplejson"
)

type RespHandle func(js *Json) (interface{}, error)

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
