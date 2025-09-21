package kademlia

import "testing"

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