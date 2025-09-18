package kademlia

import "testing"

func TestKademliaMessageHandlerPing(t *testing.T) {
	config := DefaultConfig()
	nodeID := NewRandomKademliaID()
	nodeAddress := "127.0.0.1:8000"
	me := NewContact(nodeID, nodeAddress)
	rt := NewRoutingTable(me)
	
	handler := NewKademliaMessageHandler(rt, nodeID, nodeAddress, config)
	
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
	
	// Add some contacts to routing table
	for i := 0; i < 5; i++ {
		contactID := NewRandomKademliaID()
		contact := NewContact(contactID, "127.0.0.1:800"+string(rune('1'+i)))
		rt.AddContact(contact)
	}
	
	handler := NewKademliaMessageHandler(rt, nodeID, nodeAddress, config)
	
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
	
	if len(responseData.Contacts) > config.Alpha {
		t.Errorf("Response should contain at most %d contacts, got %d", config.Alpha, len(responseData.Contacts))
	}
}