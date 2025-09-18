package kademlia

import "testing"

func TestBasicNetworkFlow(t *testing.T) {
	// Create two nodes
	config := DefaultConfig()
	
	node1 := NewNode("127.0.0.1:8001", "", config)
	node2 := NewNode("127.0.0.1:8002", "", config)
	
	// Add node2 to node1's routing table
	node1.routingTable.AddContact(NewContact(node2.ID, node2.Address))
	
	// Test that node1 knows about node2
	contacts := node1.routingTable.FindClosestContacts(node2.ID, 1)
	if len(contacts) == 0 {
		t.Error("Node1 should know about node2")
	}
	
	if !contacts[0].ID.Equals(node2.ID) {
		t.Error("Found contact should be node2")
	}
}

// Benchmark tests
func BenchmarkKademliaIDDistance(b *testing.B) {
	id1 := NewRandomKademliaID()
	id2 := NewRandomKademliaID()
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		id1.CalcDistance(id2)
	}
}

func BenchmarkRoutingTableAddContact(b *testing.B) {
	nodeID := NewRandomKademliaID()
	me := NewContact(nodeID, "127.0.0.1:8000")
	rt := NewRoutingTable(me)
	
	contacts := make([]Contact, b.N)
	for i := 0; i < b.N; i++ {
		id := NewRandomKademliaID()
		contacts[i] = NewContact(id, "127.0.0.1:8001")
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rt.AddContact(contacts[i])
	}
}