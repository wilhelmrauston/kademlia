package kademlia

import "testing"

func TestLookupNode(t *testing.T) {
	// Create routing table with some contacts
	nodeID := NewRandomKademliaID()
	me := NewContact(nodeID, "127.0.0.1:8000")
	rt := NewRoutingTable(me)
	
	// Add contacts
	for i := 0; i < 10; i++ {
		contactID := NewRandomKademliaID()
		contact := NewContact(contactID, "127.0.0.1:800"+string(rune('1'+i)))
		rt.AddContact(contact)
	}
	
	targetID := NewRandomKademliaID()
	result := LookupNode(rt, targetID)
	
	if len(result) == 0 {
		t.Error("LookupNode should return some contacts")
	}
	
	if len(result) > DefaultConfig().Alpha {
		t.Errorf("LookupNode should return at most %d contacts, got %d", DefaultConfig().Alpha, len(result))
	}
}