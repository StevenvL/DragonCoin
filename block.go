package main

import (
	"encoding/json"
	"fmt"
	"math/big"
)

/**
 * A block is a collection of transactions, with a hash connecting it
 * to a previous block.
 */

type Block struct {
	prevBlockHash  string
	target         *big.Int
	transactions   map[string]Transaction
	balances       map[string]int
	nextNonce      map[string]int
	chainLength    int
	timestamp      string
	rewardAddr     string
	coinbaseReward int
}

/**
 * Creates a new Block.  Note that the previous block will not be stored;
 * instead, its hash value will be maintained in this block.
 *
 * @constructor
 * @param {String} rewardAddr - The address to receive all mining rewards for this block.
 * @param {Block} [prevBlock] - The previous block in the blockchain.
 * @param {Number} [target] - The POW target.  The miner must find a proof that
 *      produces a smaller value when hashed.
 * @param {Number} [coinbaseReward] - The gold that a miner earns for finding a block proof.
 * blockChain BlockChain, target=Blockchain.powTarget, coinbaseReward=Blockchain.cfg.coinBase, rewardAddr, prevBlock
 */
func (base Block) newBlock(blockChain BlockChain, rewardAddr string, prevBlock ...Block) *Block {
	block := new(Block)
	block.target = blockChain.cfg.powTarget
	block.coinbaseReward = blockChain.coinbaseAmount
	block.balances = make(map[string]int)
	block.transactions = make(map[string]Transaction)
	block.nextNonce = make(map[string]int)
	block.chainLength = 0
	block.rewardAddr = rewardAddr

	for argNum, arg := range prevBlock {
		switch argNum {
		case 0:
			block.prevBlockHash = arg.hashVal()
			block.balances = arg.balances
			block.nextNonce = arg.nextNonce
			block.chainLength = arg.chainLength + 1
			if arg.rewardAddr != "" {
				block.balances[arg.rewardAddr] += arg.totalRewards()
			}
		}
	}

	// Adding toJSON methods for transactions and balances, which help with
	// serialization.
	// this.transactions.toJSON = () => {
	//   return JSON.stringify(Array.from(this.transactions.entries()));
	// }
	// this.balances.toJSON = () => {
	//   return JSON.stringify(Array.from(this.balances.entries()));
	// }

	// Used to determine the winner between competing chains.
	// Note that this is a little simplistic -- an attacker
	// could make a long, but low-work chain.  However, this works
	// well enough for us.

	// this.timestamp = Date.now();
	return block
}

func (base Block) emptyBlock() *Block {
	block := new(Block)
	return block
}

/**
 * Determines whether the block is the beginning of the chain.
 *
 * @returns {Boolean} - True if this is the first block in the chain.
 */
func (base Block) isGenesisBlock() bool {
	return base.chainLength == 0
}

/**
 * Returns true if the hash of the block is less than the target
 * proof of work value.
 *
 * @returns {Boolean} - True if the block has a valid proof.
 */
func (base Block) hasValidProof() bool {
	h := sha256hash(base.serialize())
	n := big.NewInt(0)
	if _, ok := n.SetString(h, 16); ok {
	} else {
		fmt.Printf("rip")
	}
	return n.Cmp(base.target) < 0
}

/**
 * Converts a Block into string form.  Some fields are deliberately omitted.
 * Note that Block.deserialize plus block.rerun should restore the block.
 *
 * @returns {String} - The block in JSON format.
 */
func (base Block) serialize() string {
	b, _ := json.Marshal(base)
	return string(b)
	//return JSON.stringify(this);
	//if (this.isGenesisBlock()) {
	//  // The genesis block does not contain a proof or transactions,
	//  // but is the only block than can specify balances.
	//  /*******return `
	//     {"chainLength": "${this.chainLength}",
	//      "timestamp": "${this.timestamp}",
	//      "balances": ${JSON.stringify(Array.from(this.balances.entries()))}
	//     }
	//  `;****/
	//  let o = {
	//    chainLength: this.chainLength,
	//    timestamp: this.timestamp,
	//    balances: Array.from(this.balances.entries()),
	//  };
	//  return JSON.stringify(o, ['chainLength', 'timestamp', 'balances']);
	//} else {
	//  // Other blocks must specify transactions and proof details.
	//  /******return `
	//     {"chainLength": "${this.chainLength}",
	//      "timestamp": "${this.timestamp}",
	//      "transactions": ${JSON.stringify(Array.from(this.transactions.entries()))},
	//      "prevBlockHash": "${this.prevBlockHash}",
	//      "proof": "${this.proof}",
	//      "rewardAddr": "${this.rewardAddr}"
	//     }
	//  `;*****/
	//  let o = {
	//    chainLength: this.chainLength,
	//    timestamp: this.timestamp,
	//    transactions: Array.from(this.transactions.entries()),
	//    prevBlockHash: this.prevBlockHash,
	//    proof: this.proof,
	//    rewardAddr: this.rewardAddr,
	//  };
	//  return JSON.stringify(o, ['chainLength', 'timestamp', 'transactions',
	//       'prevBlockHash', 'proof', 'rewardAddr']);
	//}
}

/*
  toJSON() {
    let o = {
      chainLength: this.chainLength,
      timestamp: this.timestamp,
    };
    if (this.isGenesisBlock()) {
      // The genesis block does not contain a proof or transactions,
      // but is the only block than can specify balances.
      o.balances = Array.from(this.balances.entries());
    } else {
      // Other blocks must specify transactions and proof details.
      o.transactions = Array.from(this.transactions.entries());
      o.prevBlockHash = this.prevBlockHash;
      o.proof = this.proof;
      o.rewardAddr = this.rewardAddr;
    }
    return o;
  }*/

/**
 * Returns the cryptographic hash of the current block.
 * The block is first converted to its serial form, so
 * any unimportant fields are ignored.
 *
 * @returns {String} - cryptographic hash of the block.
 */
func (base Block) hashVal() string {
	return sha256hash(base.serialize())
}

/**
 * Returns the hash of the block as its id.
 *
 * @returns {String} - A unique ID for the block.
 */
func (base Block) getID() string {
	return base.hashVal()
}

/**
 * Accepts a new transaction if it is valid and adds it to the block.
 *
 * @param {Transaction} tx - The transaction to add to the block.
 * @param {Client} [client] - A client object, for logging useful messages.
 *
 * @returns {Boolean} - True if the transaction was added successfully.
 */
func (base Block) addTransaction(tx Transaction) bool {
	if _, dupped := base.transactions[tx.id]; dupped {
		fmt.Printf(`Duplicate transaction ${tx.id}.`)
		return false
	}
	if len(tx.sig) == 0 {
		fmt.Printf(`Unsigned transaction ${tx.id}.`)
		return false
	} else if !validSignatureTransaction(tx) {
		fmt.Printf(`Invalid signature for transaction ${tx.id}.`)
		return false
	} else if !tx.sufficientFunds(base) {
		fmt.Printf(`Insufficient gold for transaction ${tx.id}.`)
		return false
	}

	// Checking and updating nonce value.
	// This portion prevents replay attacks.
	nonce := base.nextNonce[tx.from]
	if tx.nonce < nonce {
		fmt.Printf(`Replayed transaction ${tx.id}.`)
		return false
	} else if tx.nonce > nonce {
		// FIXME: Need to do something to handle this case more gracefully.
		fmt.Printf(`Out of order transaction ${tx.id}.`)
		return false
	} else {
		base.nextNonce[tx.from] = nonce + 1
	}

	// Adding the transaction to the block
	base.transactions[tx.id] = tx

	// Taking gold from the sender
	senderBalance := base.balanceOf(tx.from)
	base.balances[tx.from] = senderBalance - totalOutputs(tx)

	// Giving gold to the specified output addresses
	for address, amount := range tx.outputs {
		oldBalance := base.balanceOf(address)
		base.balances[address] = amount + oldBalance
	}

	return true
}

/**
 * When a block is received from another party, it does not include balances or a record of
 * the latest nonces for each client.  This method restores this information be wiping out
 * and re-adding all transactions.  This process also identifies if any transactions were
 * invalid due to insufficient funds or replayed transactions, in which case the block
 * should be rejected.
 *
 * @param {Block} prevBlock - The previous block in the blockchain, used for initial balances.
 *
 * @returns {Boolean} - True if the block's transactions are all valid.
 */
func (base Block) rerun(prevBlock Block) bool {
	// Setting balances to the previous block's balances.
	base.balances = prevBlock.balances
	base.nextNonce = prevBlock.nextNonce

	// Adding coinbase reward for prevBlock.
	winnerBalance := base.balanceOf(prevBlock.rewardAddr)
	if prevBlock.rewardAddr != "" {
		base.balances[prevBlock.rewardAddr] = winnerBalance + prevBlock.totalRewards()
	}

	// Re-adding all transactions.
	txs := base.transactions
	base.transactions = make(map[string]Transaction)
	for _, tx := range txs {
		success := base.addTransaction(tx)
		if !success {
			return false
		}
	}

	return true
}

/**
 * Gets the available gold of a user identified by an address.
 * Note that this amount is a snapshot in time - IF the block is
 * accepted by the network, ignoring any pending transactions,
 * this is the amount of funds available to the client.
 *
 * @param {String} addr - Address of a client.
 *
 * @returns {Number} - The available gold for the specified user.
 */
func (base Block) balanceOf(addr string) int {
	return base.balances[addr]
}

/**
 * The total amount of gold paid to the miner who produced this block,
 * if the block is accepted.  This includes both the coinbase transaction
 * and any transaction fees.
 *
 * @returns {Number} Total reward in gold for the user.
 *
 */
func (base Block) totalRewards() int {
	reward := base.coinbaseReward
	for _, tx := range base.transactions {
		reward += tx.fee
	}
	return reward
}

/**
 * Determines whether a transaction is in the block.  Note that only the
 * block itself is checked; if it returns false, the transaction might
 * still be included in one of its ancestor blocks.
 *
 * @param {Transaction} tx - The transaction that we are checking for.
 *
 * @returns {boolean} - True if the transaction is contained in this block.
 */
func (base Block) contains(tx Transaction) bool {
	if _, ok := base.transactions[tx.id]; ok {
		return true
	} else {
		return false
	}
}
