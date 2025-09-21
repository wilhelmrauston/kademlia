package kademlia

import "testing"

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()
	
	if config == nil {
		t.Fatal("DefaultConfig returned nil")
	}
	
	if config.K <= 0 {
		t.Error("K should be positive")
	}
	
	if config.Alpha <= 0 {
		t.Error("Alpha should be positive")
	}
	
	if config.IDLength != 20 {
		t.Errorf("IDLength should be 20, got %d", config.IDLength)
	}
	
	if config.PingTimeout <= 0 {
		t.Error("PingTimeout should be positive")
	}
}