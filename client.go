package main

import (
	"fmt"

	"github.com/chuckpreslar/emission"
)

var blockchain = newBlockchain()
var emitter = emission.NewEmitter()

//Client NEED TO ADD STUFF DEALING WITH BLOCK
type Client struct {
	name                           string
	keypairClient                  keypair
	nonce                          int
	pendingOutGoingTransactionsMap map[string]Transaction
	pendingReceivedTransactionsMap map[string]Transaction
	startingBlock                  Block
	address                        string
	blocks                         map[string]Block
	pendingBlocks                  map[string]Block
	lastBlock                      Block
	lastConfirmedBlock             Block
}

func newClient(name string, keypairClient keypair, startingBlock Block) {
	client := new(Client)
	client.name = name

	if keypairClient.pubKey.N.String() == "" {
		client.keypairClient = generateKeypair()
	} else {
		client.keypairClient = keypairClient
	}

	client.address = calcAddress(&client.keypairClient.pubKey)
	client.nonce = 0

	client.pendingOutGoingTransactionsMap = make(map[string]Transaction)
	client.pendingReceivedTransactionsMap = make(map[string]Transaction)

	// A map of all block hashes to the accepted blocks.
	client.blocks = make(map[string]Block)
	client.pendingBlocks = make(map[string]Block)

	if startingBlock.getID() != "" {
		client.setGenesisBlock(startingBlock)
	}

}

/**
 * The genesis block can only be set if the client does not already
 * have the genesis block.
 *
 * @param {Block} startingBlock - The genesis block of the blockchain.
 */
func (base Client) setGenesisBlock(startingBlock Block) {
	if base.lastBlock.getID() != "" {
		fmt.Print("ERROR!, Cannot set genesis block for existing blockchain")
	} else {

		// Transactions from this block or older are assumed to be confirmed,
		// and therefore are spendable by the client. The transactions could
		// roll back, but it is unlikely.
		base.lastConfirmedBlock = startingBlock

		// The last block seen.  Any transactions after lastConfirmedBlock
		// up to lastBlock are considered pending.
		base.lastBlock = startingBlock

		base.blocks[startingBlock.getID()] = startingBlock
	}
}

/**
 * The amount of gold available to the client, not counting any pending
 * transactions.  This getter looks at the last confirmed block, since
 * transactions in newer blocks may roll back.
 */
func (base Client) confirmedBalance() int {
	return base.lastConfirmedBlock.balanceOf(base.address)
}

/**
 * Any gold received in the last confirmed block or before is considered
 * spendable, but any gold received more recently is not yet available.
 * However, any gold given by the client to other clients in unconfirmed
 * transactions is treated as unavailable.
 */
func (base Client) availableGold() int {
	var pendingSpent = 0
	for _, element := range base.pendingOutGoingTransactionsMap {
		pendingSpent += element.totalOutputs()
	}

	return pendingSpent
}

/**

   TODO!!!. HOW DO WE BROADCAST?
   We could send it using function somehow....
   Might have to send send array of pointer of client structs

  * Broadcasts a transaction from the client giving gold to the clients
  * specified in 'outputs'. A transaction fee may be specified, which can
  * be more or less than the default value.
  *
  * @param {Array} outputs - The list of outputs of other addresses and
  *    amounts to pay.
  * @param {number} [fee] - The transaction fee reward to pay the miner.
  *
  * @returns {Transaction} - The posted transaction.
*/
func (base Client) postTransaction(outputs map[string]int, fee int, payPerAddress []int, clientList *[]Client) Transaction {
	var totalPayments = 0
	for _, element := range outputs {
		totalPayments += element
	}
	if totalPayments > base.availableGold() {
		fmt.Printf("ERROR!!!, Request %d, but account only has %d", totalPayments, base.availableGold())
	}

	var tx Transaction
	var sig = []byte{0}
	tx.newTransaction(base.address, base.nonce, base.keypairClient.pubKey, sig, outputs, fee, "")
	resTx := blockchain.makeTransaction(tx)
	signTransaction(base.keypairClient.privKey, &resTx)

	base.pendingOutGoingTransactionsMap[resTx.id] = resTx

	base.nonce++

	//this.net.broadcast(Blockchain.POST_TRANSACTION, tx); HOW TO DO THIS???
	emitter.Emit(blockchain.POST_TRANSACTION, resTx)

	return resTx
}

/*
* @param {Block | Object} block - The block to add to the clients list of available blocks.
   *
   * @returns {Block | null} The block with rerun transactions, or null for an invalid block.
*/

func (base Client) receiveBlock(block Block) *Block {
	block
	block = deserializeBlock(block, blockchain)

	if val, ok := base.blocks[block.getID()]; ok {
		return
	}

	return block
}
