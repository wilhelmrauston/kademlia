package kademlia

import (
	"fmt"
	"time"
)

type Node struct {
	ID           *KademliaID
	Address      string
	transport    NetworkTransport
	routingTable RoutingTableManager
	handler      MessageHandler
	config       *Config
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

	// Create message handler
	handler := NewKademliaMessageHandler(rt, id, address, config)

	// Create transport with handler
	transport := NewUDPTransport(config, handler)

	return &Node{
		ID:           id,
		Address:      address,
		transport:    transport,
		routingTable: rt,
		handler:      handler,
		config:       config,
	}
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

// Helper function to generate message IDs
func generateMessageID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}