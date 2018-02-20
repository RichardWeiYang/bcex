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

const (
	Alive     = "Alive"
	Cancelled = "Cancelled"
	Unknown   = "Unknown"
)

type Order struct {
	Id                       string
	CP                       CurrencyPair
	Side                     string
	Price                    float64
	Amount, Remain, Executed float64
	State                    string
}

type Exchange interface {
	ToSymbol(cp *CurrencyPair) string
	NormSymbol(cp *string) string

	OrderState(interface{}) string
	OrderSide(string) string

	Alive() bool
	GetPrice(cp *CurrencyPair) (Price, error)
	GetSymbols() ([]string, error)
	GetDepth(cp *CurrencyPair) (Depth, error)

	SetKey(access, secret string)

	GetBalance() ([]Balance, error)
	NewOrder(o *Order) (string, error)
	CancelOrder(o *Order) error
	QueryOrder(o *Order) (Order, error)
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

func recvResp(req *http.Request) (int, *Json, error) {
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return 0, nil, err
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return 0, nil, err
	}
	resp.Body.Close()
	js, err := NewJson(body)
	if err != nil {
		return 0, nil, err
	}

	return resp.StatusCode, js, nil
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
