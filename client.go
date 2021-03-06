package main

import (
	"encoding/json"
	"errors"
	"fmt"

	emission "github.com/chuckpreslar/emission"
)

type Client struct {
	name                           string
	keypairClient                  keypair
	nonce                          int
	pendingOutGoingTransactionsMap map[string]Transaction
	pendingReceivedTransactionsMap map[string]Transaction
	address                        string
	blocks                         map[string]Block
	pendingBlocks                  map[string]Block
	lastBlock                      Block
	lastConfirmedBlock             Block
	emitter                        *emission.Emitter
	fakeNet                        *FakeNet
}

/*
toString() method if needed.
func (base Client) String() string {
	return fmt.Sprintf("Name: %s, Address: %s\n", base.name, base.address)
}*/

func newClient(name string, keypairClient keypair, startingBlock Block, fakeNet *FakeNet) *Client {
	client := new(Client)
	client.name = name

	if keypairClient.pubKey.N == nil {
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

	if startingBlock.NotEmpty {
		client.setGenesisBlock(startingBlock)
	}

	client.emitter = emission.NewEmitter()
	client.fakeNet = fakeNet

	client.emitter.On(PROOF_FOUND, client.receiveBlock)
	client.emitter.On(MISSING_BLOCK, client.provideMissingBlock)

	return client
}

/**
 * The genesis block can only be set if the client does not already
 * have the genesis block.
 *
 * @param {Block} startingBlock - The genesis block of the blockchain.
 */
func (base *Client) setGenesisBlock(startingBlock Block) {
	if base.lastBlock.NotEmpty {
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

	return base.confirmedBalance() - pendingSpent
}

/**
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
func (base *Client) postTransaction(outputs map[string]int, fee int) Transaction {
	var totalPayments = 0
	for _, element := range outputs {
		totalPayments += element
	}
	if totalPayments > base.availableGold() {
		fmt.Printf("ERROR!!!, Request %d, but account only has %d\n", totalPayments, base.availableGold())
	}

	var tx *Transaction
	var sig = []byte{0}
	tx = newTransaction(base.address, base.nonce, base.keypairClient.pubKey, sig, outputs, fee, "")
	signTransaction(base.keypairClient.privKey, tx)

	base.pendingOutGoingTransactionsMap[tx.Id] = *tx

	base.nonce++

	txJSON, _ := json.Marshal(tx)
	base.fakeNet.broadcast(POST_TRANSACTION, txJSON)

	return *tx
}

/*
* @param {Block | Object} block - The block to add to the clients list of available blocks.
   *
   * @returns {Block | null} The block with rerun transactions, or null for an invalid block.
   if NIL, that means no erros and valid block
   otherwise we have invalid block.

*/
func (base *Client) receiveBlock(block Block) (Block, error) {
	// If the block is a string, then deserialize it.
	// It literally can't be, so this is pointless.
	// If we want to handle strings, we'll need a new function for that.
	// Ignore the block if it has been received previously.

	if _, ok := base.blocks[block.getID()]; ok {
		return block, errors.New("Block Recieved Previously")
	}

	// First, make sure that the block has a valid proof.
	if !block.hasValidProof() && !block.isGenesisBlock() {
		return block, errors.New("Block does not have valid proof")
	}

	// Make sure that we have the previous blocks, unless it is the genesis block.
	// If we don't have the previous blocks, request the missing blocks and exit.
	prevBlock := base.blocks[block.PrevBlockHash]
	if !prevBlock.NotEmpty && !block.isGenesisBlock() {
		stuckBlocks := base.pendingBlocks[block.PrevBlockHash]

		// If this is the first time that we have identified this block as missing,
		// send out a request for the block.
		var stuckBlocksMap map[string]Block
		if !stuckBlocks.NotEmpty {
			base.requestMissingBlock(block)
			stuckBlocksMap = make(map[string]Block)
		}
		stuckBlocksMap[block.getID()] = block
	}

	if block.isGenesisBlock() {
		success := block.rerun(prevBlock)
		if success == false {
			return block, errors.New("Success is false")
		}
	}

	base.blocks[block.getID()] = block
	if base.lastBlock.ChainLength < block.ChainLength {
		base.lastBlock = block
		base.setLastConfirmed()
	}

	var unstuckBlocks []Block
	if _, ok := base.pendingBlocks[block.getID()]; ok {
		unstuckBlocks = append(unstuckBlocks, base.pendingBlocks[block.getID()])
	}

	delete(base.pendingBlocks, block.getID())

	for _, block := range unstuckBlocks {
		fmt.Printf("Processing unstuck block %s\n", block.getID())
		base.receiveBlock(block)
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
	fmt.Printf("%s asking for missing block %s \n", base.name, block.PrevBlockHash)
	m := Message{base.address, block.PrevBlockHash}
	b, err := json.Marshal(m)
	if err == nil {
		base.fakeNet.broadcast(MISSING_BLOCK, b)
	} else {
		fmt.Print("Error in JSON encoding in requestMissingBlock()")
	}
}

/**
 * Resend any transactions in the pending list.
 */
func (base Client) resendPendingTransactions() {
	for _, value := range base.pendingOutGoingTransactionsMap {
		valueJSON, err := json.Marshal(value)
		if err == nil {
			base.fakeNet.broadcast(POST_TRANSACTION, valueJSON)
		} else {
			fmt.Print("Error in JSON encoding in resendPendingTransactions()")
		}

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
	if _, ok := base.blocks[message.missing]; ok {
		fmt.Print("Providing missing block %s", message.missing)
		newBlock := base.blocks[message.missing]

		base.emitter.Emit(PROOF_FOUND, message.from, newBlock)
	}
}

/**
 * Sets the last confirmed block according to the most recently accepted block,
 * also updating pending transactions according to this block.
 * Note that the genesis block is always considered to be confirmed.
 */
func (base Client) setLastConfirmed() {
	block := base.lastBlock
	confirmedBlockHeight := block.ChainLength - CONFIRMED_DEPTH

	if confirmedBlockHeight < 0 {
		confirmedBlockHeight = 0
	}

	//While loop in go
	for block.ChainLength > confirmedBlockHeight {
		if _, ok := base.blocks[block.PrevBlockHash]; ok {
			block = base.blocks[block.PrevBlockHash]
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

	for id, balance := range base.lastConfirmedBlock.Balances {
		fmt.Printf("%s: %d\n", id, balance)
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
	for !block.NotEmpty {
		fmt.Printf("%s\n", block.getID())
		if _, ok := base.blocks[block.PrevBlockHash]; ok {
			block = base.blocks[block.PrevBlockHash]
		}
	}
}
