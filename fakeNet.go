package main

//Map of client address to client obj
type FakeNet struct {
	clients map[string]Client
}
