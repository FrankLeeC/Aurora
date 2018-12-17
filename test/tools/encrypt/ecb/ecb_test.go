package test

import (
	"encoding/base64"
	"testing"

	"github.com/FrankLeeC/Aurora/tools/encrypt/ecb"
)

func TestECB(t *testing.T) {
	src := "12345"
	key := "0123456789abcdef" // 16位对应128bit 24对应192bit 32对应256bit

	crypted, err := ecb.Encrypt(src, key)
	if err != nil {
		t.Error(err.Error())
		return
	}
	text := base64.StdEncoding.EncodeToString(crypted)
	t.Log(text)
	plain, err := ecb.Decrypt(crypted, key)
	if err != nil {
		t.Error(err.Error())
		return
	}
	t.Log(string(plain))
}
