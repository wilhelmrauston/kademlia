package kademlia

import (
	"encoding/json"
	"testing"
)

func TestNewKademliaID(t *testing.T) {
	hexStr := "1234567890abcdef1234567890abcdef12345678"
	id := NewKademliaID(hexStr)
	
	if id == nil {
		t.Fatal("NewKademliaID returned nil")
	}
	
	if id.String() != hexStr {
		t.Errorf("Expected %s, got %s", hexStr, id.String())
	}
}

func TestNewRandomKademliaID(t *testing.T) {
	id1 := NewRandomKademliaID()
	id2 := NewRandomKademliaID()
	
	if id1 == nil || id2 == nil {
		t.Fatal("NewRandomKademliaID returned nil")
	}
	
	if id1.Equals(id2) {
		t.Error("Two random IDs should not be equal")
	}
}

func TestKademliaIDDistance(t *testing.T) {
	id1 := NewKademliaID("1234567890abcdef1234567890abcdef12345678")
	id2 := NewKademliaID("1234567890abcdef1234567890abcdef12345678")
	id3 := NewKademliaID("fedcba0987654321fedcba0987654321fedcba09")
	
	// Distance to self should be 0
	dist := id1.CalcDistance(id2)
	for i := 0; i < IDLength; i++ {
		if dist[i] != 0 {
			t.Error("Distance to self should be zero")
		}
	}
	
	// Distance should be symmetric
	dist1 := id1.CalcDistance(id3)
	dist2 := id3.CalcDistance(id1)
	if !dist1.Equals(dist2) {
		t.Error("Distance should be symmetric")
	}
}

func TestKademliaIDJSON(t *testing.T) {
	original := NewKademliaID("1234567890abcdef1234567890abcdef12345678")
	
	// Test marshaling
	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("Failed to marshal KademliaID: %v", err)
	}
	
	// Test unmarshaling
	var unmarshaled KademliaID
	err = json.Unmarshal(data, &unmarshaled)
	if err != nil {
		t.Fatalf("Failed to unmarshal KademliaID: %v", err)
	}
	
	if !original.Equals(&unmarshaled) {
		t.Error("Marshaled and unmarshaled KademliaID are not equal")
	}
}