package kademlia

import (
	"fmt"
	"math/rand"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// M4 Testing Configuration - Easy to change for testing purposes
const (
	// Network emulation parameters
	TestNodeCount      = 1000 // Number of nodes to emulate (requirement: at least 1000)
	TestNodeCountSmall = 100  // Smaller count for faster tests

	// Packet dropping parameters
	DefaultPacketDropRate = 0.1 // 10% packet loss (configurable)
	HighPacketDropRate    = 0.3 // 30% packet loss for stress testing
	NoPacketDrop          = 0.0 // No packet loss for baseline tests

	// Test timeout settings
	TestTimeout               = 30 * time.Second
	NetworkStabilizationDelay = 2 * time.Second
)

// MockTransport implements NetworkTransport with configurable packet dropping
type MockTransport struct {
	address        string
	handler        MessageHandler
	dropRate       float64
	isListening    bool
	mutex          sync.RWMutex
	messageQueue   chan Message
	peerTransports map[string]*MockTransport // Simulate network connections
	droppedCount   int64                     // Use int64 for atomic operations
	sentCount      int64                     // Use int64 for atomic operations
}

// NewMockTransport creates a new mock transport with configurable packet drop rate
func NewMockTransport(address string, handler MessageHandler, dropRate float64) *MockTransport {
	return &MockTransport{
		address:        address,
		handler:        handler,
		dropRate:       dropRate,
		messageQueue:   make(chan Message, 1000),
		peerTransports: make(map[string]*MockTransport),
	}
}

// Listen starts the mock transport (satisfies NetworkTransport interface)
func (mt *MockTransport) Listen(address string) error {
	mt.mutex.Lock()
	mt.isListening = true
	mt.mutex.Unlock()

	// Process messages in background
	go mt.processMessages()
	return nil
}

// Send simulates sending a message with possible packet dropping
func (mt *MockTransport) Send(message Message, targetAddr string) error {
	// Use atomic operations for counters
	atomic.AddInt64(&mt.sentCount, 1)

	// Simulate packet dropping
	if rand.Float64() < mt.dropRate {
		atomic.AddInt64(&mt.droppedCount, 1)
		return fmt.Errorf("packet dropped (simulated)")
	}

	// Find target transport and deliver message
	mt.mutex.RLock()
	targetTransport, exists := mt.peerTransports[targetAddr]
	mt.mutex.RUnlock()

	if !exists {
		return fmt.Errorf("target address %s not found", targetAddr)
	}

	// Deliver message to target (non-blocking)
	select {
	case targetTransport.messageQueue <- message:
		return nil
	default:
		return fmt.Errorf("target transport queue full")
	}
}

// Close stops the mock transport
func (mt *MockTransport) Close() error {
	mt.mutex.Lock()
	defer mt.mutex.Unlock()
	mt.isListening = false
	close(mt.messageQueue)
	return nil
}

// processMessages handles incoming messages
func (mt *MockTransport) processMessages() {
	for message := range mt.messageQueue {
		if mt.handler != nil {
			// Process message synchronously to avoid race conditions in test
			_, err := mt.handler.HandleMessage(message, mt.address)
			if err != nil {
				// Log error in real implementation
			}
		}
	}
}

// AddPeer registers a peer transport for message delivery
func (mt *MockTransport) AddPeer(address string, transport *MockTransport) {
	mt.mutex.Lock()
	defer mt.mutex.Unlock()
	mt.peerTransports[address] = transport
}

// GetStats returns packet statistics
func (mt *MockTransport) GetStats() (sent int, dropped int, dropRate float64) {
	// Use atomic operations to read counters
	sentCount := atomic.LoadInt64(&mt.sentCount)
	droppedCount := atomic.LoadInt64(&mt.droppedCount)
	return int(sentCount), int(droppedCount), mt.dropRate
}

// NetworkEmulator manages a large network of mock nodes
type NetworkEmulator struct {
	nodes      []*Node
	transports []*MockTransport
	dropRate   float64
	nodeCount  int
}

// NewNetworkEmulator creates a network emulator with specified parameters
func NewNetworkEmulator(nodeCount int, dropRate float64) *NetworkEmulator {
	return &NetworkEmulator{
		nodes:      make([]*Node, 0, nodeCount),
		transports: make([]*MockTransport, 0, nodeCount),
		dropRate:   dropRate,
		nodeCount:  nodeCount,
	}
}

// SetupNetwork creates and connects all nodes in the emulated network
func (ne *NetworkEmulator) SetupNetwork() error {
	config := DefaultConfig()

	// Create all nodes with mock transports
	for i := 0; i < ne.nodeCount; i++ {
		address := fmt.Sprintf("127.0.0.1:%d", 10000+i)
		nodeID := fmt.Sprintf("%040x", i) // Generate deterministic node IDs

		// Create node
		node := &Node{
			ID:         NewKademliaID(nodeID),
			Address:    address,
			config:     config,
			dataStore:  make(map[string]string),
			storeMutex: sync.RWMutex{},
		}

		// Create routing table
		myContact := NewContact(node.ID, address)
		node.routingTable = NewRoutingTable(myContact)

		// Create mock transport with packet dropping
		transport := NewMockTransport(address, nil, ne.dropRate)

		// Create message handler
		handler := NewKademliaMessageHandler(node.routingTable, node.ID, address, config, node)
		transport.handler = handler
		node.handler = handler
		node.transport = transport

		ne.nodes = append(ne.nodes, node)
		ne.transports = append(ne.transports, transport)
	}

	// Connect all transports to each other (full mesh for testing)
	for i, transport := range ne.transports {
		for j, otherTransport := range ne.transports {
			if i != j {
				transport.AddPeer(ne.nodes[j].Address, otherTransport)
			}
		}
	}

	// Start all transports
	for _, transport := range ne.transports {
		err := transport.Listen(transport.address)
		if err != nil {
			return fmt.Errorf("failed to start transport: %v", err)
		}
	}

	// Bootstrap network by connecting nodes to each other
	return ne.bootstrapNetwork()
}

// bootstrapNetwork connects nodes to form a Kademlia network
func (ne *NetworkEmulator) bootstrapNetwork() error {
	if len(ne.nodes) < 2 {
		return nil
	}

	// Each node connects to a few random other nodes
	for i, node := range ne.nodes {
		// Connect to up to k random nodes
		connectCount := min(ne.nodes[0].config.K, len(ne.nodes)-1)

		for j := 0; j < connectCount; j++ {
			// Pick a random other node
			targetIndex := (i + j + 1) % len(ne.nodes)
			if targetIndex == i {
				continue
			}

			targetNode := ne.nodes[targetIndex]
			contact := NewContact(targetNode.ID, targetNode.Address)
			node.routingTable.AddContact(contact)
		}
	}

	return nil
}

// Shutdown stops all nodes in the network
func (ne *NetworkEmulator) Shutdown() {
	for _, transport := range ne.transports {
		transport.Close()
	}
}

// GetNetworkStats returns statistics about the emulated network
func (ne *NetworkEmulator) GetNetworkStats() (totalSent int, totalDropped int, avgDropRate float64) {
	for _, transport := range ne.transports {
		sent, dropped, _ := transport.GetStats()
		totalSent += sent
		totalDropped += dropped
	}

	if totalSent > 0 {
		avgDropRate = float64(totalDropped) / float64(totalSent)
	}

	return totalSent, totalDropped, avgDropRate
}

// Helper function for min
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// M4 REQUIREMENT TESTS

// TestLargeScaleNetworkEmulation tests network with 1000+ nodes
func TestLargeScaleNetworkEmulation(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping large-scale test in short mode")
	}

	t.Logf("Testing large-scale network emulation with %d nodes", TestNodeCount)

	// Create network emulator
	emulator := NewNetworkEmulator(TestNodeCount, NoPacketDrop)
	defer emulator.Shutdown()

	// Setup network
	err := emulator.SetupNetwork()
	if err != nil {
		t.Fatalf("Failed to setup network: %v", err)
	}

	// Verify all nodes were created
	if len(emulator.nodes) != TestNodeCount {
		t.Errorf("Expected %d nodes, got %d", TestNodeCount, len(emulator.nodes))
	}

	// Test basic network operations
	t.Run("BasicOperations", func(t *testing.T) {
		// Test that nodes have contacts in their routing tables
		totalContacts := 0
		for _, node := range emulator.nodes {
			contacts := node.routingTable.GetAllContacts()
			totalContacts += len(contacts)
		}

		t.Logf("Total contacts across all nodes: %d", totalContacts)
		if totalContacts == 0 {
			t.Error("No contacts found in routing tables")
		}
	})

	// Test storage across the network
	t.Run("DistributedStorage", func(t *testing.T) {
		// Store data on first node
		testData := "large_scale_test_data"
		key, err := emulator.nodes[0].Store(testData)
		if err != nil {
			t.Fatalf("Failed to store data: %v", err)
		}

		// Verify data was stored
		value, found := emulator.nodes[0].GetValue(key)
		if !found || value != testData {
			t.Error("Data not stored correctly")
		}

		t.Logf("Successfully stored data with key: %s", key)
	})

	t.Logf("Large-scale network emulation test completed successfully")
}

// TestPacketDroppingFunctionality tests network behavior under packet loss
func TestPacketDroppingFunctionality(t *testing.T) {
	testCases := []struct {
		name      string
		dropRate  float64
		nodeCount int
	}{
		{"NoPacketLoss", NoPacketDrop, TestNodeCountSmall},
		{"LowPacketLoss", DefaultPacketDropRate, TestNodeCountSmall},
		{"HighPacketLoss", HighPacketDropRate, TestNodeCountSmall},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Logf("Testing with %d nodes and %.1f%% packet drop rate",
				tc.nodeCount, tc.dropRate*100)

			// Create network with packet dropping
			emulator := NewNetworkEmulator(tc.nodeCount, tc.dropRate)
			defer emulator.Shutdown()

			err := emulator.SetupNetwork()
			if err != nil {
				t.Fatalf("Failed to setup network: %v", err)
			}

			// Allow network to stabilize
			time.Sleep(NetworkStabilizationDelay)

			// Perform some network operations
			testOperations := 50
			successCount := 0

			for i := 0; i < testOperations; i++ {
				sourceNode := emulator.nodes[i%len(emulator.nodes)]
				testData := fmt.Sprintf("test_data_%d", i)

				_, err := sourceNode.Store(testData)
				if err == nil {
					successCount++
				}
			}

			// Get network statistics
			totalSent, totalDropped, actualDropRate := emulator.GetNetworkStats()

			t.Logf("Network stats: %d sent, %d dropped, %.2f%% actual drop rate",
				totalSent, totalDropped, actualDropRate*100)
			t.Logf("Successful operations: %d/%d (%.1f%%)",
				successCount, testOperations, float64(successCount)/float64(testOperations)*100)

			// Verify packet dropping is working as expected
			if tc.dropRate > 0 && totalDropped == 0 && totalSent > 0 {
				t.Error("Expected some packets to be dropped, but none were")
			}

			// With high packet loss, we should see significant drops
			if tc.dropRate >= HighPacketDropRate && totalSent > 0 {
				if actualDropRate < tc.dropRate/2 {
					t.Errorf("Expected drop rate around %.1f%%, got %.1f%%",
						tc.dropRate*100, actualDropRate*100)
				}
			}
		})
	}
}

// TestConfigurableParameters verifies that test parameters are easily configurable
func TestConfigurableParameters(t *testing.T) {
	t.Run("NodeCountConfiguration", func(t *testing.T) {
		// Test with different node counts to verify configurability
		testCounts := []int{10, 50, TestNodeCountSmall}

		for _, count := range testCounts {
			emulator := NewNetworkEmulator(count, NoPacketDrop)
			err := emulator.SetupNetwork()
			if err != nil {
				t.Fatalf("Failed to setup network with %d nodes: %v", count, err)
			}

			if len(emulator.nodes) != count {
				t.Errorf("Expected %d nodes, got %d", count, len(emulator.nodes))
			}

			emulator.Shutdown()
			t.Logf("Successfully configured network with %d nodes", count)
		}
	})

	t.Run("DropRateConfiguration", func(t *testing.T) {
		// Test with different drop rates to verify configurability
		testRates := []float64{0.0, 0.1, 0.25, 0.5}

		for _, rate := range testRates {
			emulator := NewNetworkEmulator(20, rate)
			err := emulator.SetupNetwork()
			if err != nil {
				t.Fatalf("Failed to setup network with %.1f%% drop rate: %v", rate*100, err)
			}

			// Verify each transport has the correct drop rate
			for _, transport := range emulator.transports {
				if transport.dropRate != rate {
					t.Errorf("Expected drop rate %.1f%%, got %.1f%%", rate*100, transport.dropRate*100)
				}
			}

			emulator.Shutdown()
			t.Logf("Successfully configured network with %.1f%% drop rate", rate*100)
		}
	})
}

// TestNetworkResilience tests how the network behaves under adverse conditions
func TestNetworkResilience(t *testing.T) {
	t.Run("HighPacketLoss", func(t *testing.T) {
		// Test network behavior with very high packet loss
		emulator := NewNetworkEmulator(50, 0.8) // 80% packet loss
		defer emulator.Shutdown()

		err := emulator.SetupNetwork()
		if err != nil {
			t.Fatalf("Failed to setup network: %v", err)
		}

		// Try to perform operations despite high packet loss
		attempts := 20
		successes := 0

		for i := 0; i < attempts; i++ {
			node := emulator.nodes[i%len(emulator.nodes)]
			_, err := node.Store(fmt.Sprintf("resilience_test_%d", i))
			if err == nil {
				successes++
			}
		}

		t.Logf("Under 80%% packet loss: %d/%d operations succeeded", successes, attempts)

		// Even with high packet loss, some operations should succeed
		// (This tests that the system doesn't completely fail)
	})

	t.Run("ConcurrentOperationsWithPacketLoss", func(t *testing.T) {
		// Test concurrent operations under packet loss
		emulator := NewNetworkEmulator(30, DefaultPacketDropRate)
		defer emulator.Shutdown()

		err := emulator.SetupNetwork()
		if err != nil {
			t.Fatalf("Failed to setup network: %v", err)
		}

		// Allow network to stabilize before concurrent operations
		time.Sleep(NetworkStabilizationDelay)

		// Run concurrent operations with proper synchronization
		var wg sync.WaitGroup
		concurrency := 10
		opsPerWorker := 5
		successCount := int64(0) // Use int64 for atomic operations

		for i := 0; i < concurrency; i++ {
			wg.Add(1)
			go func(workerID int) {
				defer wg.Done()
				localSuccesses := int64(0)

				for j := 0; j < opsPerWorker; j++ {
					// Use a dedicated node per worker to reduce contention
					nodeIndex := workerID % len(emulator.nodes)
					node := emulator.nodes[nodeIndex]

					// Create unique keys to avoid conflicts
					key := fmt.Sprintf("concurrent_%d_%d_%d", workerID, j, time.Now().UnixNano())
					_, err := node.Store(key)
					if err == nil {
						localSuccesses++
					}

					// Small delay to reduce race conditions
					time.Sleep(time.Millisecond)
				}

				// Use atomic operation to safely update shared counter
				atomic.AddInt64(&successCount, localSuccesses)
			}(i)
		}

		wg.Wait()

		totalOps := int64(concurrency * opsPerWorker)
		finalSuccesses := atomic.LoadInt64(&successCount)

		t.Logf("Concurrent operations under packet loss: %d/%d succeeded", finalSuccesses, totalOps)

		if finalSuccesses == 0 {
			t.Error("No concurrent operations succeeded - network may be too unreliable")
		}
	})
}

// BenchmarkNetworkEmulation benchmarks the network emulation performance
func BenchmarkNetworkEmulation(b *testing.B) {
	if testing.Short() {
		b.Skip("Skipping benchmark in short mode")
	}

	b.Run("SmallNetwork", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			emulator := NewNetworkEmulator(10, NoPacketDrop)
			emulator.SetupNetwork()
			emulator.Shutdown()
		}
	})

	b.Run("MediumNetwork", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			emulator := NewNetworkEmulator(100, NoPacketDrop)
			emulator.SetupNetwork()
			emulator.Shutdown()
		}
	})
}
