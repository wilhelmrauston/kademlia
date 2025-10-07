package kademlia

import (
	"sync"
	"testing"
	"time"
)

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
	for i := 0; i < 10; i++ {
		contactID := NewRandomKademliaID()
		contact := NewContact(contactID, "127.0.0.1:800"+string(rune('1'+i)))
		rt.AddContact(contact)
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

func TestRoutingTableBucketDistribution(t *testing.T) {
	id := NewRandomKademliaID()
	me := NewContact(id, "127.0.0.1:8000")
	rt := NewRoutingTable(me)

	// Add many contacts to test bucket distribution
	contactCount := 50
	for i := 0; i < contactCount; i++ {
		contactID := NewRandomKademliaID()
		contact := NewContact(contactID, "127.0.0.1:8000")
		rt.AddContact(contact)
	}

	allContacts := rt.GetAllContacts()
	if len(allContacts) == 0 {
		t.Error("Should have added contacts")
	}

	// Should have added all contacts (assuming they go to different buckets)
	if len(allContacts) > contactCount {
		t.Errorf("Should not have more contacts than added: got %d, added %d", len(allContacts), contactCount)
	}
}

func TestRoutingTableGetBucketIndex(t *testing.T) {
	id := NewRandomKademliaID()
	me := NewContact(id, "127.0.0.1:8000")
	rt := NewRoutingTable(me)

	// Test with same ID (should go to last bucket)
	bucketIndex := rt.getBucketIndex(id)
	expectedIndex := IDLength*8 - 1
	if bucketIndex != expectedIndex {
		t.Errorf("Same ID should go to bucket %d, got %d", expectedIndex, bucketIndex)
	}

	// Test with different ID
	differentID := NewRandomKademliaID()
	differentIndex := rt.getBucketIndex(differentID)
	if differentIndex < 0 || differentIndex >= IDLength*8 {
		t.Errorf("Bucket index out of range: %d", differentIndex)
	}
}

func TestRoutingTableFindClosestContactsOrdering(t *testing.T) {
	id := NewRandomKademliaID()
	me := NewContact(id, "127.0.0.1:8000")
	rt := NewRoutingTable(me)

	// Create a specific target and add contacts at known distances
	target := NewRandomKademliaID()

	// Add several contacts
	for i := 0; i < 10; i++ {
		contactID := NewRandomKademliaID()
		contact := NewContact(contactID, "127.0.0.1:8000")
		rt.AddContact(contact)
	}

	// Find closest contacts
	closest := rt.FindClosestContacts(target, 5)

	// Verify ordering (each contact should be closer or equal distance than the next)
	for i := 0; i < len(closest)-1; i++ {
		dist1 := target.CalcDistance(closest[i].ID)
		dist2 := target.CalcDistance(closest[i+1].ID)

		// Compare distances (dist1 should be <= dist2)
		if dist1.Less(dist2) == false && !dist1.Equals(dist2) {
			t.Error("Contacts are not properly ordered by distance")
		}
	}
}

func TestRoutingTableEmptyBehavior(t *testing.T) {
	id := NewRandomKademliaID()
	me := NewContact(id, "127.0.0.1:8000")
	rt := NewRoutingTable(me)

	// Test with empty routing table
	target := NewRandomKademliaID()
	closest := rt.FindClosestContacts(target, 3)

	if len(closest) != 0 {
		t.Errorf("Empty routing table should return 0 contacts, got %d", len(closest))
	}

	allContacts := rt.GetAllContacts()
	if len(allContacts) != 0 {
		t.Errorf("Empty routing table should return 0 contacts, got %d", len(allContacts))
	}
}

func TestRoutingTableConcurrentAccess(t *testing.T) {
	id := NewRandomKademliaID()
	me := NewContact(id, "127.0.0.1:8000")
	rt := NewRoutingTable(me)

	var wg sync.WaitGroup
	numGoroutines := 10
	contactsPerGoroutine := 10

	// Test concurrent AddContact operations
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(routineID int) {
			defer wg.Done()
			for j := 0; j < contactsPerGoroutine; j++ {
				contactID := NewRandomKademliaID()
				contact := NewContact(contactID, "127.0.0.1:8000")
				rt.AddContact(contact)
			}
		}(i)
	}

	// Test concurrent FindClosestContacts operations
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			target := NewRandomKademliaID()
			for j := 0; j < 5; j++ {
				rt.FindClosestContacts(target, 3)
				time.Sleep(time.Millisecond) // Small delay to allow interleaving
			}
		}()
	}

	// Test concurrent GetAllContacts operations
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 10; j++ {
				rt.GetAllContacts()
				time.Sleep(time.Millisecond)
			}
		}()
	}

	wg.Wait()

	// Verify final state
	allContacts := rt.GetAllContacts()
	expectedCount := numGoroutines * contactsPerGoroutine
	if len(allContacts) == 0 {
		t.Error("Should have added contacts concurrently")
	}

	// Note: We might have fewer contacts than expected due to bucket size limits
	// and duplicate contacts, but we should have some
	t.Logf("Added %d contacts concurrently, final count: %d", expectedCount, len(allContacts))
}

func TestRoutingTableMaxContacts(t *testing.T) {
	id := NewRandomKademliaID()
	me := NewContact(id, "127.0.0.1:8000")
	rt := NewRoutingTable(me)

	// Test requesting more contacts than available
	for i := 0; i < 5; i++ {
		contactID := NewRandomKademliaID()
		contact := NewContact(contactID, "127.0.0.1:8000")
		rt.AddContact(contact)
	}

	target := NewRandomKademliaID()

	// Request more contacts than we have
	closest := rt.FindClosestContacts(target, 20)
	if len(closest) > 5 {
		t.Errorf("Should not return more contacts than available, got %d", len(closest))
	}

	// Request exactly what we have
	closest = rt.FindClosestContacts(target, 5)
	if len(closest) > 5 {
		t.Errorf("Should not return more than 5 contacts, got %d", len(closest))
	}
}
