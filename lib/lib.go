package lib

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
)

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

func GetBytes(key interface{}) ([]byte, error) {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	err := enc.Encode(key)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

type Exchange interface {
	Alive() bool
	GetBalance() ([]Balance, error)
}

var ex = map[string]Exchange{}

func RegisterEx(name string, e Exchange) {
	ex[name] = e
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
