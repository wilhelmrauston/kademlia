package kademlia

import (
	"crypto/sha1"
	"encoding/hex"
)


type Kademlia struct {
	me Contact
	network *Network
	rt *RoutingTable
	k int
}

func newKademliaNode(address string, me string) (kademlia Kademlia){
	sha1.Sum([]byte(me))
	sha1 := sha1.Sum([]byte(me))
	key := hex.EncodeToString(sha1[:])
	id := NewKademliaID(key)
	kademlia.me = NewContact(id, address)
	kademlia.rt = NewRoutingTable(kademlia.me)
	kademlia.network = &Network{&kademlia}
	return
}

func (kademlia *Kademlia) LookupContact(target *Contact) {
	// TODO
}

func (kademlia *Kademlia) LookupData(hash string) {
	// TODO
}

func (kademlia *Kademlia) Store(data []byte) {
	// TODO
}