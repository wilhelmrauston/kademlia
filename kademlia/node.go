package kademlia

import (
	"crypto/sha1"
	"fmt"
	"sync"
	"time"
)

type Node struct {
	ID           *KademliaID
	Address      string
	transport    NetworkTransport
	routingTable RoutingTableManager
	handler      MessageHandler
	config       *Config
	dataStore    map[string]string // Key-value store for data objects
	storeMutex   sync.RWMutex      // Protect concurrent access to dataStore
}

func NewNode(address string, nodeID string, config *Config) *Node {
	var id *KademliaID
	if nodeID == "" {
		id = NewRandomKademliaID()
	} else {
		id = NewKademliaID(nodeID)
	}

	// Create our contact
	myContact := NewContact(id, address)

	// Create routing table
	rt := NewRoutingTable(myContact)

	// Create the node first
	node := &Node{
		ID:           id,
		Address:      address,
		routingTable: rt,
		config:       config,
		dataStore:    make(map[string]string),
		storeMutex:   sync.RWMutex{},
	}

	// Create message handler with node as datastore
	handler := NewKademliaMessageHandler(rt, id, address, config, node)

	// Create transport with handler
	transport := NewUDPTransport(config, handler)

	// Set transport in node
	node.transport = transport
	node.handler = handler

	return node
}

func (n *Node) Start() error {
	fmt.Printf("Starting node %s on %s\n", n.ID.String(), n.Address)

	// Start transport layer
	go func() {
		err := n.transport.Listen(n.Address)
		if err != nil {
			fmt.Printf("ERROR: Transport failed: %v\n", err)
		}
	}()

	// Give transport time to start
	time.Sleep(1 * time.Second)
	fmt.Printf("Node started successfully\n")
	return nil
}

func (n *Node) Stop() error {
	fmt.Printf("Stopping node %s\n", n.ID.String())
	return n.transport.Close()
}

func (n *Node) SendPing(targetAddr string) error {

	myContact := NewContact(n.ID, n.Address)

	msg := Message{
		Type:      PING,
		MessageID: generateMessageID(),
		Sender:    myContact,
		Timestamp: time.Now().Unix(),
		Data: PingData{
			Message: "ping",
		},
	}

	return n.transport.Send(msg, targetAddr)
}

func (n *Node) SendFindNode(targetAddr string, targetID *KademliaID) error {
	myContact := NewContact(n.ID, n.Address)

	msg := Message{
		Type:      FIND_NODE,
		MessageID: generateMessageID(),
		Sender:    myContact,
		Timestamp: time.Now().Unix(),
		Data: FindNodeData{
			TargetID: targetID,
		},
	}

	return n.transport.Send(msg, targetAddr)
}

func (n *Node) GetRoutingTableInfo() string {
	contacts := n.routingTable.GetAllContacts()
	return fmt.Sprintf("Node %s has %d contacts in routing table", n.ID.String(), len(contacts))
}

// StoreValue stores a key-value pair locally
func (n *Node) StoreValue(key string, value string) {
	n.storeMutex.Lock()
	defer n.storeMutex.Unlock()
	n.dataStore[key] = value
	fmt.Printf("DEBUG: Stored value for key %s locally\n", key)
}

// GetValue retrieves a value by key from local storage
func (n *Node) GetValue(key string) (string, bool) {
	n.storeMutex.RLock()
	defer n.storeMutex.RUnlock()
	value, exists := n.dataStore[key]
	return value, exists
}

// HashData generates a SHA-1 hash of the input data
func HashData(data string) string {
	hash := sha1.Sum([]byte(data))
	return fmt.Sprintf("%x", hash)
}

// SendStore sends a STORE message to store data on the network
func (n *Node) SendStore(data string) (string, error) {
	// Generate hash of the data
	key := HashData(data)
	keyID := NewKademliaID(key)

	// Find k closest nodes to the key
	closestContacts := n.routingTable.FindClosestContacts(keyID, n.config.K)

	if len(closestContacts) == 0 {
		// If no contacts, store locally
		n.StoreValue(key, data)
		fmt.Printf("No contacts found, stored locally. Key: %s\n", key)
		return key, nil
	}

	// Send STORE messages to closest nodes
	myContact := NewContact(n.ID, n.Address)

	for _, contact := range closestContacts {
		msg := Message{
			Type:      STORE,
			MessageID: generateMessageID(),
			Sender:    myContact,
			Timestamp: time.Now().Unix(),
			Data: StoreData{
				Key:   key,
				Value: data,
			},
		}

		err := n.transport.Send(msg, contact.Address)
		if err != nil {
			fmt.Printf("Failed to send STORE to %s: %v\n", contact.Address, err)
		} else {
			fmt.Printf("Sent STORE message to %s for key %s\n", contact.Address, key)
		}
	}

	// Also store locally (as per Kademlia protocol)
	n.StoreValue(key, data)

	return key, nil
}

// SendFindValue sends a FIND_VALUE message to retrieve data from the network
func (n *Node) SendFindValue(key string) (string, bool, error) {
	// First check if we have it locally
	if value, exists := n.GetValue(key); exists {
		fmt.Printf("Found value locally for key %s\n", key)
		return value, true, nil
	}

	keyID := NewKademliaID(key)
	closestContacts := n.routingTable.FindClosestContacts(keyID, n.config.Alpha)

	if len(closestContacts) == 0 {
		return "", false, fmt.Errorf("no contacts available for lookup")
	}

	// Send FIND_VALUE to closest contacts
	myContact := NewContact(n.ID, n.Address)

	for _, contact := range closestContacts {
		msg := Message{
			Type:      FIND_VALUE,
			MessageID: generateMessageID(),
			Sender:    myContact,
			Timestamp: time.Now().Unix(),
			Data: FindValueData{
				Key: key,
			},
		}

		err := n.transport.Send(msg, contact.Address)
		if err != nil {
			fmt.Printf("Failed to send FIND_VALUE to %s: %v\n", contact.Address, err)
		} else {
			fmt.Printf("Sent FIND_VALUE message to %s for key %s\n", contact.Address, key)
		}
	}

	// Note: In a real implementation, this would wait for responses
	// For now, we'll return that the value wasn't found immediately
	// The actual response handling would be done in the message handler
	return "", false, nil
}

// Store is a convenience method that calls SendStore
func (n *Node) Store(data string) (string, error) {
	return n.SendStore(data)
}

// LookupValue is a convenience method that calls SendFindValue
func (n *Node) LookupValue(key string) (string, string, error) {
	value, found, err := n.SendFindValue(key)
	if found {
		return value, n.Address, err
	}
	return "", "", err
}

// AddContact adds a contact to the routing table
func (n *Node) AddContact(contact Contact) {
	n.routingTable.AddContact(contact)
}

// Helper function to generate message IDs
func generateMessageID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}
