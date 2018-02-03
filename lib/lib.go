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

type Unit struct {
	Price  float64
	Amount float64
}

type Depth struct {
	Bids []Unit
	Asks []Unit
}

type Exchange interface {
	ToSymbol(cp *CurrencyPair) string
	SetKey(access, secret string)
	Alive() bool
	GetBalance() ([]Balance, error)
	GetPrice(cp *CurrencyPair) (Price, error)
	GetSymbols() ([]string, error)
	GetDepth(cp *CurrencyPair) (Depth, error)
}

type NewExchange func() Exchange

var exs = map[string]NewExchange{}

func RegisterEx(name string, ne NewExchange) {
	if ne != nil {
		exs[name] = ne
	}
}

func GetEx(name string) Exchange {
	if ne, ok := exs[name]; ok {
		return ne()
	}
	return nil
}

func ListEx() (exchanges []string) {
	for key, _ := range exs {
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
