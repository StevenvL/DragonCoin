package main

import (
	"crypto/rsa"
	"encoding/json"
	"fmt"
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
	nonce       int
	pubKey      rsa.PublicKey
	sig         []byte
	outputs     map[string]int
	fee         int
	data        string
	id          string
}

func (base Transaction) newTransaction(from string, nonce int, pubKey rsa.PublicKey, sig []byte, outputs map[string]int, fee int, data string) *Transaction {
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
	/*
	   return sha256hash(TX_CONST +
	       transaction.from +
	       strconv.Itoa(transaction.nonce) +
	       getStringPubKey(&transaction.pubKey) +
	       arrayToString(transaction.outputs, ",") +
	       strconv.Itoa(transaction.fee) +
	       transaction.data)
	*/
	b, _ := json.Marshal(transaction)
	return sha256hash(TX_CONST + string(b))
}

//Passes transaction by pointer so we can modify inside it.
/**
 * Determines whether the signature of the transaction is valid
 * and if the from address matches the public key.
 *
 * @returns {Boolean} - Validity of the signature and from address.
 */
func signTransaction(privKey *rsa.PrivateKey, transaction *Transaction) {
	id := getID(*transaction)
	res := sign(privKey, id)
	transaction.sig = res
}

/**
 * Determines whether the signature of the transaction is valid
 * and if the from address matches the public key.
 *
 * @returns {Boolean} - Validity of the signature and from address.
 */
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

/**
 * Verifies that there is currently sufficient gold for the transaction.
 *
 * @param {Block} block - Block used to check current balances
 *
 * @returns {boolean} - True if there are sufficient funds for the transaction,
 *    according to the balances from the specified block.
 */
func (base Transaction) sufficientFunds(block Block) bool {
	blockBalanceMap := block.Balances
	blockValue := blockBalanceMap[base.from]
	return base.totalOutputs() <= blockValue
}

/**
 * Calculates the total value of all outputs, including the transaction fee.
 *
 * @returns {Number} - Total amount of gold given out with this transaction.
 */
func (base Transaction) totalOutputs() int {
	var total = 0
	for _, value := range base.outputs {
		total += value
	}
	return total
}

//from https://stackoverflow.com/questions/37532255/one-liner-to-transform-int-into-string
func arrayToString(a []int, delim string) string {
	return strings.Trim(strings.Replace(fmt.Sprint(a), " ", delim, -1), "[]")
	//return strings.Trim(strings.Join(strings.Split(fmt.Sprint(a), " "), delim), "[]")
	//return strings.Trim(strings.Join(strings.Fields(fmt.Sprint(a)), delim), "[]")
}
