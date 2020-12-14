package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
)

type Miner struct {
	*Client
	startingBlock Block
	keypairMiner  keypair
	miningRounds  int
	currentBlock  *Block
}

func newMiner(name string, keypairMiner keypair, startingBlock Block, fakeNet *FakeNet) *Miner {
	miner := new(Miner)
	miner.Client = newClient(name, keypairMiner, startingBlock, fakeNet)
	miner.miningRounds = NUM_ROUNDS_MINING

	return miner
}

func (base Miner) initialize() {
	var set []Transaction
	base.startNewSearch(set)

	base.emitter.On(START_MINING, base.findProof)
	base.emitter.On(POST_TRANSACTION, base.addTransaction)
	base.Client.emitter.Off(PROOF_FOUND, base.Client.receiveBlock)
	base.emitter.On(PROOF_FOUND, base.receiveBlock)

	base.emitStartMining()
}

func (base Miner) emitStartMining() {
	base.emitter.Emit(START_MINING)
}

//This method creates a new array if empty.
//Otherwise use specified array
func (base *Miner) startNewSearch(set []Transaction) {
	base.currentBlock = base.Client.lastBlock.makeBlock(base.Client.address)
	for _, tx := range set {
		base.addTransaction(tx)
	}

	base.currentBlock.Proof = 0
}

func (base *Miner) findProof() {
	pausePoint := base.currentBlock.Proof + base.miningRounds

	for base.currentBlock.Proof < pausePoint {
		if base.currentBlock.hasValidProof() == true {
			fmt.Printf("%v Found proof for block %v: %v, Character: %s\n", base.Client.name, base.currentBlock.ChainLength, base.currentBlock.Proof, base.currentBlock.generateDnDCharacter())
			base.announceProof()
			var set []Transaction
			base.startNewSearch(set)
			break
		}
		base.currentBlock.Proof++
	}

	base.emitStartMining()

}
func (base Miner) addTransaction(tx Transaction) bool {
	return base.currentBlock.addTransaction(tx)
}

func (base Miner) announceProof() {
	blockJSON, _ := json.Marshal(*base.currentBlock)
	base.Client.fakeNet.broadcast(PROOF_FOUND, blockJSON)
}

func (base *Miner) receiveBlock(block Block) error {
	b, err := base.Client.receiveBlock(block)

	if err != nil {
		fmt.Printf("%v encountered error %v\n", base.Client.name, err)
		return errors.New("Invalid block")
	} else if base.currentBlock.NotEmpty && b.ChainLength >= base.currentBlock.ChainLength {
		fmt.Printf("%v: Cutting over to new chain length %v from current length %v\n", base.Client.name, b.ChainLength, base.currentBlock.ChainLength)
		txSet := base.syncTransactions(b)
		base.startNewSearch(txSet)
	} else {
		fmt.Printf("New Chain Rejected because current chain is empty: %v, or current chain %v > new chain %v, or we announced this block\n", !base.currentBlock.NotEmpty, base.currentBlock.ChainLength, b.ChainLength)
	}

	return nil
}

func (base Miner) syncTransactions(nb Block) []Transaction {
	cb := *base.currentBlock
	var cbTxs []Transaction
	var nbTxs []Transaction

	for nb.ChainLength > cb.ChainLength {
		for _, element := range nb.Transactions {
			nbTxs = append(nbTxs, element)
			if !nb.NotEmpty {
				fmt.Print("no result found in map")
			}
		}
		nb = base.blocks[nb.PrevBlockHash]
	}

	for cb.NotEmpty && cb.getID() != nb.getID() {
		for _, element := range cb.Transactions {
			cbTxs = append(cbTxs, element)
		}
		for _, element := range nb.Transactions {
			nbTxs = append(nbTxs, element)
		}

		if val, ok := base.blocks[cb.PrevBlockHash]; ok {
			cb = val
		} else {
			cb = *new(Block)
		}
		nb = base.blocks[nb.PrevBlockHash]
	}

	for _, element := range nbTxs {
		indexInCbTxs := indexOf(element, cbTxs)
		if indexInCbTxs != -1 {
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

func (base Miner) postTransaction(outputs map[string]int) bool {
	tx := base.Client.postTransaction(outputs, DEFAULT_TX_FEE)
	return base.addTransaction(tx)
}
