package main

import (
	"encoding/json"
	"errors"
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

func newClient(name string, keypairClient keypair, startingBlock Block) *Client {
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
	return client
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
func (base Client) postTransaction(outputs map[string]int, fee int) Transaction {
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
   if NIL, that means no erros and valid block
   otherwise we have invalid block.

*/

///TODO
func (base Client) receiveBlock(block Block) (Block, error) {
	// If the block is a string, then deserialize it.
	block = deserializeBlock(block, blockchain)

	// Ignore the block if it has been received previously.
	if val, ok := base.blocks[block.getID()]; ok {
		return block, errors.New("Invalid block")
	}

	// First, make sure that the block has a valid proof.
	if !block.hasValidProof() && !block.isGenesisBlock() {
		fmt.Printf("Block %s does not have a valid proof", block.getID())
		return block, errors.New("Block does not have valid proof")
	}

	// Make sure that we have the previous blocks, unless it is the genesis block.
	// If we don't have the previous blocks, request the missing blocks and exit.
	prevBlock := base.blocks[block.prevBlockHash]
	if prevBlock.getID() != "" && !prevBlock.isGenesisBlock() {
		stuckBlocks := base.pendingBlocks[block.prevBlockHash]

		// If this is the first time that we have identified this block as missing,
		// send out a request for the block.
		if stuckBlocks.getID() != "" {
			base.requestMissingBlock(block)
			//stuckBlocks = set()
		}
	}

	return block, nil
}

type Message struct {
	from    string
	missing string
}

/**
 * Request the previous block from the network.
 *
 * @param {Block} block - The block that is connected to a missing block.
 */
func (base Client) requestMissingBlock(block Block) {
	fmt.Print("Asking for missing block %s", block.prevBlockHash)
	m := Message{base.address, block.prevBlockHash}
	b, err := json.Marshal(m)
	if err == nil {
		emitter.Emit(blockchain.MISSING_BLOCK(), b)
	} else {
		fmt.Print("Error in JSON encoding in requestMissingBlock()")
	}
}

/**
 * Resend any transactions in the pending list.
 */
func (base Client) resendPendingTransactions() {
	for key, value := range base.pendingOutGoingTransactionsMap {
		emitter.Emit(blockchain.POST_TRANSACTION(), value)
	}
}

/**
 * Takes an object representing a request for a misssing block.
 * If the client has the block, it will send the block to the
 * client that requested it.
 *
 * @param {Object} msg - Request for a missing block.
 * @param {String} msg.missing - ID of the missing block.
 */
func (base Client) provideMissingBlock(msg []byte) {
	var message Message
	json.Unmarshal([]byte(msg), &message)
	if val, ok := base.blocks[message.missing]; ok {
		fmt.Print("Providing missing block %s", message.missing)
		newBlock := base.blocks[message.missing]

		//this.net.sendMessage(msg.from, Blockchain.PROOF_FOUND, block);
		emitter.Emit(blockchain.PROOF_FOUND(), message.from, newBlock)
	}
}

/**
 * Sets the last confirmed block according to the most recently accepted block,
 * also updating pending transactions according to this block.
 * Note that the genesis block is always considered to be confirmed.
 */
func (base Client) setLastConfirmed() {
	block := base.lastBlock
	confirmedBlockHeight := block.chainLength - blockchain.getCONFIRMED_DEPTH()

	if confirmedBlockHeight < 0 {
		confirmedBlockHeight = 0
	}

	//no such thing as while loop in GO
	for block.chainLength > confirmedBlockHeight {
		if _, ok := base.blocks[block.prevBlockHash]; ok {
			block = base.blocks[block.prevBlockHash]
		}
	}
	base.lastConfirmedBlock = block

	// Update pending transactions according to the new last confirmed block.
	for txID, tx := range base.pendingOutGoingTransactionsMap {
		if base.lastConfirmedBlock.contains(tx) {
			delete(base.pendingOutGoingTransactionsMap, txID)
		}
	}
}

func (base Client) showAllBalances() {
	fmt.Print("Show all balances:")

	for id, balance := range base.lastConfirmedBlock.balances {
		fmt.Printf("%s: %d", id, balance)
	}
}

func (base Client) log(msg string) {
	nameToDisplay := base.name
	if base.name == "" {
		nameToDisplay = base.address[0:10]
	}

	fmt.Printf("%s: %s", nameToDisplay, msg)
}

func (base Client) showBlockChain() {
	block := base.lastBlock
	fmt.Print("BLOCKCHAIN:")
	for block.getID() != "" {
		fmt.Printf("%s", block.getID())
		if _, ok := base.blocks[block.prevBlockHash]; ok {
			block = base.blocks[block.prevBlockHash]
		}
	}
}
