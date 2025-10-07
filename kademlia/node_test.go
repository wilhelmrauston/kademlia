package kademlia

import (
	"testing"
	"time"
)

func TestNewNode(t *testing.T) {
	config := DefaultConfig()
	address := "127.0.0.1:8000"
	nodeID := ""

	node := NewNode(address, nodeID, config)

	if node == nil {
		t.Fatal("NewNode returned nil")
	}

	if node.Address != address {
		t.Errorf("Expected address %s, got %s", address, node.Address)
	}

	if node.ID == nil {
		t.Error("Node ID should not be nil")
	}

	// Test components are initialized
	if node.routingTable == nil {
		t.Error("Routing table should be initialized")
	}

	if node.dataStore == nil {
		t.Error("Data store should be initialized")
	}

	if node.transport == nil {
		t.Error("Transport should be initialized")
	}

	if node.handler == nil {
		t.Error("Message handler should be initialized")
	}

	if node.config == nil {
		t.Error("Config should be initialized")
	}
}

func TestNewNodeWithSpecificID(t *testing.T) {
	config := DefaultConfig()
	address := "127.0.0.1:8000"
	nodeID := "1234567890abcdef1234567890abcdef12345678"

	node := NewNode(address, nodeID, config)

	if node.ID.String() != nodeID {
		t.Errorf("Expected node ID %s, got %s", nodeID, node.ID.String())
	}
}

func TestNodeStartStop(t *testing.T) {
	config := DefaultConfig()
	address := "127.0.0.1:0" // Use port 0 to let OS assign available port
	node := NewNode(address, "", config)

	// Test Start
	err := node.Start()
	if err != nil {
		t.Fatalf("Failed to start node: %v", err)
	}

	// Give it a moment to start
	time.Sleep(100 * time.Millisecond)

	// Test Stop
	err = node.Stop()
	if err != nil {
		t.Errorf("Failed to stop node: %v", err)
	}
}

func TestDataStore(t *testing.T) {
	ds := NewDataStore()

	// Test empty store
	_, exists := ds.Get("nonexistent")
	if exists {
		t.Error("Should not find nonexistent key")
	}

	keys := ds.GetAllKeys()
	if len(keys) != 0 {
		t.Errorf("Empty store should have 0 keys, got %d", len(keys))
	}

	// Test store and retrieve
	ds.Store("key1", "value1")
	value, exists := ds.Get("key1")
	if !exists {
		t.Error("Should find stored key")
	}
	if value != "value1" {
		t.Errorf("Expected value1, got %s", value)
	}

	// Test multiple keys
	ds.Store("key2", "value2")
	ds.Store("key3", "value3")

	keys = ds.GetAllKeys()
	if len(keys) != 3 {
		t.Errorf("Expected 3 keys, got %d", len(keys))
	}

	// Test overwrite
	ds.Store("key1", "newvalue1")
	value, _ = ds.Get("key1")
	if value != "newvalue1" {
		t.Errorf("Expected newvalue1, got %s", value)
	}
}

func TestHashKey(t *testing.T) {
	// Test consistent hashing
	key := "test_key"
	hash1 := HashKey(key)
	hash2 := HashKey(key)

	if !hash1.Equals(hash2) {
		t.Error("HashKey should be deterministic")
	}

	// Test different keys produce different hashes
	hash3 := HashKey("different_key")
	if hash1.Equals(hash3) {
		t.Error("Different keys should produce different hashes")
	}

	// Test hash length (SHA-1 produces 160-bit/20-byte hash)
	hashStr := hash1.String()
	if len(hashStr) != 40 { // 20 bytes = 40 hex characters
		t.Errorf("Hash should be 40 hex characters, got %d", len(hashStr))
	}
}

func TestGenerateMessageID(t *testing.T) {
	id1 := generateMessageID()
	id2 := generateMessageID()

	if id1 == id2 {
		t.Error("Message IDs should be unique")
	}

	if id1 == "" || id2 == "" {
		t.Error("Message IDs should not be empty")
	}
}

func TestNodeGetters(t *testing.T) {
	config := DefaultConfig()
	address := "127.0.0.1:8000"
	node := NewNode(address, "", config)

	// Test GetRoutingTable
	rt := node.GetRoutingTable()
	if rt == nil {
		t.Error("GetRoutingTable should return non-nil routing table")
	}

	// Test GetDataStore
	ds := node.GetDataStore()
	if ds == nil {
		t.Error("GetDataStore should return non-nil data store")
	}

	// Test GetRoutingTableInfo
	info := node.GetRoutingTableInfo()
	if info == "" {
		t.Error("GetRoutingTableInfo should return non-empty string")
	}
}

func TestDataStoreConcurrency(t *testing.T) {
	ds := NewDataStore()

	// Test concurrent writes
	done := make(chan bool, 10)

	for i := 0; i < 10; i++ {
		go func(id int) {
			key := "key" + string(rune('0'+id))
			value := "value" + string(rune('0'+id))
			ds.Store(key, value)
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	// Test concurrent reads
	for i := 0; i < 10; i++ {
		go func(id int) {
			key := "key" + string(rune('0'+id))
			expectedValue := "value" + string(rune('0'+id))

			value, exists := ds.Get(key)
			if !exists {
				t.Errorf("Key %s should exist", key)
				return
			}
			if value != expectedValue {
				t.Errorf("Expected %s, got %s", expectedValue, value)
				return
			}
			done <- true
		}(i)
	}

	// Wait for all read goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	// Verify all keys exist
	keys := ds.GetAllKeys()
	if len(keys) != 10 {
		t.Errorf("Expected 10 keys after concurrent operations, got %d", len(keys))
	}
}

func TestNodeConfiguration(t *testing.T) {
	config := &Config{
		K:           10,
		Alpha:       2,
		IDLength:    20,
		PingTimeout: 3 * time.Second,
		JoinTimeout: 15 * time.Second,
	}

	node := NewNode("127.0.0.1:8000", "", config)

	if node.config.K != 10 {
		t.Errorf("Expected K=10, got %d", node.config.K)
	}

	if node.config.Alpha != 2 {
		t.Errorf("Expected Alpha=2, got %d", node.config.Alpha)
	}
}

func TestNodeIDGeneration(t *testing.T) {
	config := DefaultConfig()

	// Test random ID generation
	node1 := NewNode("127.0.0.1:8001", "", config)
	node2 := NewNode("127.0.0.1:8002", "", config)

	if node1.ID.Equals(node2.ID) {
		t.Error("Random node IDs should be different")
	}

	// Test specific ID
	specificID := "abcdef1234567890abcdef1234567890abcdef12"
	node3 := NewNode("127.0.0.1:8003", specificID, config)

	if node3.ID.String() != specificID {
		t.Errorf("Expected specific ID %s, got %s", specificID, node3.ID.String())
	}
}

func TestDataStoreEdgeCases(t *testing.T) {
	ds := NewDataStore()

	// Test empty key
	ds.Store("", "empty_key_value")
	value, exists := ds.Get("")
	if !exists || value != "empty_key_value" {
		t.Error("Should handle empty key")
	}

	// Test empty value
	ds.Store("empty_value", "")
	value, exists = ds.Get("empty_value")
	if !exists || value != "" {
		t.Error("Should handle empty value")
	}

	// Test large value
	largeValue := string(make([]byte, 10000))
	ds.Store("large", largeValue)
	value, exists = ds.Get("large")
	if !exists || len(value) != 10000 {
		t.Error("Should handle large values")
	}

	// Test unicode
	ds.Store("unicode", "🌟🚀💫")
	value, exists = ds.Get("unicode")
	if !exists || value != "🌟🚀💫" {
		t.Error("Should handle unicode values")
	}
}

func TestSendPing(t *testing.T) {
	config := DefaultConfig()
	address := "127.0.0.1:0" // Use port 0 for automatic port assignment
	node := NewNode(address, "", config)

	err := node.Start()
	if err != nil {
		t.Fatalf("Failed to start node: %v", err)
	}
	defer node.Stop()

	// Test sending ping
	err = node.SendPing("127.0.0.1:9999") // Non-existent target
	// We expect this to not return an error immediately (it's async)
	if err != nil {
		t.Errorf("SendPing should not return immediate error: %v", err)
	}
}

func TestSendFindNode(t *testing.T) {
	config := DefaultConfig()
	address := "127.0.0.1:0"
	node := NewNode(address, "", config)

	err := node.Start()
	if err != nil {
		t.Fatalf("Failed to start node: %v", err)
	}
	defer node.Stop()

	// Test sending find node
	targetID := NewRandomKademliaID()
	err = node.SendFindNode("127.0.0.1:9999", targetID) // Non-existent target
	// We expect this to not return an error immediately (it's async)
	if err != nil {
		t.Errorf("SendFindNode should not return immediate error: %v", err)
	}
}

func TestPerformSelfLookup(t *testing.T) {
	config := DefaultConfig()
	address := "127.0.0.1:8000"
	node := NewNode(address, "", config)

	// Add some contacts to routing table first
	for i := 0; i < 3; i++ {
		contactID := NewRandomKademliaID()
		contact := NewContact(contactID, "127.0.0.1:800"+string(rune('1'+i)))
		node.routingTable.AddContact(contact)
	}

	// This should not panic or error, even if no actual network calls succeed
	node.PerformSelfLookup()

	// The function should complete without errors
	// (It will fail to contact nodes but that's expected in test)
}

func TestMergeAndSortContacts(t *testing.T) {
	config := DefaultConfig()
	address := "127.0.0.1:8000"
	node := NewNode(address, "", config)

	targetID := NewRandomKademliaID()

	// Create existing contacts
	var existing []Contact
	for i := 0; i < 3; i++ {
		contactID := NewRandomKademliaID()
		contact := NewContact(contactID, "127.0.0.1:800"+string(rune('1'+i)))
		existing = append(existing, contact)
	}

	// Create new contacts
	var new []Contact
	for i := 0; i < 2; i++ {
		contactID := NewRandomKademliaID()
		contact := NewContact(contactID, "127.0.0.1:900"+string(rune('1'+i)))
		new = append(new, contact)
	}

	// Test merge and sort
	result := node.mergeAndSortContacts(existing, new, targetID, 4)

	if len(result) > 4 {
		t.Errorf("Result should be limited to maxCount, got %d", len(result))
	}

	if len(result) == 0 {
		t.Error("Result should contain some contacts")
	}

	// Verify contacts are sorted by distance (closest first)
	for i := 0; i < len(result)-1; i++ {
		if result[i].distance == nil || result[i+1].distance == nil {
			t.Error("Contacts should have distance calculated")
			continue
		}
		if !result[i].distance.Less(result[i+1].distance) && !result[i].distance.Equals(result[i+1].distance) {
			t.Error("Contacts should be sorted by distance")
		}
	}
}

func TestMergeAndSortContactsWithDuplicates(t *testing.T) {
	config := DefaultConfig()
	address := "127.0.0.1:8000"
	node := NewNode(address, "", config)

	targetID := NewRandomKademliaID()

	// Create contacts with some duplicates
	contactID := NewRandomKademliaID()
	contact1 := NewContact(contactID, "127.0.0.1:8001")
	contact2 := NewContact(contactID, "127.0.0.1:8001") // Duplicate

	existing := []Contact{contact1}
	new := []Contact{contact2}

	result := node.mergeAndSortContacts(existing, new, targetID, 10)

	if len(result) != 1 {
		t.Errorf("Duplicates should be merged, expected 1 contact, got %d", len(result))
	}
}

func TestQueryNodeWithTimeout(t *testing.T) {
	config := DefaultConfig()
	address := "127.0.0.1:8000"
	node := NewNode(address, "", config)

	contact := NewContact(NewRandomKademliaID(), "127.0.0.1:9999") // Non-existent
	targetID := NewRandomKademliaID()
	timeout := 1 * time.Second

	// This should return empty result due to timeout/failure
	result := node.queryNodeWithTimeout(contact, targetID, timeout)

	if len(result) != 0 {
		t.Errorf("Expected empty result for failed query, got %d contacts", len(result))
	}
}

func TestMinFunction(t *testing.T) {
	// Test min function
	result := min(5, 10)
	if result != 5 {
		t.Errorf("min(5, 10) should return 5, got %d", result)
	}

	result = min(10, 5)
	if result != 5 {
		t.Errorf("min(10, 5) should return 5, got %d", result)
	}

	result = min(7, 7)
	if result != 7 {
		t.Errorf("min(7, 7) should return 7, got %d", result)
	}

	result = min(0, 1)
	if result != 0 {
		t.Errorf("min(0, 1) should return 0, got %d", result)
	}
}

func TestParseFindValueResponse(t *testing.T) {
	// Test with valid response data
	responseData := FindValueResponse{
		Found: true,
		Value: "test_value",
	}

	result, err := parseFindValueResponse(responseData)
	if err != nil {
		t.Fatalf("parseFindValueResponse failed: %v", err)
	}

	if !result.Found {
		t.Error("Response should indicate found")
	}

	if result.Value != "test_value" {
		t.Errorf("Expected value 'test_value', got '%s'", result.Value)
	}

	// Test with contacts
	contacts := []Contact{
		NewContact(NewRandomKademliaID(), "127.0.0.1:8001"),
	}

	responseData2 := FindValueResponse{
		Found:    false,
		Contacts: contacts,
	}

	result2, err := parseFindValueResponse(responseData2)
	if err != nil {
		t.Fatalf("parseFindValueResponse failed: %v", err)
	}

	if result2.Found {
		t.Error("Response should indicate not found")
	}

	if len(result2.Contacts) != 1 {
		t.Errorf("Expected 1 contact, got %d", len(result2.Contacts))
	}
}

func TestParseFindValueResponseInvalidData(t *testing.T) {
	// Test with invalid data that can't be marshaled
	invalidData := make(chan int) // Channels can't be marshaled to JSON

	_, err := parseFindValueResponse(invalidData)
	if err == nil {
		t.Error("Should return error for invalid data")
	}
}
