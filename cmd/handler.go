package cmd

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	_ "encoding/base64"
	"encoding/pem"
	"errors"
	"io/ioutil"
	"os"
)

func GenRsaKey() error {
	bits := 2048
	privateKey, err := rsa.GenerateKey(rand.Reader, bits)
	if err != nil {
		return err
	}
	derStream := x509.MarshalPKCS1PrivateKey(privateKey)
	block := &pem.Block{
		Type:  "private",
		Bytes: derStream,
	}
	file, err := os.Create(".private.pem")
	if err != nil {
		return err
	}
	err = pem.Encode(file, block)
	if err != nil {
		return err
	}

	publicKey := &privateKey.PublicKey
	derPkix, err := x509.MarshalPKIXPublicKey(publicKey)
	if err != nil {
		return err
	}
	block = &pem.Block{
		Type:  "public",
		Bytes: derPkix,
	}
	file, err = os.Create(".public.pem")
	if err != nil {
		return err
	}
	err = pem.Encode(file, block)
	if err != nil {
		return err
	}
	return nil
}

func RsaEncrypt(origData, publicKey []byte) ([]byte, error) {
	block, _ := pem.Decode(publicKey)
	if block == nil {
		return nil, errors.New("public key error")
	}
	pubInterface, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	pub := pubInterface.(*rsa.PublicKey)
	return rsa.EncryptPKCS1v15(rand.Reader, pub, origData)
}

func RsaDecrypt(ciphertext, privateKey []byte) ([]byte, error) {
	block, _ := pem.Decode(privateKey)
	if block == nil {
		return nil, errors.New("private key error!")
	}
	priv, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	return rsa.DecryptPKCS1v15(rand.Reader, priv, ciphertext)
}

func getPrivateKey() []byte {
	privateKey, err := ioutil.ReadFile(".private.pem")
	if err != nil {
		panic(err)
	}

	return privateKey
}

func getPublicKey() []byte {
	publicKey, err := ioutil.ReadFile(".public.pem")
	if err != nil {
		panic(err)
	}

	return publicKey
}

func init() {
	if _, err := os.Stat(".private.pem"); os.IsNotExist(err) {
		GenRsaKey()
	}
}
