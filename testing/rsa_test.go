package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/hex"
	"encoding/pem"
	"errors"
	"testing"

	"goapptol/utils"
)

func TestRsa(t *testing.T) {
	x, err := RsaEncrypt(utils.PubPem, []byte("idss"))
	if err != nil {
		t.Errorf("RsaEncrypt error:%v\n", err)
	} else {
		t.Logf("RsaEncrypt: %d\n%s\n", len(x), hex.Dump(x))
		t.Logf("%s\n", hex.EncodeToString(x))
	}

	y, err := RsaDecrypt(utils.KeyPem, x)
	if err != nil {
		t.Errorf("RsaDecrypt error:%v\n", err)
	} else {
		t.Logf("RsaDecrypt: %d\n%s\n", len(y), hex.Dump(y))
		t.Logf("%s\n", y)
	}
}

func RsaEncrypt(pubpem, data []byte) ([]byte, error) {
	pemblock, _ := pem.Decode(pubpem)
	if pemblock == nil {
		return nil, errors.New("decode public key error")
	}

	pub, err := x509.ParsePKIXPublicKey(pemblock.Bytes)
	if err != nil {
		return nil, err
	}
	return rsa.EncryptPKCS1v15(rand.Reader, pub.(*rsa.PublicKey), data)
}

func RsaDecrypt(prvkey, cipher []byte) ([]byte, error) {
	pemblock, _ := pem.Decode(prvkey)
	if pemblock == nil {
		return nil, errors.New("decode private key error")
	}
	prv, err := x509.ParsePKCS1PrivateKey(pemblock.Bytes)
	if err != nil {
		return nil, err
	}
	return rsa.DecryptPKCS1v15(rand.Reader, prv, cipher)
}
