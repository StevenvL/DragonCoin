package main

import (
	"encoding/json"
	"fmt"
	"reflect"
)

//Map of client address to client obj
type FakeNet struct {
	clients map[string]Client
}

func newFakeNet() *FakeNet {
	fakeNet := new(FakeNet)
	return fakeNet
}

//Takes in an array of clients to register
func (base FakeNet) register(clientList []Client) {
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
	for address, _ := range base.clients {
		base.sendMessage(address, message, jsonObject)
	}
}

//TODO
//UNSURE IF THIS EMITS TO THE RIGHT CLIENT OR EVERY CLIENT
func (base FakeNet) sendMessage(address string, message string, jsonObject []byte) {
	//curClient := base.clients[address]
	var block Block
	err := json.Unmarshal(jsonObject, &block)
	fmt.Printf(`Error code in sendMessage is %s`, err)
	emitter.Emit(message, block)
}

func (base FakeNet) recognizes(client Client) bool {
	for address, client := range base.clients {
		if reflect.DeepEqual(base.clients[address], client) {
			return true
		}
	}
	return false
}