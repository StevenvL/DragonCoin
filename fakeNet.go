package main

import (
	"encoding/json"
	"fmt"
	"reflect"
)

//Map of client address to client obj
type FakeNet struct {
	clients map[string]*Client
}

func newFakeNet() *FakeNet {
	fakeNet := new(FakeNet)
	fakeNet.clients = make(map[string]*Client)
	return fakeNet
}

//Takes in an array of clients to register
func (base FakeNet) register(clientList []*Client) {
	//fmt.Print(clientList)

	for _, client := range clientList {
		base.clients[client.address] = client
	}
}

/**
 * Broadcasts to all clients within this.clients the message msg and payload o.
 *
 * @param {String} msg - the name of the event being broadcasted (e.g. "PROOF_FOUND")
 * @param {Object} o - payload of the message
 */
func (base FakeNet) broadcast(message string, jsonObject []byte) {
	for address := range base.clients {
		base.sendMessage(address, message, jsonObject)
	}
}

/**
 * Sends message msg and payload o directly to Client name.
 *
 * @param {String} address - the public key address of the client or miner to which to send the message
 * @param {String} msg - the name of the event being broadcasted (e.g. "PROOF_FOUND")
 * @param {Object} o - payload of the message
 */
func (base FakeNet) sendMessage(address string, message string, jsonObject []byte) {
	if message == "POST_TRANSACTION" {
		var tx Transaction
		err := json.Unmarshal(jsonObject, &tx)
		if err != nil {
			fmt.Printf(`Error tx in sendMessage is %s`, err)
		}
		base.clients[address].emitter.Emit(message, tx)
	} else {
		var block Block
		err := json.Unmarshal(jsonObject, &block)
		if err != nil {
			fmt.Printf(`Error block in sendMessage is %s`, err)
		}
		base.clients[address].emitter.Emit(message, block)
	}
}

/**
 * Tests whether a client is registered with the network.
 *
 * @param {Client} client - the client to test for.
 *
 * @returns {boolean} True if the client is already registered.
 */
func (base FakeNet) recognizes(client Client) bool {
	for address, client := range base.clients {
		if reflect.DeepEqual(base.clients[address], client) {
			return true
		}
	}
	return false
}
