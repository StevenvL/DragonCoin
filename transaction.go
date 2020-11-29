package main

import (
	"crypto/rsa"
	"strings"
)

const TX_CONST = "TX"

// TransactionIO is apart of Transaction
type TransactionIO struct {
	inputs  []string
	outputs []string
}

// Transaction object
type Transaction struct {
	tranctionIO TransactionIO
	from        string
	nonce       string
	pubKey      rsa.PublicKey
	sig         []byte
	outputs     []int
	fee         string
	data        string
	id          string
}

func newTransaction(from string, nonce string, pubKey rsa.PublicKey, sig []byte, outputs []int, fee string, data string) *Transaction {
	transaction := new(Transaction)
	transaction.from = from
	transaction.nonce = nonce
	transaction.pubKey = pubKey
	transaction.sig = sig
	transaction.outputs = outputs
	transaction.fee = fee
	transaction.data = data
	transaction.id = getID(*transaction)

	if len(outputs) > 0 {

	}
	/* UNSURE IF WE NEED TO CONVER TO GOLANG
		   Looks like it just parses the int to decimal.
		 if (outputs) outputs.forEach(({amount, address}) => {
	      if (typeof amount !== 'number') {
	        amount = parseInt(amount, 10);
	      }
	      this.outputs.push({amount, address});
		});
	*/

	return transaction
}

func getID(transaction Transaction) string {
	return sha256hash(TX_CONST +
		transaction.from +
		transaction.nonce +
		getStringPubKey(&transaction.pubKey) +
		strings.Join(transaction.outputs, "") +
		transaction.fee +
		transaction.data)
}

//Passes transaction by pointer so we can modify inside it.
func signTransaction(privKey *rsa.PrivateKey, transaction *Transaction) {
	id := getID(*transaction)
	res := sign(privKey, id)
	transaction.sig = res
}

func validSignatureTransaction(transaction Transaction) bool {
	bool1 := len(transaction.sig) != 0
	bool2 := addressMatchesKey(transaction.from, &transaction.pubKey)
	response := verifySignature(&transaction.pubKey, transaction.id, transaction.sig)
	bool3 := false
	if response == nil {
		bool3 = true
	}

	return bool1 && bool2 && bool3
}

//TODO!!!!!!! needs BLOCK.GO
func sufficientFunds(block Block) bool {

}

func totalOutputs(transaction Transaction) int {
	var total = 0
	for index, value := range transaction.outputs {
		total += value
	}
	return total
}
