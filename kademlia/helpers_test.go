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

	// LookupNode calls FindClosestContacts with K parameter (20), not Alpha (3)
	if len(result) > DefaultConfig().K {
		t.Errorf("LookupNode should return at most %d contacts, got %d", DefaultConfig().K, len(result))
	}

	// Since we added 10 contacts, we should get all of them back (max 10)
	if len(result) > 10 {
		t.Errorf("Should not return more contacts than we added, got %d", len(result))
	}
}
