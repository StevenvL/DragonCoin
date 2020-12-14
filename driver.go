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
}

func endSimulation() {
	fmt.Println("Simulation ended")
}
