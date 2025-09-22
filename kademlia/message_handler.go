package kademlia

import (
	"encoding/json"
	"fmt"
	"time"
)

type KademliaMessageHandler struct {
	routingTable RoutingTableManager
	nodeID       *KademliaID
	nodeAddress  string
	config       *Config
	node         *Node // Direct reference to node for storage operations
}

func NewKademliaMessageHandler(routingTable RoutingTableManager, nodeID *KademliaID, nodeAddress string, config *Config, node *Node) *KademliaMessageHandler {
	return &KademliaMessageHandler{
		routingTable: routingTable,
		nodeID:       nodeID,
		nodeAddress:  nodeAddress,
		config:       config,
		node:         node,
	}
}

func (h *KademliaMessageHandler) HandleMessage(msg Message, sender string) (Message, error) {
	// Always update routing table with sender info
	fmt.Printf("DEBUG: Adding sender %s to routing table\n", msg.Sender.ID.String())
	h.routingTable.AddContact(msg.Sender)

	switch msg.Type {
	case PING:
		return h.handlePing(msg)
	case PONG:
		return h.handlePong(msg)
	case FIND_NODE:
		return h.handleFindNode(msg)
	case FIND_NODE_RESPONSE:
		return h.handleFindNodeResponse(msg)
	case STORE:
		return h.handleStore(msg)
	case FIND_VALUE:
		return h.handleFindValue(msg)
	default:
		return Message{}, fmt.Errorf("unknown message type: %d", msg.Type)
	}
}

func (h *KademliaMessageHandler) handlePing(msg Message) (Message, error) {
	fmt.Printf("DEBUG: Processing PING from %s\n", msg.Sender.ID.String())

	// Create our contact info for the response
	myContact := NewContact(h.nodeID, h.nodeAddress)

	response := Message{
		Type:      PONG,
		MessageID: msg.MessageID,
		Sender:    myContact,
		Timestamp: time.Now().Unix(),
		Data: PongData{
			Message: "pong",
		},
	}

	fmt.Printf("DEBUG: Created PONG response\n")
	return response, nil
}

func (h *KademliaMessageHandler) handlePong(msg Message) (Message, error) {
	fmt.Printf("SUCCESS: Received PONG from %s (MessageID: %s)\n", msg.Sender.ID.String(), msg.MessageID)
	return Message{}, nil
}

func (h *KademliaMessageHandler) handleFindNode(msg Message) (Message, error) {
	fmt.Printf("DEBUG: Processing FIND_NODE request\n")

	// Handle the JSON unmarshaling properly
	dataBytes, err := json.Marshal(msg.Data)
	if err != nil {
		return Message{}, fmt.Errorf("failed to marshal message data: %v", err)
	}

	fmt.Printf("DEBUG: Raw data bytes: %s\n", string(dataBytes))

	// Create a new KademliaID instance for unmarshaling
	var findData struct {
		TargetID KademliaID `json:"target_id"`
	}

	err = json.Unmarshal(dataBytes, &findData)
	if err != nil {
		return Message{}, fmt.Errorf("failed to unmarshal FIND_NODE data: %v", err)
	}

	targetID := &findData.TargetID // Now take address

	fmt.Printf("DEBUG: Target ID: %s\n", targetID.String())

	// Rest of your function stays the same...
	closestContacts := LookupNode(h.routingTable, targetID)

	fmt.Printf("DEBUG: Found %d closest contacts for target %s\n", len(closestContacts), targetID.String())

	myContact := NewContact(h.nodeID, h.nodeAddress)

	response := Message{
		Type:      FIND_NODE_RESPONSE, //I changed this because it from find_node PERMA LOOPS otherwise
		MessageID: msg.MessageID,
		Sender:    myContact,
		Timestamp: time.Now().Unix(),
		Data: FindNodeResponse{
			Contacts: closestContacts,
		},
	}

	return response, nil
}

func (h *KademliaMessageHandler) handleFindNodeResponse(msg Message) (Message, error) {
	fmt.Printf("SUCCESS: Received FIND_NODE_RESPONSE with contacts\n")

	// Extract contacts from response and add them to routing table
	dataBytes, err := json.Marshal(msg.Data)
	if err != nil {
		return Message{}, fmt.Errorf("failed to marshal response data: %v", err)
	}

	var responseData FindNodeResponse
	err = json.Unmarshal(dataBytes, &responseData)
	if err != nil {
		return Message{}, fmt.Errorf("failed to unmarshal FIND_NODE response: %v", err)
	}

	// Add all returned contacts to routing table
	for _, contact := range responseData.Contacts {
		h.routingTable.AddContact(contact)
	}

	fmt.Printf("DEBUG: Added %d contacts from FIND_NODE response\n", len(responseData.Contacts))

	// No response needed for a response message
	return Message{}, nil
}

func (h *KademliaMessageHandler) handleStore(msg Message) (Message, error) {
	fmt.Printf("DEBUG: Processing STORE request\n")

	// Extract store data
	dataBytes, err := json.Marshal(msg.Data)
	if err != nil {
		return Message{}, fmt.Errorf("failed to marshal store data: %v", err)
	}

	var storeData StoreData
	err = json.Unmarshal(dataBytes, &storeData)
	if err != nil {
		return Message{}, fmt.Errorf("failed to unmarshal STORE data: %v", err)
	}

	fmt.Printf("DEBUG: Storing key=%s, value=%s\n", storeData.Key, storeData.Value)

	// Store the value locally using the node
	h.node.StoreValue(storeData.Key, storeData.Value)

	// STORE messages typically don't require a response in Kademlia
	return Message{}, nil
}

func (h *KademliaMessageHandler) handleFindValue(msg Message) (Message, error) {
	fmt.Printf("DEBUG: Processing FIND_VALUE request\n")

	// Extract find value data
	dataBytes, err := json.Marshal(msg.Data)
	if err != nil {
		return Message{}, fmt.Errorf("failed to marshal find value data: %v", err)
	}

	var findValueData FindValueData
	err = json.Unmarshal(dataBytes, &findValueData)
	if err != nil {
		return Message{}, fmt.Errorf("failed to unmarshal FIND_VALUE data: %v", err)
	}

	key := findValueData.Key
	fmt.Printf("DEBUG: Looking for key: %s\n", key)

	// Check if we have the value locally
	value, found := h.node.GetValue(key)

	myContact := NewContact(h.nodeID, h.nodeAddress)

	if found {
		// Return the value
		fmt.Printf("DEBUG: Found value locally for key %s\n", key)
		response := Message{
			Type:      FIND_VALUE,
			MessageID: msg.MessageID,
			Sender:    myContact,
			Timestamp: time.Now().Unix(),
			Data: FindValueResponse{
				Found: true,
				Value: value,
			},
		}
		return response, nil
	} else {
		// Return closest contacts (like FIND_NODE)
		fmt.Printf("DEBUG: Value not found locally, returning closest contacts\n")
		keyID := NewKademliaID(key)
		closestContacts := h.routingTable.FindClosestContacts(keyID, h.config.Alpha)

		response := Message{
			Type:      FIND_VALUE,
			MessageID: msg.MessageID,
			Sender:    myContact,
			Timestamp: time.Now().Unix(),
			Data: FindValueResponse{
				Found:    false,
				Contacts: closestContacts,
			},
		}
		return response, nil
	}
}
