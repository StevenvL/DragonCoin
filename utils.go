package main

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"strconv"
)

type keypair struct {
	pubKey  rsa.PublicKey
	privKey *rsa.PrivateKey
}

func main() {
	/*
		UNCOMMENT TO TEST METHODS
		sum := sha256.Sum256([]byte("hello world\n"))
		fmt.Printf("%x", sum)

		sha256hash("hello world")
		keypair := generateKeypair()
		toSign := "date: Thu, 05 Jan 2012 21:31:40 GMT"
		fmt.Println()
		var signedMsg = sign(keypair.privKey, toSign)
		fmt.Println(signedMsg)

		var res = verifySignature(&keypair.pubKey, toSign, signedMsg)
		if res != nil {
			fmt.Fprintf(os.Stderr, "Error from verification: %s\n", res)
			return
		} else {
			fmt.Println("same")
		}

		var addr = calcAddress(&keypair.pubKey)
		fmt.Println(addressMatchesKey(addr, &keypair.pubKey))
	*/
}

func sha256hash(s string) string {
	res := sha256.Sum256([]byte(s))
	return hex.EncodeToString(res[:])
}

func generateKeypair() keypair {

	bitSize := 512
	// Private Key generation
	key, err := rsa.GenerateKey(rand.Reader, bitSize)
	if err != nil {
		fmt.Println(err)
	}
	pubKey := key.PublicKey
	res := keypair{}
	res.privKey = key
	res.pubKey = pubKey
	return res
}

func sign(privKey *rsa.PrivateKey, message string) []byte {
	data := ([]byte(message))
	h := sha256.New()
	h.Write(data)
	d := h.Sum(nil)
	res, _ := rsa.SignPKCS1v15(rand.Reader, privKey, crypto.SHA256, d)
	return res
}

//if return nil
//they are the same.
func verifySignature(pubKey *rsa.PublicKey, message string, sig []byte) error {
	messageBytes := ([]byte(message))
	h := sha256.New()
	h.Write(messageBytes)
	d := h.Sum(nil)
	return rsa.VerifyPKCS1v15(pubKey, crypto.SHA256, d, sig)
}

func calcAddress(pubKey *rsa.PublicKey) string {
	var stringPubKey = pubKey.N.String() + "" + strconv.Itoa(pubKey.E)
	return base64.StdEncoding.EncodeToString([]byte(stringPubKey))
}

func addressMatchesKey(addr string, pubKey *rsa.PublicKey) bool {
	return addr == calcAddress(pubKey)
}

func getStringPubKey(pubKey *rsa.PublicKey) string {
	var stringPubKey = pubKey.N.String() + "" + strconv.Itoa(pubKey.E)
	return stringPubKey
}
