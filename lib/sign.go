package lib

import (
	"crypto/hmac"
	"crypto/md5"
	"crypto/rand"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/base64"
	"encoding/hex"
)

func ComputeHmac256(message string, secret string) string {
	key := []byte(secret)
	h := hmac.New(sha256.New, key)
	h.Write([]byte(message))
	return hex.EncodeToString(h.Sum(nil))
}

func ComputeHmac256Base64(message string, secret string) string {
	key := []byte(secret)
	h := hmac.New(sha256.New, key)
	h.Write([]byte(message))
	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}

func GetMD5Hash(text string) string {
	hasher := md5.New()
	hasher.Write([]byte(text))
	return hex.EncodeToString(hasher.Sum(nil))
}

func GetParamHmacSha384Sign(secret, params string) string {
	mac := hmac.New(sha512.New384, []byte(secret))
	_, err := mac.Write([]byte(params))
	if err != nil {
		return ""
	}
	return hex.EncodeToString(mac.Sum(nil))
}

type UUID [16]byte

var byteGroups = []int{8, 4, 4, 4, 12}

func (u *UUID) Bytes() []byte {
	return u[:]
}

func (u *UUID) String() string {
	bytes := u.Bytes()
	result := hex.EncodeToString(bytes[0 : byteGroups[0]/2])
	start := byteGroups[0] / 2
	for i := 1; i < len(byteGroups); i++ {
		nBytes := byteGroups[i] / 2
		result += "-"
		result += hex.EncodeToString(bytes[start : start+nBytes])
		start += nBytes
	}
	return result
}

func GetUUID() string {
	uuid := new(UUID)
	rand.Read(uuid.Bytes())
	return uuid.String()
}
