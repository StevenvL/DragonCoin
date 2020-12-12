package main

//const BigInteger = require('jsbn').BigInteger;
import (
	"fmt"
	"math/big"
)

// Network message constants
const MISSING_BLOCK string = "MISSING_BLOCK"
const POST_TRANSACTION string = "POST_TRANSACTION"
const PROOF_FOUND string = "PROOF_FOUND"
const START_MINING string = "START_MINING"

// Constants for mining
const NUM_ROUNDS_MINING int = 2000

// Constants related to proof-of-work target
var POW_BASE_TARGET = big.NewInt(0)

const POW_LEADING_ZEROES uint = 15

// Constants for mining rewards and default transaction fees
const COINBASE_AMT_ALLOWED int = 25
const DEFAULT_TX_FEE int = 1

// If a block is 6 blocks older than the current block, it is considered
// confirmed, for no better reason than that is what Bitcoin does.
// Note that the genesis block is always considered to be confirmed.
const CONFIRMED_DEPTH int = 6

/*
func main() {
	/*
		if _, ok := POW_BASE_TARGET.SetString("ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff", 16); ok {
			POW_BASE_TARGET.Rsh(POW_BASE_TARGET, POW_LEADING_ZEROES)
			fmt.Printf("number = %v\n", POW_BASE_TARGET)
		} else {
			fmt.Printf("rip")
		}
	block := new(Block)
	fmt.Printf(block.getCharacter())
}
*/

type BlockChaincfg struct {
	powTarget        *big.Int
	blockClass       Block
	transactionClass Transaction
	coinbaseAmount   int
	defaultTxFee     int
	confirmedDepth   int
}

type BlockChain struct {
	//blockClass       Block
	cfg              *BlockChaincfg
	transactionClass Transaction
	powLeadingZeroes uint
	coinbaseAmount   int
	defaultTxFee     int
	confirmedDepth   int
	//clientBalanceMap map
	//startingBalances map
}

func newBlockchain() *BlockChain {
	blockchain := new(BlockChain)
	blockchain.powLeadingZeroes = POW_LEADING_ZEROES
	blockchain.coinbaseAmount = COINBASE_AMT_ALLOWED
	blockchain.defaultTxFee = DEFAULT_TX_FEE
	blockchain.confirmedDepth = CONFIRMED_DEPTH
	blockchain.cfg = new(BlockChaincfg)
	return blockchain
}

func makeGenesis(blockClass Block, transactionClass Transaction, clientBalanceMap map[string]int, clientAddrMap map[string]*Client, blockchain *BlockChain) *Block {

	//if (clientBalanceMap && startingBalances) {
	//  throw new Error("You may set clientBalanceMap OR set startingBalances, but not both.");
	//}

	// Setting blockchain configuration
	blockchain.cfg.blockClass = blockClass
	blockchain.cfg.transactionClass = transactionClass
	blockchain.cfg.coinbaseAmount = COINBASE_AMT_ALLOWED
	blockchain.cfg.defaultTxFee = DEFAULT_TX_FEE
	blockchain.cfg.confirmedDepth = CONFIRMED_DEPTH
	if _, ok := POW_BASE_TARGET.SetString("ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff", 16); ok {
	} else {
		fmt.Printf("rip")
	}
	blockchain.cfg.powTarget = POW_BASE_TARGET.Rsh(POW_BASE_TARGET, POW_LEADING_ZEROES)
	//fmt.Println(blockchain.cfg.powTarget)

	// If startingBalances was specified, we initialize our balances to that object.
	//BlockChain.balances = startingBalances //|| {};

	// If clientBalanceMap was initialized instead, we copy over those values.
	/*
		if (clientBalanceMap !== undefined) {
		  for (let [client, balance] of clientBalanceMap.entries()) {
			balances[client.address] = balance;
		  }
		}*/
	//fmt.Printf("%+v\n", blockchain.cfg)
	g := blockchain.makeEmptyBlock()

	// Initializing starting balances in the genesis block.
	for address, balance := range clientBalanceMap {
		g.Balances[address] = balance
	}

	// If clientBalanceMap was specified, we set the genesis block for every client.

	for _, client := range clientAddrMap {
		client.setGenesisBlock(*g)
	}

	return g
}

/**
 * Converts a string representation of a block to a new Block instance.
 *
 * @param {Object} o - An object representing a block, but not necessarily an instance of Block.
 *
 * @returns {Block}
 */
func deserializeBlock(o Block) Block {
	//if reflect.TypeOf(o) == reflect.TypeOf(blockchain.cfg.blockClass) {
	return o
	//}

	/*
		b := blockchain.cfg.blockClass.newBlock(blockchain)
		b.chainLength, _ = strconv.ParseInt(o.chainLength, 10, 32)
		b.timestamp = o.timestamp

		if b.isGenesisBlock() {
			// Balances need to be recreated and restored in a map.
			for clientID, amount := range o.balances {
				b.balances[clientID] = amount
			}
		} else {
			b.prevBlockHash = o.prevBlockHash
			b.proof = o.proof
			b.rewardAddr = o.rewardAddr
			// Likewise, transactions need to be recreated and restored in a map.
			b.transactions = make(map[string]Transaction)
			if len(o.transactions) != 0 {
				for txID, txJson := range o.transactions {
					//tx := blockchain.cfg.transactionClass.newTransaction(txJson) NO JSON SUPPORT
					tx := txJson
					b.transactions[txID] = tx
				}
			}
		}

		return b
	*/
}

func (blockchain *BlockChain) makeEmptyBlock() *Block {
	//fmt.Printf("%+v\n", blockchain.cfg)
	return blockchain.cfg.blockClass.emptyBlock(*blockchain)
}

/* This function is useless.
func (blockchain BlockChain) makeTransaction(o Transaction) Transaction {
	if reflect.TypeOf(o) == reflect.TypeOf(blockchain.cfg.transactionClass) {
		return o
	} /*else { //CAN't HANDLE ANONYMOUS OBJECTS
		return blockchain.cfg.transactionClass.newTransaction(o)
	}
	return o

}*/
/*
These are all useless, everything here is publically available.

func (blockchain BlockChain) getPOW_TARGET() *big.Int {
	return blockchain.cfg.powTarget
}
func (blockchain BlockChain) getCOINBASE_AMT_ALLOWED() int {
	return blockchain.cfg.coinbaseAmount
}
func (blockchain BlockChain) getDEFAULT_TX_FEE() int {
	return blockchain.cfg.defaultTxFee
}
func (blockchain BlockChain) getCONFIRMED_DEPTH() int {
	return blockchain.cfg.confirmedDepth
}

func (blockchain BlockChain) MISSING_BLOCK() string {
	return MISSING_BLOCK
}

func (blockchain BlockChain) POST_TRANSACTION() string {
	return POST_TRANSACTION
}

func (blockchain BlockChain) PROOF_FOUND() string {
	return PROOF_FOUND
}

func (blockchain BlockChain) START_MINING() string {
	return START_MINING
}

func (blockchain BlockChain) NUM_ROUNDS_MINING() int {
	return NUM_ROUNDS_MINING
}
*/
