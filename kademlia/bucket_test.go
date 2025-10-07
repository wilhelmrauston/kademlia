package kademlia

import (
	"testing"
)

func TestNewBucket(t *testing.T) {
	bucket := newBucket()

	if bucket == nil {
		t.Fatal("newBucket returned nil")
	}

	if bucket.list == nil {
		t.Error("bucket list should be initialized")
	}

	if bucket.Len() != 0 {
		t.Errorf("new bucket should be empty, got length %d", bucket.Len())
	}
}

func TestBucketAddContact(t *testing.T) {
	bucket := newBucket()

	// Add first contact
	contact1 := NewContact(NewRandomKademliaID(), "127.0.0.1:8001")
	bucket.AddContact(contact1)

	if bucket.Len() != 1 {
		t.Errorf("Expected bucket length 1, got %d", bucket.Len())
	}

	// Add second contact
	contact2 := NewContact(NewRandomKademliaID(), "127.0.0.1:8002")
	bucket.AddContact(contact2)

	if bucket.Len() != 2 {
		t.Errorf("Expected bucket length 2, got %d", bucket.Len())
	}

	// Verify order: most recent contact should be at front
	front := bucket.list.Front().Value.(Contact)
	if !front.ID.Equals(contact2.ID) {
		t.Error("Most recently added contact should be at front of bucket")
	}
}

func TestBucketAddDuplicateContact(t *testing.T) {
	bucket := newBucket()

	contact := NewContact(NewRandomKademliaID(), "127.0.0.1:8001")

	// Add contact first time
	bucket.AddContact(contact)
	if bucket.Len() != 1 {
		t.Errorf("Expected bucket length 1, got %d", bucket.Len())
	}

	// Add same contact again
	bucket.AddContact(contact)
	if bucket.Len() != 1 {
		t.Errorf("Adding duplicate contact should not increase bucket size, got %d", bucket.Len())
	}

	// Verify the contact is still at the front (moved to front)
	front := bucket.list.Front().Value.(Contact)
	if !front.ID.Equals(contact.ID) {
		t.Error("Duplicate contact should be moved to front")
	}
}

func TestBucketAddContactWithUpdate(t *testing.T) {
	bucket := newBucket()

	// Add multiple contacts
	contact1 := NewContact(NewRandomKademliaID(), "127.0.0.1:8001")
	contact2 := NewContact(NewRandomKademliaID(), "127.0.0.1:8002")
	contact3 := NewContact(NewRandomKademliaID(), "127.0.0.1:8003")

	bucket.AddContact(contact1)
	bucket.AddContact(contact2)
	bucket.AddContact(contact3)

	// At this point: [contact3, contact2, contact1] (front to back)

	// Re-add contact1 (should move to front)
	bucket.AddContact(contact1)

	// Verify contact1 is now at front
	front := bucket.list.Front().Value.(Contact)
	if !front.ID.Equals(contact1.ID) {
		t.Error("Re-added contact should be moved to front")
	}

	// Verify bucket size didn't change
	if bucket.Len() != 3 {
		t.Errorf("Bucket size should remain 3, got %d", bucket.Len())
	}
}

func TestBucketSizeLimit(t *testing.T) {
	bucket := newBucket()

	// Add contacts up to bucket size limit
	var contacts []Contact
	for i := 0; i < bucketSize; i++ {
		contact := NewContact(NewRandomKademliaID(), "127.0.0.1:8000")
		contacts = append(contacts, contact)
		bucket.AddContact(contact)
	}

	if bucket.Len() != bucketSize {
		t.Errorf("Expected bucket length %d, got %d", bucketSize, bucket.Len())
	}

	// Try to add one more contact (should be ignored since bucket is full)
	extraContact := NewContact(NewRandomKademliaID(), "127.0.0.1:9000")
	bucket.AddContact(extraContact)

	// Bucket size should remain the same
	if bucket.Len() != bucketSize {
		t.Errorf("Bucket should not exceed size limit, got %d", bucket.Len())
	}

	// Verify the extra contact was not added
	found := false
	for elt := bucket.list.Front(); elt != nil; elt = elt.Next() {
		contact := elt.Value.(Contact)
		if contact.ID.Equals(extraContact.ID) {
			found = true
			break
		}
	}
	if found {
		t.Error("Extra contact should not be added when bucket is full")
	}
}

func TestBucketGetContactAndCalcDistance(t *testing.T) {
	bucket := newBucket()

	// Add some contacts
	contact1 := NewContact(NewRandomKademliaID(), "127.0.0.1:8001")
	contact2 := NewContact(NewRandomKademliaID(), "127.0.0.1:8002")
	contact3 := NewContact(NewRandomKademliaID(), "127.0.0.1:8003")

	bucket.AddContact(contact1)
	bucket.AddContact(contact2)
	bucket.AddContact(contact3)

	target := NewRandomKademliaID()
	contacts := bucket.GetContactAndCalcDistance(target)

	if len(contacts) != 3 {
		t.Errorf("Expected 3 contacts, got %d", len(contacts))
	}

	// Verify all contacts have distance calculated
	for i, contact := range contacts {
		if contact.distance == nil {
			t.Errorf("Contact %d should have distance calculated", i)
		}

		// Verify distance is correct
		expectedDistance := contact.ID.CalcDistance(target)
		if !contact.distance.Equals(expectedDistance) {
			t.Errorf("Contact %d has incorrect distance", i)
		}
	}

	// Verify order is maintained (front to back)
	if !contacts[0].ID.Equals(contact3.ID) {
		t.Error("First contact should be contact3 (most recently added)")
	}
	if !contacts[1].ID.Equals(contact2.ID) {
		t.Error("Second contact should be contact2")
	}
	if !contacts[2].ID.Equals(contact1.ID) {
		t.Error("Third contact should be contact1 (least recently added)")
	}
}

func TestBucketGetContactAndCalcDistanceEmpty(t *testing.T) {
	bucket := newBucket()
	target := NewRandomKademliaID()

	contacts := bucket.GetContactAndCalcDistance(target)

	if len(contacts) != 0 {
		t.Errorf("Empty bucket should return empty contact list, got %d contacts", len(contacts))
	}
}

func TestBucketLen(t *testing.T) {
	bucket := newBucket()

	// Test empty bucket
	if bucket.Len() != 0 {
		t.Errorf("Empty bucket should have length 0, got %d", bucket.Len())
	}

	// Add contacts and verify length
	for i := 1; i <= 5; i++ {
		contact := NewContact(NewRandomKademliaID(), "127.0.0.1:8000")
		bucket.AddContact(contact)

		if bucket.Len() != i {
			t.Errorf("After adding %d contacts, expected length %d, got %d", i, i, bucket.Len())
		}
	}
}

func TestBucketContactOrdering(t *testing.T) {
	bucket := newBucket()

	// Add contacts in order
	contact1 := NewContact(NewRandomKademliaID(), "127.0.0.1:8001")
	contact2 := NewContact(NewRandomKademliaID(), "127.0.0.1:8002")
	contact3 := NewContact(NewRandomKademliaID(), "127.0.0.1:8003")

	bucket.AddContact(contact1)
	bucket.AddContact(contact2)
	bucket.AddContact(contact3)

	// Verify order: most recently added should be first
	target := NewRandomKademliaID()
	contacts := bucket.GetContactAndCalcDistance(target)

	expectedOrder := []Contact{contact3, contact2, contact1}
	for i, expected := range expectedOrder {
		if !contacts[i].ID.Equals(expected.ID) {
			t.Errorf("Contact at position %d should be %s, got %s",
				i, expected.ID.String()[:8], contacts[i].ID.String()[:8])
		}
	}
}

func TestBucketFullScenario(t *testing.T) {
	bucket := newBucket()

	// Fill bucket to capacity
	var addedContacts []Contact
	for i := 0; i < bucketSize; i++ {
		contact := NewContact(NewRandomKademliaID(), "127.0.0.1:8000")
		addedContacts = append(addedContacts, contact)
		bucket.AddContact(contact)
	}

	// Verify bucket is full
	if bucket.Len() != bucketSize {
		t.Errorf("Bucket should be full with %d contacts, got %d", bucketSize, bucket.Len())
	}

	// Try to add existing contact (should move to front)
	existingContact := addedContacts[5] // Pick a contact from middle
	bucket.AddContact(existingContact)

	// Bucket size should remain same
	if bucket.Len() != bucketSize {
		t.Errorf("Bucket size should remain %d after re-adding existing contact, got %d",
			bucketSize, bucket.Len())
	}

	// Existing contact should now be at front
	front := bucket.list.Front().Value.(Contact)
	if !front.ID.Equals(existingContact.ID) {
		t.Error("Re-added existing contact should be moved to front")
	}
}
