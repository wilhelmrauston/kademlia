package kademlia

import "testing"

func TestNewContact(t *testing.T) {
	id := NewRandomKademliaID()
	address := "127.0.0.1:8000"
	
	contact := NewContact(id, address)
	
	if !contact.ID.Equals(id) {
		t.Error("Contact ID does not match")
	}
	
	if contact.Address != address {
		t.Error("Contact address does not match")
	}
	
	if contact.distance != nil {
		t.Error("New contact should have nil distance")
	}
}

func TestContactDistance(t *testing.T) {
	id1 := NewKademliaID("1234567890abcdef1234567890abcdef12345678")
	id2 := NewKademliaID("fedcba0987654321fedcba0987654321fedcba09")
	target := NewKademliaID("aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa")
	
	contact1 := NewContact(id1, "127.0.0.1:8001")
	contact2 := NewContact(id2, "127.0.0.1:8002")
	
	contact1.CalcDistance(target)
	contact2.CalcDistance(target)
	
	if contact1.distance == nil || contact2.distance == nil {
		t.Error("Distance should be calculated")
	}
}

func TestContactCandidates(t *testing.T) {
	target := NewKademliaID("aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa")
	
	var candidates ContactCandidates
	
	// Add some contacts
	for i := 0; i < 5; i++ {
		id := NewRandomKademliaID()
		contact := NewContact(id, "127.0.0.1:800"+string(rune('0'+i)))
		contact.CalcDistance(target)
		candidates.Append([]Contact{contact})
	}
	
	if candidates.Len() != 5 {
		t.Errorf("Expected 5 contacts, got %d", candidates.Len())
	}
	
	// Test sorting
	candidates.Sort()
	
	// Verify sorted order
	contacts := candidates.GetContacts(5)
	for i := 1; i < len(contacts); i++ {
		if contacts[i-1].distance.Less(contacts[i].distance) {
			// This is correct - should not fail
		} else if contacts[i].distance.Less(contacts[i-1].distance) {
			t.Error("Contacts are not properly sorted")
		}
	}
}