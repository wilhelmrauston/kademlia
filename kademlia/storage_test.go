package kademlia

import (
	"testing"
	"time"
)

func TestNodeStoreAndRetrieve(t *testing.T) {
	config := DefaultConfig()
	node := NewNode("127.0.0.1:8000", "", config)

	testData := "Hello, Kademlia!"

	// Test storing data
	key, err := node.Store(testData)
	if err != nil {
		t.Fatalf("Failed to store data: %v", err)
	}

	if key == "" {
		t.Error("Store should return a non-empty key")
	}

	// Test retrieving data locally
	value, found := node.GetValue(key)
	if !found {
		t.Error("Should find stored value locally")
	}

	if value != testData {
		t.Errorf("Expected %s, got %s", testData, value)
	}
}

func TestDataHashing(t *testing.T) {
	testData1 := "Hello, World!"
	testData2 := "Different data"

	hash1 := HashData(testData1)
	hash2 := HashData(testData2)

	if hash1 == "" || hash2 == "" {
		t.Error("HashData should return non-empty hashes")
	}

	if hash1 == hash2 {
		t.Error("Different data should produce different hashes")
	}

	// Test that same data produces same hash
	hash1Again := HashData(testData1)
	if hash1 != hash1Again {
		t.Error("Same data should produce same hash")
	}

	// Hash should be hex string
	if len(hash1) != 40 { // SHA-1 produces 40-character hex string
		t.Errorf("Expected 40-character hash, got %d characters", len(hash1))
	}
}

func TestStoreMessageHandler(t *testing.T) {
	config := DefaultConfig()
	node := NewNode("127.0.0.1:8000", "", config)
	handler := node.handler

	// Create sender
	senderID := NewRandomKademliaID()
	sender := NewContact(senderID, "127.0.0.1:8001")

	testKey := "test_key"
	testValue := "test_value"

	// Create STORE message
	storeMsg := Message{
		Type:      STORE,
		MessageID: "store123",
		Sender:    sender,
		Timestamp: time.Now().Unix(),
		Data: StoreData{
			Key:   testKey,
			Value: testValue,
		},
	}

	// Handle the message
	response, err := handler.HandleMessage(storeMsg, "127.0.0.1:8001")
	if err != nil {
		t.Fatalf("Failed to handle STORE message: %v", err)
	}

	// STORE messages typically don't return responses
	if response.Type != 0 {
		t.Error("STORE handler should not return a response message")
	}

	// Verify data was stored
	value, found := node.GetValue(testKey)
	if !found {
		t.Error("Value should be stored after STORE message")
	}

	if value != testValue {
		t.Errorf("Expected %s, got %s", testValue, value)
	}
}

func TestFindValueMessageHandler(t *testing.T) {
	config := DefaultConfig()
	node := NewNode("127.0.0.1:8000", "", config)
	handler := node.handler

	// Store a value first
	testKey := "find_test_key"
	testValue := "find_test_value"
	node.StoreValue(testKey, testValue)

	// Create sender
	senderID := NewRandomKademliaID()
	sender := NewContact(senderID, "127.0.0.1:8001")

	// Create FIND_VALUE message
	findValueMsg := Message{
		Type:      FIND_VALUE,
		MessageID: "find123",
		Sender:    sender,
		Timestamp: time.Now().Unix(),
		Data: FindValueData{
			Key: testKey,
		},
	}

	// Handle the message
	response, err := handler.HandleMessage(findValueMsg, "127.0.0.1:8001")
	if err != nil {
		t.Fatalf("Failed to handle FIND_VALUE message: %v", err)
	}

	if response.Type != FIND_VALUE {
		t.Errorf("Expected FIND_VALUE response, got %s", response.Type)
	}

	// Check that we got a FindValueResponse
	responseData, ok := response.Data.(FindValueResponse)
	if !ok {
		t.Error("Response data should be FindValueResponse type")
	}

	if !responseData.Found {
		t.Error("Should find the stored value")
	}

	if responseData.Value != testValue {
		t.Errorf("Expected %s, got %s", testValue, responseData.Value)
	}
}

func TestFindValueMessageHandlerNotFound(t *testing.T) {
	config := DefaultConfig()
	node := NewNode("127.0.0.1:8000", "", config)
	handler := node.handler

	// Add some contacts to routing table for testing
	for i := 0; i < 3; i++ {
		contactID := NewRandomKademliaID()
		contact := NewContact(contactID, "127.0.0.1:800"+string(rune('1'+i)))
		node.routingTable.AddContact(contact)
	}

	// Create sender
	senderID := NewRandomKademliaID()
	sender := NewContact(senderID, "127.0.0.1:8001")

	// Create FIND_VALUE message for non-existent key
	nonExistentKey := HashData("non_existent_data") // Use proper hash
	findValueMsg := Message{
		Type:      FIND_VALUE,
		MessageID: "find456",
		Sender:    sender,
		Timestamp: time.Now().Unix(),
		Data: FindValueData{
			Key: nonExistentKey,
		},
	}

	// Handle the message
	response, err := handler.HandleMessage(findValueMsg, "127.0.0.1:8001")
	if err != nil {
		t.Fatalf("Failed to handle FIND_VALUE message: %v", err)
	}

	if response.Type != FIND_VALUE {
		t.Errorf("Expected FIND_VALUE response, got %s", response.Type)
	}

	// Check that we got a FindValueResponse with contacts
	responseData, ok := response.Data.(FindValueResponse)
	if !ok {
		t.Error("Response data should be FindValueResponse type")
	}

	if responseData.Found {
		t.Error("Should not find the non-existent value")
	}

	if len(responseData.Contacts) == 0 {
		t.Error("Should return closest contacts when value not found")
	}
}

func TestConcurrentStorage(t *testing.T) {
	config := DefaultConfig()
	node := NewNode("127.0.0.1:8000", "", config)

	// Test concurrent storage operations
	done := make(chan bool, 10)

	for i := 0; i < 10; i++ {
		go func(index int) {
			key := HashData("concurrent_test_" + string(rune('0'+index)))
			value := "value_" + string(rune('0'+index))
			node.StoreValue(key, value)

			// Verify it was stored
			retrievedValue, found := node.GetValue(key)
			if !found || retrievedValue != value {
				t.Errorf("Concurrent storage failed for index %d", index)
			}
			done <- true
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}
}
