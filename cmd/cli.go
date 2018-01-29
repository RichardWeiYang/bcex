package cmd

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	. "github.com/RichardWeiYang/bcex/lib"
	"github.com/jawher/mow.cli"
)

var bcexKey *string

type ExchangeKey struct {
	AccessKeyId string `json:"AccessKeyId"`
	SecretKeyId string `json:"SecretKeyId"`
}

var keys map[string]ExchangeKey

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

	ReadConf()

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
	raw := Encrypt(string(plain), *bcexKey)
	ioutil.WriteFile("config.json", []byte(raw), 0644)
}

func ReadConf() {
	raw, err := ioutil.ReadFile("config.json")
	if err != nil {
		return
	}

	plain := Decrypt(string(raw), *bcexKey)

	json.Unmarshal([]byte(plain), &keys)
}

// CLI struct for main
type CLI struct {
	*cli.Cli
}

// NewCLI initializes new command line interface
func NewCLI() *CLI {
	c := &CLI{cli.App("bcex", "A BlockChain Exchange CLI")}

	bcexKey = c.String(cli.StringOpt{
		Name:      "k bcex-key",
		Desc:      "BCEX Key",
		EnvVar:    "BCEX_KEY",
		HideValue: true,
	})

	return c
}
