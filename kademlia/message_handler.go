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
}

func NewKademliaMessageHandler(routingTable RoutingTableManager, nodeID *KademliaID, nodeAddress string, config *Config) *KademliaMessageHandler {
	return &KademliaMessageHandler{
		routingTable: routingTable,
		nodeID:       nodeID,
		nodeAddress:  nodeAddress,
		config:       config,
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