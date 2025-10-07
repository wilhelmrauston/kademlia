package kademlia

import "testing"

func TestKademliaMessageHandlerPing(t *testing.T) {
	config := DefaultConfig()
	nodeID := NewRandomKademliaID()
	nodeAddress := "127.0.0.1:8000"
	me := NewContact(nodeID, nodeAddress)
	rt := NewRoutingTable(me)
	ds := NewDataStore()

	handler := NewKademliaMessageHandler(rt, nodeID, nodeAddress, config, ds)

	// Create a ping message
	senderID := NewRandomKademliaID()
	sender := NewContact(senderID, "127.0.0.1:8001")

	pingMsg := Message{
		Type:      PING,
		MessageID: "test123",
		Sender:    sender,
		Timestamp: 12345,
		Data: PingData{
			Message: "ping",
		},
	}

	response, err := handler.HandleMessage(pingMsg, "127.0.0.1:8001")
	if err != nil {
		t.Fatalf("HandleMessage failed: %v", err)
	}

	if response.Type != PONG {
		t.Errorf("Expected PONG response, got %s", response.Type)
	}

	if response.MessageID != "test123" {
		t.Error("Response should have same MessageID as request")
	}

	// Verify sender was added to routing table
	allContacts := rt.GetAllContacts()
	found := false
	for _, contact := range allContacts {
		if contact.ID.Equals(senderID) {
			found = true
			break
		}
	}
	if !found {
		t.Error("Sender should have been added to routing table")
	}
}

func TestKademliaMessageHandlerFindNode(t *testing.T) {
	config := DefaultConfig()
	nodeID := NewRandomKademliaID()
	nodeAddress := "127.0.0.1:8000"
	me := NewContact(nodeID, nodeAddress)
	rt := NewRoutingTable(me)
	ds := NewDataStore()

	// Add some contacts to routing table
	for i := 0; i < 5; i++ {
		contactID := NewRandomKademliaID()
		contact := NewContact(contactID, "127.0.0.1:800"+string(rune('1'+i)))
		rt.AddContact(contact)
	}

	handler := NewKademliaMessageHandler(rt, nodeID, nodeAddress, config, ds)

	// Create a find node message
	senderID := NewRandomKademliaID()
	sender := NewContact(senderID, "127.0.0.1:8001")
	targetID := NewRandomKademliaID()

	findNodeMsg := Message{
		Type:      FIND_NODE,
		MessageID: "test456",
		Sender:    sender,
		Timestamp: 12345,
		Data: FindNodeData{
			TargetID: targetID,
		},
	}

	response, err := handler.HandleMessage(findNodeMsg, "127.0.0.1:8001")
	if err != nil {
		t.Fatalf("HandleMessage failed: %v", err)
	}

	if response.Type != FIND_NODE_RESPONSE {
		t.Errorf("Expected FIND_NODE_RESPONSE, got %s", response.Type)
	}

	// Check response data
	responseData, ok := response.Data.(FindNodeResponse)
	if !ok {
		t.Error("Response data should be FindNodeResponse type")
	}

	if len(responseData.Contacts) == 0 {
		t.Error("Response should contain some contacts")
	}

	// FIND_NODE should return up to K contacts (not Alpha)
	if len(responseData.Contacts) > config.K {
		t.Errorf("Response should contain at most %d contacts, got %d", config.K, len(responseData.Contacts))
	}

	// We added 5 contacts, plus the sender gets added automatically
	// So we expect at most 6 contacts (5 original + 1 sender)
	if len(responseData.Contacts) > 6 {
		t.Errorf("Should not return more than 6 contacts (5 added + sender), got %d", len(responseData.Contacts))
	}
}

func TestKademliaMessageHandlerPong(t *testing.T) {
	config := DefaultConfig()
	nodeID := NewRandomKademliaID()
	nodeAddress := "127.0.0.1:8000"
	me := NewContact(nodeID, nodeAddress)
	rt := NewRoutingTable(me)
	ds := NewDataStore()

	handler := NewKademliaMessageHandler(rt, nodeID, nodeAddress, config, ds)

	// Create a pong message
	senderID := NewRandomKademliaID()
	sender := NewContact(senderID, "127.0.0.1:8001")

	pongMsg := Message{
		Type:      PONG,
		MessageID: "test789",
		Sender:    sender,
		Timestamp: 12345,
		Data: PongData{
			Message: "pong",
		},
	}

	response, err := handler.HandleMessage(pongMsg, "127.0.0.1:8001")
	if err != nil {
		t.Fatalf("HandleMessage failed: %v", err)
	}

	// PONG handler should return empty message
	if response.Type != MessageType(0) {
		t.Errorf("PONG handler should return empty message type")
	}

	// Verify sender was added to routing table
	allContacts := rt.GetAllContacts()
	found := false
	for _, contact := range allContacts {
		if contact.ID.Equals(senderID) {
			found = true
			break
		}
	}
	if !found {
		t.Error("Sender should have been added to routing table")
	}
}

func TestKademliaMessageHandlerFindNodeResponse(t *testing.T) {
	config := DefaultConfig()
	nodeID := NewRandomKademliaID()
	nodeAddress := "127.0.0.1:8000"
	me := NewContact(nodeID, nodeAddress)
	rt := NewRoutingTable(me)
	ds := NewDataStore()

	handler := NewKademliaMessageHandler(rt, nodeID, nodeAddress, config, ds)

	// Create contacts to return in response
	var returnedContacts []Contact
	for i := 0; i < 3; i++ {
		contactID := NewRandomKademliaID()
		contact := NewContact(contactID, "127.0.0.1:900"+string(rune('1'+i)))
		returnedContacts = append(returnedContacts, contact)
	}

	// Create a find node response message
	senderID := NewRandomKademliaID()
	sender := NewContact(senderID, "127.0.0.1:8001")

	findNodeResponseMsg := Message{
		Type:      FIND_NODE_RESPONSE,
		MessageID: "test456",
		Sender:    sender,
		Timestamp: 12345,
		Data: FindNodeResponse{
			Contacts: returnedContacts,
		},
	}

	response, err := handler.HandleMessage(findNodeResponseMsg, "127.0.0.1:8001")
	if err != nil {
		t.Fatalf("HandleMessage failed: %v", err)
	}

	// Response handler should return empty message
	if response.Type != MessageType(0) {
		t.Errorf("Response handler should return empty message type")
	}

	// Verify returned contacts were added to routing table
	allContacts := rt.GetAllContacts()

	// Should have sender + 3 returned contacts = 4 total
	if len(allContacts) < 4 {
		t.Errorf("Expected at least 4 contacts (sender + 3 returned), got %d", len(allContacts))
	}

	// Verify specific contacts were added
	for _, expectedContact := range returnedContacts {
		found := false
		for _, contact := range allContacts {
			if contact.ID.Equals(expectedContact.ID) {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Contact %s should have been added to routing table", expectedContact.ID.String()[:8])
		}
	}
}

func TestKademliaMessageHandlerFindValue(t *testing.T) {
	config := DefaultConfig()
	nodeID := NewRandomKademliaID()
	nodeAddress := "127.0.0.1:8000"
	me := NewContact(nodeID, nodeAddress)
	rt := NewRoutingTable(me)
	ds := NewDataStore()

	// Store a value in data store
	testKey := "test_key"
	testValue := "test_value"
	ds.Store(testKey, testValue)

	handler := NewKademliaMessageHandler(rt, nodeID, nodeAddress, config, ds)

	// Create a find value message for existing key
	senderID := NewRandomKademliaID()
	sender := NewContact(senderID, "127.0.0.1:8001")

	findValueMsg := Message{
		Type:      FIND_VALUE,
		MessageID: "test123",
		Sender:    sender,
		Timestamp: 12345,
		Data: FindValueData{
			Key: testKey,
		},
	}

	response, err := handler.HandleMessage(findValueMsg, "127.0.0.1:8001")
	if err != nil {
		t.Fatalf("HandleMessage failed: %v", err)
	}

	if response.Type != FIND_VALUE_RESPONSE {
		t.Errorf("Expected FIND_VALUE_RESPONSE, got %s", response.Type)
	}

	// Check response data - should contain the value
	responseData, ok := response.Data.(FindValueResponse)
	if !ok {
		t.Fatal("Response data should be FindValueResponse type")
	}

	if !responseData.Found {
		t.Error("Response should indicate value was found")
	}

	if responseData.Value != testValue {
		t.Errorf("Expected value %s, got %s", testValue, responseData.Value)
	}

	if responseData.Contacts != nil {
		t.Error("When value is found, contacts should be nil")
	}
}

func TestKademliaMessageHandlerFindValueNotFound(t *testing.T) {
	config := DefaultConfig()
	nodeID := NewRandomKademliaID()
	nodeAddress := "127.0.0.1:8000"
	me := NewContact(nodeID, nodeAddress)
	rt := NewRoutingTable(me)
	ds := NewDataStore()

	// Add some contacts to routing table
	for i := 0; i < 3; i++ {
		contactID := NewRandomKademliaID()
		contact := NewContact(contactID, "127.0.0.1:800"+string(rune('1'+i)))
		rt.AddContact(contact)
	}

	handler := NewKademliaMessageHandler(rt, nodeID, nodeAddress, config, ds)

	// Create a find value message for non-existing key
	senderID := NewRandomKademliaID()
	sender := NewContact(senderID, "127.0.0.1:8001")

	findValueMsg := Message{
		Type:      FIND_VALUE,
		MessageID: "test456",
		Sender:    sender,
		Timestamp: 12345,
		Data: FindValueData{
			Key: "nonexistent_key",
		},
	}

	response, err := handler.HandleMessage(findValueMsg, "127.0.0.1:8001")
	if err != nil {
		t.Fatalf("HandleMessage failed: %v", err)
	}

	if response.Type != FIND_VALUE_RESPONSE {
		t.Errorf("Expected FIND_VALUE_RESPONSE, got %s", response.Type)
	}

	// Check response data - should contain contacts
	responseData, ok := response.Data.(FindValueResponse)
	if !ok {
		t.Fatal("Response data should be FindValueResponse type")
	}

	if responseData.Found {
		t.Error("Response should indicate value was not found")
	}

	if responseData.Value != "" {
		t.Error("When value is not found, value should be empty")
	}

	if len(responseData.Contacts) == 0 {
		t.Error("When value is not found, should return closest contacts")
	}
}

func TestKademliaMessageHandlerFindValueResponse(t *testing.T) {
	config := DefaultConfig()
	nodeID := NewRandomKademliaID()
	nodeAddress := "127.0.0.1:8000"
	me := NewContact(nodeID, nodeAddress)
	rt := NewRoutingTable(me)
	ds := NewDataStore()

	handler := NewKademliaMessageHandler(rt, nodeID, nodeAddress, config, ds)

	// Test case 1: Response with value found
	senderID := NewRandomKademliaID()
	sender := NewContact(senderID, "127.0.0.1:8001")

	findValueResponseMsg := Message{
		Type:      FIND_VALUE_RESPONSE,
		MessageID: "test123",
		Sender:    sender,
		Timestamp: 12345,
		Data: FindValueResponse{
			Found: true,
			Value: "found_value",
		},
	}

	response, err := handler.HandleMessage(findValueResponseMsg, "127.0.0.1:8001")
	if err != nil {
		t.Fatalf("HandleMessage failed: %v", err)
	}

	// Response handler should return empty message
	if response.Type != MessageType(0) {
		t.Errorf("Response handler should return empty message type")
	}

	// Test case 2: Response with contacts (value not found)
	var returnedContacts []Contact
	for i := 0; i < 2; i++ {
		contactID := NewRandomKademliaID()
		contact := NewContact(contactID, "127.0.0.1:900"+string(rune('1'+i)))
		returnedContacts = append(returnedContacts, contact)
	}

	findValueResponseMsg2 := Message{
		Type:      FIND_VALUE_RESPONSE,
		MessageID: "test456",
		Sender:    sender,
		Timestamp: 12345,
		Data: FindValueResponse{
			Found:    false,
			Contacts: returnedContacts,
		},
	}

	response, err = handler.HandleMessage(findValueResponseMsg2, "127.0.0.1:8001")
	if err != nil {
		t.Fatalf("HandleMessage failed: %v", err)
	}

	// Verify returned contacts were added to routing table
	allContacts := rt.GetAllContacts()

	// Should have sender + 2 returned contacts = 3 total
	if len(allContacts) < 3 {
		t.Errorf("Expected at least 3 contacts (sender + 2 returned), got %d", len(allContacts))
	}
}

func TestKademliaMessageHandlerStore(t *testing.T) {
	config := DefaultConfig()
	nodeID := NewRandomKademliaID()
	nodeAddress := "127.0.0.1:8000"
	me := NewContact(nodeID, nodeAddress)
	rt := NewRoutingTable(me)
	ds := NewDataStore()

	handler := NewKademliaMessageHandler(rt, nodeID, nodeAddress, config, ds)

	// Create a store message
	senderID := NewRandomKademliaID()
	sender := NewContact(senderID, "127.0.0.1:8001")

	storeMsg := Message{
		Type:      STORE,
		MessageID: "test123",
		Sender:    sender,
		Timestamp: 12345,
		Data: StoreData{
			Key:   "store_key",
			Value: "store_value",
		},
	}

	response, err := handler.HandleMessage(storeMsg, "127.0.0.1:8001")
	if err != nil {
		t.Fatalf("HandleMessage failed: %v", err)
	}

	if response.Type != STORE_RESPONSE {
		t.Errorf("Expected STORE_RESPONSE, got %s", response.Type)
	}

	if response.MessageID != "test123" {
		t.Error("Response should have same MessageID as request")
	}

	// Check response data
	responseData, ok := response.Data.(StoreResponse)
	if !ok {
		t.Fatal("Response data should be StoreResponse type")
	}

	if responseData.Status != "success" {
		t.Errorf("Expected status 'success', got '%s'", responseData.Status)
	}

	// Verify the key-value pair was stored
	value, exists := ds.Get("store_key")
	if !exists {
		t.Error("Key should have been stored in data store")
	}

	if value != "store_value" {
		t.Errorf("Expected stored value 'store_value', got '%s'", value)
	}
}

func TestKademliaMessageHandlerStoreResponse(t *testing.T) {
	config := DefaultConfig()
	nodeID := NewRandomKademliaID()
	nodeAddress := "127.0.0.1:8000"
	me := NewContact(nodeID, nodeAddress)
	rt := NewRoutingTable(me)
	ds := NewDataStore()

	handler := NewKademliaMessageHandler(rt, nodeID, nodeAddress, config, ds)

	// Create a store response message
	senderID := NewRandomKademliaID()
	sender := NewContact(senderID, "127.0.0.1:8001")

	storeResponseMsg := Message{
		Type:      STORE_RESPONSE,
		MessageID: "test123",
		Sender:    sender,
		Timestamp: 12345,
		Data: StoreResponse{
			Status: "success",
		},
	}

	response, err := handler.HandleMessage(storeResponseMsg, "127.0.0.1:8001")
	if err != nil {
		t.Fatalf("HandleMessage failed: %v", err)
	}

	// Response handler should return empty message
	if response.Type != MessageType(0) {
		t.Errorf("Response handler should return empty message type")
	}
}

func TestKademliaMessageHandlerUnknownMessageType(t *testing.T) {
	config := DefaultConfig()
	nodeID := NewRandomKademliaID()
	nodeAddress := "127.0.0.1:8000"
	me := NewContact(nodeID, nodeAddress)
	rt := NewRoutingTable(me)
	ds := NewDataStore()

	handler := NewKademliaMessageHandler(rt, nodeID, nodeAddress, config, ds)

	// Create a message with unknown type
	senderID := NewRandomKademliaID()
	sender := NewContact(senderID, "127.0.0.1:8001")

	unknownMsg := Message{
		Type:      MessageType(999), // Unknown type
		MessageID: "test123",
		Sender:    sender,
		Timestamp: 12345,
		Data:      nil,
	}

	_, err := handler.HandleMessage(unknownMsg, "127.0.0.1:8001")
	if err == nil {
		t.Error("Should return error for unknown message type")
	}

	expectedError := "unknown message type: 999"
	if err.Error() != expectedError {
		t.Errorf("Expected error '%s', got '%s'", expectedError, err.Error())
	}
}

func TestKademliaMessageHandlerFindValueErrors(t *testing.T) {
	config := DefaultConfig()
	nodeID := NewRandomKademliaID()
	nodeAddress := "127.0.0.1:8000"
	me := NewContact(nodeID, nodeAddress)
	rt := NewRoutingTable(me)
	ds := NewDataStore()

	handler := NewKademliaMessageHandler(rt, nodeID, nodeAddress, config, ds)

	// Test case 1: Missing key in FindValueData
	senderID := NewRandomKademliaID()
	sender := NewContact(senderID, "127.0.0.1:8001")

	// Create message with map data that has no key
	findValueMsg := Message{
		Type:      FIND_VALUE,
		MessageID: "test123",
		Sender:    sender,
		Timestamp: 12345,
		Data:      map[string]interface{}{}, // Empty map, no key
	}

	_, err := handler.HandleMessage(findValueMsg, "127.0.0.1:8001")
	if err == nil {
		t.Error("Should return error for missing key")
	}

	expectedError := "FIND_VALUE: missing key"
	if err.Error() != expectedError {
		t.Errorf("Expected error '%s', got '%s'", expectedError, err.Error())
	}
}
