package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/url"
	"sort"
	"strings"
)

func validateInitData(initData, token string) bool {
	params, _ := url.ParseQuery(initData)
	hash := params.Get("hash")
	if hash == "" {
		return false
	}
	params.Del("hash")
	
	var keys []string
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	
	var parts []string
	for _, k := range keys {
		parts = append(parts, k+"="+params.Get(k))
	}
	dataCheckString := strings.Join(parts, "\n")
	
	h := hmac.New(sha256.New, []byte("WebAppData"))
	h.Write([]byte(token))
	secretKey := h.Sum(nil)
	
	h2 := hmac.New(sha256.New, secretKey)
	h2.Write([]byte(dataCheckString))
	expectedHash := hex.EncodeToString(h2.Sum(nil))
	
	return expectedHash == hash
}

func main() {
    // just a syntax check
	fmt.Println("OK")
}
