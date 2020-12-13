package main

import (
	"bytes"
	"crypto/rsa"
	"fmt"
	"strconv"
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
	//tranctionIO TransactionIO
	From    string
	Nonce   int
	PubKey  rsa.PublicKey
	Sig     []byte
	Outputs map[string]int
	Fee     int
	Data    string
	Id      string
}

func newTransaction(from string, nonce int, pubKey rsa.PublicKey, sig []byte, outputs map[string]int, fee int, data string) *Transaction {
	transaction := new(Transaction)
	transaction.From = from
	transaction.Nonce = nonce
	transaction.PubKey = pubKey
	transaction.Sig = sig
	transaction.Outputs = outputs
	transaction.Fee = fee
	transaction.Data = data
	transaction.Id = getID(*transaction)

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
		transaction.From +
		strconv.Itoa(transaction.Nonce) +
		getStringPubKey(&transaction.PubKey) +
		createKeyValuePairs(transaction.Outputs) +
		strconv.Itoa(transaction.Fee) +
		transaction.Data)

	//b, _ := json.Marshal(transaction)
	//return sha256hash(TX_CONST + string(b))
}

//From https://stackoverflow.com/a/48150584
func createKeyValuePairs(m map[string]int) string {
	b := new(bytes.Buffer)
	for key, value := range m {
		fmt.Fprintf(b, "%s=\"%s\"\n", key, value)
	}
	return b.String()
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
	transaction.Sig = res
}

/**
 * Determines whether the signature of the transaction is valid
 * and if the from address matches the public key.
 *
 * @returns {Boolean} - Validity of the signature and from address.
 */
func validSignatureTransaction(transaction Transaction) bool {
	bool1 := len(transaction.Sig) != 0
	bool2 := addressMatchesKey(transaction.From, &transaction.PubKey)
	response := verifySignature(&transaction.PubKey, getID(transaction), transaction.Sig)
	//fmt.Print("Transaction.go line 105, Response:")
	//fmt.Println(response)
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
	blockValue := blockBalanceMap[base.From]
	return base.totalOutputs() <= blockValue
}

/**
 * Calculates the total value of all outputs, including the transaction fee.
 *
 * @returns {Number} - Total amount of gold given out with this transaction.
 */
func (base Transaction) totalOutputs() int {
	var total = 0
	for _, value := range base.Outputs {
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
