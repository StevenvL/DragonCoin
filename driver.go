package main

import (
	"fmt"
	"time"
)

func main() {
	fmt.Println("Starting simulation.  This may take a moment...")

	fakeNet := newFakeNet()

	// Clients
	emptyKeys := keypair{}
	emptyBlock := Block{}
	emptyTransaction := Transaction{}
	//fmt.Println(emptyBlock.empty)
	alice := newClient("Alice", emptyKeys, emptyBlock, fakeNet)
	bob := newClient("Bob", emptyKeys, emptyBlock, fakeNet)
	charlie := newClient("Charlie", emptyKeys, emptyBlock, fakeNet)

	// Miners
	minnie := newMiner("Minnie", emptyKeys, emptyBlock, fakeNet)
	mickey := newMiner("Mickey", emptyKeys, emptyBlock, fakeNet)

	// Creating genesis block
	blockchain := newBlockchain()
	balanceMap := map[string]int{
		alice.address:         233,
		bob.address:           99,
		charlie.address:       67,
		minnie.Client.address: 400,
		mickey.Client.address: 300,
	}
	addrMap := map[string]*Client{alice.address: alice, bob.address: bob, charlie.address: charlie, minnie.Client.address: minnie.Client, mickey.Client.address: mickey.Client}
	genesis := makeGenesis(
		emptyBlock,
		emptyTransaction,
		balanceMap,
		addrMap,
		blockchain,
	)
	//fmt.Println(minnie.Client.lastBlock)
	//fmt.Println(balanceMap)
	//fmt.Printf("%+v\n", genesis)
	//fmt.Printf("%+v\n", alice)
	//alice.setGenesisBlock(*genesis)
	//fmt.Println(alice.lastBlock)

	// Late miner - Donald has more mining power, represented by the miningRounds.
	// (Mickey and Minnie have the default of 2000 rounds).
	donald := newMiner("Donald", emptyKeys, *genesis, fakeNet)
	donald.miningRounds = 3000

	showBalances := func(client Client) {
		fmt.Printf("Alice has  %v gold.\n", client.lastBlock.balanceOf(alice.address))
		fmt.Printf("Bob has  %v gold.\n", client.lastBlock.balanceOf(bob.address))
		fmt.Printf("Charlie has  %v gold.\n", client.lastBlock.balanceOf(charlie.address))
		fmt.Printf("Minnie has  %v gold.\n", client.lastBlock.balanceOf(minnie.address))
		fmt.Printf("Mickey has %v gold.\n", client.lastBlock.balanceOf(mickey.address))
		fmt.Printf("Donald has %v gold.\n", client.lastBlock.balanceOf(donald.address))
	}

	// Showing the initial balances from Alice's perspective, for no particular reason.
	fmt.Println("Initial balances:")
	showBalances(*alice)
	//fmt.Println(alice.availableGold())
	clientList := []*Client{alice, bob, charlie, minnie.Client, mickey.Client}
	fakeNet.register(clientList)

	// Miners start mining.
	go minnie.initialize()
	go mickey.initialize()

	// Alice transfers some money to Bob.

	fmt.Printf("Alice is transfering 40 gold to %v\n", bob.address)
	alice.postTransaction(map[string]int{bob.address: 40}, DEFAULT_TX_FEE)
	time.Sleep(10 * time.Second)
	fmt.Println()
	fmt.Printf("Minnie has a chain of length %v:", minnie.Client.lastBlock.ChainLength)

	fmt.Println()
	fmt.Printf("Mickey has a chain of length %v:", mickey.Client.lastBlock.ChainLength)

	fmt.Println()
	fmt.Println("Final balances (Minnie's perspective):")
	showBalances(*minnie.Client)

	fmt.Println()
	fmt.Println("Final balances (Mickey's perspective):")
	showBalances(*mickey.Client)

	fmt.Println()
	fmt.Println("Final balances (Alice's perspective):")
	showBalances(*alice)

	/*
	  setTimeout(() => {
	    fmt.Println()
	    fmt.Println("***Starting a late-to-the-party miner***")
	    fmt.Println()
	    fakeNet.register(donald)
	    donald.initialize()
	  }, 2000)
	*/

	//time.AfterFunc(10*time.Second, endSimulation)
	//time.Sleep(11 * time.Second)
	/*
	  // Print out the final balances after it has been running for some time.
	  setTimeout(() => {
	    fmt.Println()
	    fmt.Println(`Minnie has a chain of length ${minnie.currentBlock.chainLength}:`)

	    fmt.Println()
	    fmt.Println(`Mickey has a chain of length ${mickey.currentBlock.chainLength}:`)

	    fmt.Println()
	    fmt.Println(`Donald has a chain of length ${donald.currentBlock.chainLength}:`)

	    fmt.Println()
	    fmt.Println("Final balances (Minnie's perspective):")
	    showBalances(minnie)

	    fmt.Println()
	    fmt.Println("Final balances (Alice's perspective):")
	    showBalances(alice)

	    fmt.Println()
	    fmt.Println("Final balances (Donald's perspective):")
	    showBalances(donald)

	    process.exit(0)
	  }, 5000)
	*/
}

func endSimulation() {
	fmt.Println("Simulation ended")
}
