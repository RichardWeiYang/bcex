# bcex

A Blockchain Exchage Interface

# How to use

## Download and install

go get github.com/RichardWeiYang/bcex

## Configure

### BCEX KEY

BCEX KEY is a private key to encrypt your Exchange API-KEY configuration, so please remember to keep it secret.

You could set BCEX KEY in two ways:

    * ENV variable
    * -k option

ENV variable

```
export BCEX_KEY=key
```

-k option

```
./bcex -k key COMMAND
```

Note: key must be 16, 24 or 32 bytes.

### Exchange API-KEY

The second configuration you need to setup is the API-KEY of the exchange which you want to connect. 

First get the API-KEY from your exchange page, then run following command to setup.

For example, to setup the API-KEY for bitfinex

```
./bcex setkey bitfinex access_key secret_key
```

## Commands Example

### list exchanges supported

```
./bcex list
```


### show your account balance

```
./bcex balance xxx
```


# Welcome contribution

It is far from ready, welcome any contribution.
