package main

import (
	"errors"
	"fmt"
	"reflect"
	"time"
)

type Miner struct {
	Client
	name          string
	startingBlock Block
	keypairMiner  keypair
	miningRounds  int
	currentBlock  Block
}

func newMiner(name string, keypairMiner keypair, startingBlock Block) *Miner {
	miner := new(Miner)
	miner.Client = *newClient(name, keypairMiner, startingBlock)
	miner.miningRounds = blockchain.NUM_ROUNDS_MINING()

	return miner
}

func (base Miner) initialize() {
	var set []Transaction
	base.startNewSearch(set)

	//Not sure if these lines even work
	emitter.On(blockchain.START_MINING, base.findProof)
	emitter.On(blockchain.POST_TRANSACTION, base.addTransaction)

	time.AfterFunc(0*time.Second, emitStartMining)
}

func emitStartMining() {
	emitter.Emit(blockchain.START_MINING)
}

//This method creates a new array if empty.
//Otherwise use specified array
func (base Miner) startNewSearch(set []Transaction) {
	//suppoed to pass this.address and this.miningrounds to it...
	base.currentBlock = *blockchain.makeEmptyBlock()

	for _, tx := range set {
		base.addTransaction(tx)
	}

	base.currentBlock.proof = 0
}

func (base Miner) findProof() {
	pausePoint := base.currentBlock.proof + base.miningRounds

	for base.currentBlock.proof < pausePoint {
		if base.currentBlock.hasValidProof() == true {
			fmt.Printf("Found proof for block %d: %s", base.currentBlock.chainLength, base.currentBlock.proof)
			base.announceProof()
			base.receiveBlock(base.currentBlock)
			var set []Transaction
			base.startNewSearch(set)
			break
		}
		base.currentBlock.proof++
	}
	/*
			USED FOR TESTING PURPOSES, not sure if we have to port.
			// If we are testing, don't continue the search.
		    if (!oneAndDone) {
				// Check if anyone has found a block, and then return to mining.
				setTimeout(() => this.emit(Blockchain.START_MINING), 0);
			  }
	*/
}
func (base Miner) addTransaction(tx Transaction) bool {
	tx = blockchain.makeTransaction(tx)
	//supposed to do client.print but its whatever.
	return base.currentBlock.addTransaction(tx)
}

func (base Miner) announceProof() {
	//this.net.broadcast(Blockchain.PROOF_FOUND, this.currentBlock);
}

func (base Miner) receiveBlock(block Block) error {
	b, err := base.Client.receiveBlock(block)

	if err != nil {
		return errors.New("Invalid block")
	}

	if base.currentBlock.getID() != "" && b.chainLength >= base.currentBlock.chainLength {
		fmt.Print("Cutting over to new chain")
		txSet := base.syncTransactions(b)
		base.startNewSearch(txSet)
	}

	return nil
}

func (base Miner) syncTransactions(nb Block) []Transaction {
	cb := base.currentBlock
	var cbTxs []Transaction
	var nbTxs []Transaction

	for nb.chainLength > cb.chainLength {
		for _, element := range nb.transactions {
			nbTxs = append(nbTxs, element)
			nb = base.blocks[nb.prevBlockHash]
			if nb.getID() == "" {
				fmt.Print("no result found in map")
			}
		}
	}

	for cb.getID() != "" && cb.getID() != nb.getID() {
		for _, element := range cb.transactions {
			cbTxs = append(cbTxs, element)
		}
		for _, element := range nb.transactions {
			nbTxs = append(nbTxs, element)
		}

		cb = base.blocks[cb.prevBlockHash]
		nb = base.blocks[nb.prevBlockHash]
	}

	for _, element := range nbTxs {
		indexInCbTxs := indexOf(element, cbTxs)
		if indexInCbTxs != 1 {
			cbTxs = remove(cbTxs, indexInCbTxs)
		}
	}
	return cbTxs
}

//https://stackoverflow.com/a/37335777
//Removes element from slice
func remove(s []Transaction, i int) []Transaction {
	s[len(s)-1], s[i] = s[i], s[len(s)-1]
	return s[:len(s)-1]
}

func indexOf(transaction Transaction, list []Transaction) int {
	for index, element := range list {
		if reflect.DeepEqual(transaction, element) {
			return index
		}
	}
	return -1
}

//somehow take variable args
//TODO
//usualyl outputs map
//map[string]int
//Usually amounts and address
func (base Miner) postTransaction(outputs map[string]int) bool {
	tx := base.Client.postTransaction(outputs, blockchain.getDEFAULT_TX_FEE())
	return base.addTransaction(tx)
}