/*
MIT License

Copyright (c) 2018 Frank Lee

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
*/

// Package encrypt rsa encryption and decryption
// Author: Frank Lee
// Date: 2017-11-17 11:51:41
// Last Modified by:   Frank Lee
// Last Modified time: 2018-03-22 13:35:41
package encrypt

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"io/ioutil"
	"path/filepath"
)

// KeyType key type
type KeyType int64

const (

	// PKCS1 pkcs1
	PKCS1 KeyType = iota

	// PKCS8 pkcs8
	PKCS8
)

func decodeFromKey(key []byte) (*pem.Block, error) {
	b, _ := pem.Decode(key)
	if b == nil {
		return nil, errors.New("key error")
	}
	return b, nil
}

func decodeFromFile(keyFile string) (*pem.Block, error) {
	var b []byte
	var err error
	if b, err = ioutil.ReadFile(keyFile); err != nil {
		return nil, err
	}
	return decodeFromKey(b)

}

func genPubKey(publicKey []byte) (*rsa.PublicKey, error) {
	pub, err := x509.ParsePKIXPublicKey(publicKey)
	if err != nil {
		return nil, err
	}
	return pub.(*rsa.PublicKey), nil
}

func genPriKey(privateKey []byte, keyType KeyType) (*rsa.PrivateKey, error) {
	var priKey *rsa.PrivateKey
	var err error
	switch keyType {
	case PKCS1:
		{
			priKey, err = x509.ParsePKCS1PrivateKey(privateKey)
			if err != nil {
				return nil, err
			}
		}
	case PKCS8:
		{
			prk, err := x509.ParsePKCS8PrivateKey(privateKey)
			if err != nil {
				return nil, err
			}
			priKey = prk.(*rsa.PrivateKey)
		}
	default:
		{
			return nil, errors.New("unsupport private key type")
		}
	}
	return priKey, nil
}

// GeneratePrivateKey generate private key using key string
func GeneratePrivateKey(privateKey []byte, keyType KeyType) (*rsa.PrivateKey, error) {
	b, err := decodeFromKey(privateKey)
	if err != nil {
		return nil, err
	}
	return genPriKey(b.Bytes, keyType)
}

// GeneratePrivateKeyFromFile generate private key using pem file
func GeneratePrivateKeyFromFile(keyPath string, keyType KeyType) (*rsa.PrivateKey, error) {
	priPath, err := filepath.Abs(keyPath)
	if err != nil {
		return nil, err
	}
	b, err := decodeFromFile(priPath)
	if err != nil {
		return nil, err
	}
	return genPriKey(b.Bytes, keyType)
}

// GeneratePublicKey generate public key using key string
func GeneratePublicKey(publicKey []byte) (*rsa.PublicKey, error) {
	b, err := decodeFromKey(publicKey)
	if err != nil {
		return nil, err
	}
	return genPubKey(b.Bytes)
}

// GeneratePublicKeyFromFile generate public key using pem file
func GeneratePublicKeyFromFile(keyPath string) (*rsa.PublicKey, error) {
	pubPath, err := filepath.Abs(keyPath)
	if err != nil {
		return nil, err
	}
	b, err := decodeFromFile(pubPath)
	if err != nil {
		return nil, err
	}
	return genPubKey(b.Bytes)
}

// RSAEncrypt rsa encrypt
// param plainText: plain text
// param publicKey: public key string
// return []byte: cipher bytes
// return error: errors
func RSAEncrypt(plainText string, publicKey *rsa.PublicKey) ([]byte, error) {
	bytes, err := rsa.EncryptPKCS1v15(rand.Reader, publicKey, []byte(plainText))
	if err != nil {
		return nil, err
	}
	return bytes, nil
}

// RSADecrypt rsa decrypt
// param cipherText: cipher text
// param privateKey: private key
// return []byte: plain bytes
// return error: errors
func RSADecrypt(cipherText string, privateKey *rsa.PrivateKey) ([]byte, error) {
	bytes, err := rsa.DecryptPKCS1v15(rand.Reader, privateKey, []byte(cipherText))
	if err != nil {
		return nil, err
	}
	return bytes, nil
}
