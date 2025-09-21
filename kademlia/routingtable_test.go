package kademlia

import "testing"

func TestNewRoutingTable(t *testing.T) {
	id := NewRandomKademliaID()
	me := NewContact(id, "127.0.0.1:8000")
	rt := NewRoutingTable(me)
	
	if rt == nil {
		t.Fatal("NewRoutingTable returned nil")
	}
	
	if !rt.me.ID.Equals(id) {
		t.Error("Routing table 'me' contact does not match")
	}
}

func TestRoutingTableAddContact(t *testing.T) {
	id := NewRandomKademliaID()
	me := NewContact(id, "127.0.0.1:8000")
	rt := NewRoutingTable(me)
	
	// Add a contact
	contactID := NewRandomKademliaID()
	contact := NewContact(contactID, "127.0.0.1:8001")
	rt.AddContact(contact)
	
	// Verify it was added
	allContacts := rt.GetAllContacts()
	if len(allContacts) != 1 {
		t.Errorf("Expected 1 contact, got %d", len(allContacts))
	}
	
	if !allContacts[0].ID.Equals(contactID) {
		t.Error("Added contact ID does not match")
	}
}

func TestRoutingTableFindClosestContacts(t *testing.T) {
	id := NewRandomKademliaID()
	me := NewContact(id, "127.0.0.1:8000")
	rt := NewRoutingTable(me)
	
	// Add several contacts
	var addedContacts []Contact
	for i := 0; i < 10; i++ {
		contactID := NewRandomKademliaID()
		contact := NewContact(contactID, "127.0.0.1:800"+string(rune('1'+i)))
		rt.AddContact(contact)
		addedContacts = append(addedContacts, contact)
	}
	
	target := NewRandomKademliaID()
	closest := rt.FindClosestContacts(target, 3)
	
	if len(closest) == 0 {
		t.Error("Should return some contacts")
	}
	
	if len(closest) > 3 {
		t.Errorf("Should return at most 3 contacts, got %d", len(closest))
	}
}