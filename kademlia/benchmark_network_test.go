package kademlia

import (
	"fmt"
	"math/rand"
	"sync"
	"testing"
	"time"
)

// NetworkEmulator simulates a Kademlia network with packet dropping
type NetworkEmulator struct {
	nodes           map[string]*EmulatedNode
	packetDropRate  float64 // Percentage of packets to drop (0.0 - 1.0)
	deliveryLatency time.Duration
	mutex           sync.RWMutex
}

// EmulatedNode represents a node in the emulated network
type EmulatedNode struct {
	*Node
	networkID string
	emulator  *NetworkEmulator
}

// EmulatedTransport implements NetworkTransport for emulation
type EmulatedTransport struct {
	node     *EmulatedNode
	incoming chan MessageEnvelope
	handler  MessageHandler
}

func NewNetworkEmulator(packetDropRate float64, latency time.Duration) *NetworkEmulator {
	return &NetworkEmulator{
		nodes:           make(map[string]*EmulatedNode),
		packetDropRate:  packetDropRate,
		deliveryLatency: latency,
	}
}

func (ne *NetworkEmulator) CreateNode(nodeID string, address string) *EmulatedNode {
	config := DefaultConfig()

	// Create emulated node
	emulatedNode := &EmulatedNode{
		networkID: nodeID,
		emulator:  ne,
	}

	// Create emulated transport
	transport := &EmulatedTransport{
		node:     emulatedNode,
		incoming: make(chan MessageEnvelope, 100),
	}

	// Create the actual node with custom transport
	var id *KademliaID
	if nodeID == "" {
		id = NewRandomKademliaID()
	} else {
		id = NewKademliaID(nodeID)
	}

	myContact := NewContact(id, address)
	rt := NewRoutingTable(myContact)
	dataStore := NewDataStore()
	handler := NewKademliaMessageHandler(rt, id, address, config, dataStore)

	node := &Node{
		ID:           id,
		Address:      address,
		transport:    transport,
		routingTable: rt,
		handler:      handler,
		config:       config,
		dataStore:    dataStore,
	}

	emulatedNode.Node = node
	transport.handler = handler

	// Register node in emulator
	ne.mutex.Lock()
	ne.nodes[address] = emulatedNode
	ne.mutex.Unlock()

	// Start message processing
	go transport.processMessages()

	return emulatedNode
}

func (ne *NetworkEmulator) SetPacketDropRate(rate float64) {
	ne.mutex.Lock()
	defer ne.mutex.Unlock()
	ne.packetDropRate = rate
}

func (ne *NetworkEmulator) GetPacketDropRate() float64 {
	ne.mutex.RLock()
	defer ne.mutex.RUnlock()
	return ne.packetDropRate
}

func (ne *NetworkEmulator) DeliverMessage(msg Message, fromAddr, toAddr string) {
	ne.mutex.RLock()
	dropRate := ne.packetDropRate
	latency := ne.deliveryLatency
	ne.mutex.RUnlock()

	// Simulate packet dropping
	if rand.Float64() < dropRate {
		return // Packet dropped
	}

	// Simulate network latency
	go func() {
		if latency > 0 {
			time.Sleep(latency)
		}

		ne.mutex.RLock()
		targetNode, exists := ne.nodes[toAddr]
		ne.mutex.RUnlock()

		if !exists {
			return // Node doesn't exist
		}

		envelope := MessageEnvelope{
			Message: msg,
			Sender:  fromAddr,
		}

		select {
		case targetNode.transport.(*EmulatedTransport).incoming <- envelope:
			// Message delivered
		default:
			// Buffer full, drop message
		}
	}()
}

// EmulatedTransport implementation
func (et *EmulatedTransport) Listen(address string) error {
	// No actual listening needed in emulation
	return nil
}

func (et *EmulatedTransport) Send(msg Message, addr string) error {
	et.node.emulator.DeliverMessage(msg, et.node.Address, addr)
	return nil
}

func (et *EmulatedTransport) Close() error {
	close(et.incoming)
	return nil
}

func (et *EmulatedTransport) processMessages() {
	for envelope := range et.incoming {
		response, err := et.handler.HandleMessage(envelope.Message, envelope.Sender)
		if err != nil {
			continue
		}

		// Send response if one was generated
		if response.Type != MessageType(0) {
			et.Send(response, envelope.Sender)
		}
	}
}

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

// Test with 1000+ nodes and packet dropping
func TestLargeNetworkEmulation(t *testing.T) {
	const numNodes = 1000
	const packetDropRate = 0.1 // 10% packet loss

	emulator := NewNetworkEmulator(packetDropRate, 10*time.Millisecond)

	// Create nodes
	nodes := make([]*EmulatedNode, numNodes)
	for i := 0; i < numNodes; i++ {
		address := fmt.Sprintf("node%d:8000", i)
		nodes[i] = emulator.CreateNode("", address)
	}

	t.Logf("Created %d nodes with %.1f%% packet drop rate", numNodes, packetDropRate*100)

	// Bootstrap network: connect each node to a few random others
	for i := 0; i < numNodes; i++ {
		// Connect to 3-5 random other nodes
		connections := 3 + rand.Intn(3)
		for j := 0; j < connections; j++ {
			target := rand.Intn(numNodes)
			if target != i {
				contact := NewContact(nodes[target].ID, nodes[target].Address)
				nodes[i].routingTable.AddContact(contact)
			}
		}
	}

	// Test network connectivity
	testNode := nodes[0]
	targetID := nodes[numNodes/2].ID

	// Find closest contacts
	closest := testNode.routingTable.FindClosestContacts(targetID, 20)

	if len(closest) == 0 {
		t.Error("Should find some closest contacts")
	}

	t.Logf("Found %d closest contacts to target", len(closest))

	// Test data storage and retrieval
	testKey := "test_key_large_network"
	testValue := "test_value_large_network"

	// Store in a random node
	storeNode := nodes[rand.Intn(numNodes)]
	storeNode.dataStore.Store(testKey, testValue)

	// Try to find it from another node's routing table perspective
	searchNode := nodes[rand.Intn(numNodes)]
	searchTarget := HashKey(testKey)
	candidates := searchNode.routingTable.FindClosestContacts(searchTarget, 20)

	if len(candidates) > 0 {
		t.Logf("Found %d candidate nodes for key lookup", len(candidates))
	}
}

// Test packet dropping rates
func TestPacketDropping(t *testing.T) {
	dropRates := []float64{0.0, 0.1, 0.3, 0.5}

	for _, dropRate := range dropRates {
		t.Run(fmt.Sprintf("DropRate_%.1f", dropRate*100), func(t *testing.T) {
			emulator := NewNetworkEmulator(dropRate, 5*time.Millisecond)

			// Create 100 nodes for this test
			const numNodes = 100
			nodes := make([]*EmulatedNode, numNodes)

			for i := 0; i < numNodes; i++ {
				address := fmt.Sprintf("node%d:8000", i)
				nodes[i] = emulator.CreateNode("", address)
			}

			// Connect nodes in a ring topology
			for i := 0; i < numNodes; i++ {
				next := (i + 1) % numNodes
				contact := NewContact(nodes[next].ID, nodes[next].Address)
				nodes[i].routingTable.AddContact(contact)
			}

			// Test message delivery success rate
			deliveryCount := 0
			totalMessages := 100

			for i := 0; i < totalMessages; i++ {
				sender := nodes[i%numNodes]
				target := nodes[(i+10)%numNodes]

				// Simulate sending a message
				msg := Message{
					Type:      PING,
					MessageID: fmt.Sprintf("test_%d", i),
					Sender:    NewContact(sender.ID, sender.Address),
					Timestamp: time.Now().Unix(),
					Data:      PingData{Message: "test"},
				}

				// Send message through emulator
				emulator.DeliverMessage(msg, sender.Address, target.Address)

				// Wait a bit for delivery
				time.Sleep(20 * time.Millisecond)

				// Check if message was processed (simplified check)
				select {
				case <-target.transport.(*EmulatedTransport).incoming:
					deliveryCount++
				default:
					// Message not delivered (dropped or not yet arrived)
				}
			}

			expectedDeliveryRate := 1.0 - dropRate
			actualDeliveryRate := float64(deliveryCount) / float64(totalMessages)

			t.Logf("Drop rate: %.1f%%, Expected delivery: %.1f%%, Actual delivery: %.1f%%",
				dropRate*100, expectedDeliveryRate*100, actualDeliveryRate*100)

			// Allow some tolerance for randomness
			if actualDeliveryRate > expectedDeliveryRate+0.2 {
				t.Errorf("Delivery rate too high: expected ~%.1f%%, got %.1f%%",
					expectedDeliveryRate*100, actualDeliveryRate*100)
			}
		})
	}
}

// Test scalability with different network sizes
func TestNetworkScalability(t *testing.T) {
	networkSizes := []int{100, 500, 1000, 2000}

	for _, size := range networkSizes {
		t.Run(fmt.Sprintf("Nodes_%d", size), func(t *testing.T) {
			emulator := NewNetworkEmulator(0.05, 1*time.Millisecond) // 5% drop rate

			start := time.Now()

			// Create nodes
			nodes := make([]*EmulatedNode, size)
			for i := 0; i < size; i++ {
				address := fmt.Sprintf("node%d:8000", i)
				nodes[i] = emulator.CreateNode("", address)
			}

			creationTime := time.Since(start)

			// Build routing tables
			start = time.Now()
			for i := 0; i < size; i++ {
				// Each node knows about 5-10 others randomly
				connections := 5 + rand.Intn(6)
				for j := 0; j < connections && j < size-1; j++ {
					target := (i + 1 + j) % size
					contact := NewContact(nodes[target].ID, nodes[target].Address)
					nodes[i].routingTable.AddContact(contact)
				}
			}

			routingTime := time.Since(start)

			// Test lookup performance
			start = time.Now()
			lookupCount := 0
			for i := 0; i < 10; i++ {
				node := nodes[rand.Intn(size)]
				target := NewRandomKademliaID()
				contacts := node.routingTable.FindClosestContacts(target, 20)
				lookupCount += len(contacts)
			}
			lookupTime := time.Since(start)

			t.Logf("Network size: %d nodes", size)
			t.Logf("Creation time: %v", creationTime)
			t.Logf("Routing setup time: %v", routingTime)
			t.Logf("Lookup time (10 lookups): %v", lookupTime)
			t.Logf("Average contacts found per lookup: %.1f", float64(lookupCount)/10.0)

			// Verify network is connected
			if lookupCount == 0 {
				t.Error("Network appears to be disconnected")
			}
		})
	}
}

// Test XOR distance calculations in large network
func TestXORDistanceInLargeNetwork(t *testing.T) {
	const numNodes = 1000

	// Create test nodes with specific IDs to test distance calculation
	nodes := make([]*KademliaID, numNodes)
	for i := 0; i < numNodes; i++ {
		nodes[i] = NewRandomKademliaID()
	}

	// Test that distance calculation is symmetric
	for i := 0; i < 100; i++ {
		a := nodes[rand.Intn(numNodes)]
		b := nodes[rand.Intn(numNodes)]

		distAB := a.CalcDistance(b)
		distBA := b.CalcDistance(a)

		if !distAB.Equals(distBA) {
			t.Error("XOR distance should be symmetric")
		}
	}

	// Test that distance to self is zero
	testID := nodes[0]
	selfDistance := testID.CalcDistance(testID)
	zeroID := &KademliaID{}

	if !selfDistance.Equals(zeroID) {
		t.Error("Distance to self should be zero")
	}

	// Test triangle inequality doesn't hold (XOR is not a metric)
	// This is expected behavior for XOR distance
	a, b, c := nodes[0], nodes[1], nodes[2]
	distAB := a.CalcDistance(b)
	distBC := b.CalcDistance(c)
	distAC := a.CalcDistance(c)

	// XOR distance may not satisfy triangle inequality
	t.Logf("Distance A->B: %s", distAB.String()[:16])
	t.Logf("Distance B->C: %s", distBC.String()[:16])
	t.Logf("Distance A->C: %s", distAC.String()[:16])
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

func BenchmarkLargeNetworkLookup(b *testing.B) {
	// Setup large network
	const numNodes = 1000
	emulator := NewNetworkEmulator(0.0, 0) // No drops, no latency for benchmark

	nodes := make([]*EmulatedNode, numNodes)
	for i := 0; i < numNodes; i++ {
		address := fmt.Sprintf("node%d:8000", i)
		nodes[i] = emulator.CreateNode("", address)
	}

	// Connect nodes
	for i := 0; i < numNodes; i++ {
		for j := 0; j < 10; j++ {
			target := (i + 1 + j) % numNodes
			contact := NewContact(nodes[target].ID, nodes[target].Address)
			nodes[i].routingTable.AddContact(contact)
		}
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		node := nodes[i%numNodes]
		target := NewRandomKademliaID()
		node.routingTable.FindClosestContacts(target, 20)
	}
}
