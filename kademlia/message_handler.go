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
	dataStore    *DataStore
}

func NewKademliaMessageHandler(routingTable RoutingTableManager, nodeID *KademliaID, nodeAddress string, config *Config, dataStore *DataStore) *KademliaMessageHandler {
	return &KademliaMessageHandler{
		routingTable: routingTable,
		nodeID:       nodeID,
		nodeAddress:  nodeAddress,
		config:       config,
		dataStore:    dataStore,
	}
}

func (h *KademliaMessageHandler) HandleMessage(msg Message, sender string) (Message, error) {
	// Always update routing table with sender info
	//fmt.Printf("DEBUG: HandleMessage called with type %s from %s\n", msg.Type.String(), sender)
	//fmt.Printf("DEBUG: Adding sender %s to routing table\n", msg.Sender.ID.String())
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
	case FIND_VALUE:
		return h.handleFindValue(msg)
	case FIND_VALUE_RESPONSE:
		return h.handleFindValueResponse(msg)
	case STORE:
		return h.handleStore(msg)
	case STORE_RESPONSE:
		return h.handleStoreResponse(msg)
	default:
		return Message{}, fmt.Errorf("unknown message type: %d", msg.Type)
	}
}

func (h *KademliaMessageHandler) handlePing(msg Message) (Message, error) {
	//fmt.Printf("DEBUG: Processing PING from %s\n", msg.Sender.ID.String())

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

	//fmt.Printf("DEBUG: Created PONG response\n")
	return response, nil
}

func (h *KademliaMessageHandler) handlePong(msg Message) (Message, error) {
	fmt.Printf("SUCCESS: Received PONG from %s (MessageID: %s)\n", msg.Sender.ID.String(), msg.MessageID)
	return Message{}, nil
}

func (h *KademliaMessageHandler) handleFindNode(msg Message) (Message, error) {
	//fmt.Printf("DEBUG: Processing FIND_NODE request\n")

	// 1) Re-marshal the generic data and decode it into the typed struct
	dataBytes, err := json.Marshal(msg.Data)
	if err != nil {
		return Message{}, fmt.Errorf("failed to marshal message data: %v", err)
	}

	var req FindNodeData
	if err := json.Unmarshal(dataBytes, &req); err != nil {
		return Message{}, fmt.Errorf("failed to unmarshal FIND_NODE data: %v", err)
	}
	if req.TargetID == nil {
		return Message{}, fmt.Errorf("missing target_id in FIND_NODE request")
	}

	targetID := req.TargetID
	//fmt.Printf("DEBUG: Target ID (FIND_NODE): %s\n", targetID.String())

	// 2) Find closest contacts (consider returning K, not Alpha)
	closestContacts := LookupNode(h.routingTable, targetID)

	myContact := NewContact(h.nodeID, h.nodeAddress)

	// 3) Build response — IMPORTANT: echo back the SAME MessageID
	resp := Message{
		Type:      FIND_NODE_RESPONSE,
		MessageID: msg.MessageID,
		Sender:    myContact,
		Data:      FindNodeResponse{Contacts: closestContacts},
		Timestamp: time.Now().Unix(),
	}

	return resp, nil
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

	//fmt.Printf("DEBUG: Added %d contacts from FIND_NODE response\n", len(responseData.Contacts))

	// No response needed for a response message
	return Message{}, nil
}

func (h *KademliaMessageHandler) handleFindValue(msg Message) (Message, error) {
	//fmt.Printf("DEBUG: Processing FIND_VALUE request\n")

	// Parse key
	key := ""
	if d, ok := msg.Data.(FindValueData); ok {
		key = d.Key
	} else if m, ok := msg.Data.(map[string]interface{}); ok {
		if s, ok := m["key"].(string); ok {
			key = s
		}
	}
	if key == "" {
		return Message{}, fmt.Errorf("FIND_VALUE: missing key")
	}

	my := NewContact(h.nodeID, h.nodeAddress)

	// If we have it locally, return the value
	if v, ok := h.dataStore.Get(key); ok {
		return Message{
			Type:      FIND_VALUE_RESPONSE,
			MessageID: msg.MessageID,
			Sender:    my,
			Timestamp: time.Now().Unix(),
			Data: FindValueResponse{
				Found:    true,
				Value:    v,
				Contacts: nil,
			},
		}, nil
	}

	// Otherwise, return k closest contacts towards the key hash
	target := HashKey(key)
	contacts := h.routingTable.FindClosestContacts(target, h.config.K)

	return Message{
		Type:      FIND_VALUE_RESPONSE,
		MessageID: msg.MessageID,
		Sender:    my,
		Timestamp: time.Now().Unix(),
		Data: FindValueResponse{
			Found:    false,
			Contacts: contacts,
		},
	}, nil
}

func (h *KademliaMessageHandler) handleFindValueResponse(msg Message) (Message, error) {
	fmt.Printf("SUCCESS: Received FIND_VALUE_RESPONSE\n")

	// Extract response data
	dataBytes, err := json.Marshal(msg.Data)
	if err != nil {
		return Message{}, fmt.Errorf("failed to marshal response data: %v", err)
	}

	var responseData FindValueResponse
	err = json.Unmarshal(dataBytes, &responseData)
	if err != nil {
		return Message{}, fmt.Errorf("failed to unmarshal FIND_VALUE response: %v", err)
	}

	if responseData.Found {
		//fmt.Printf("DEBUG: Response contains value\n")
	} else {
		//fmt.Printf("DEBUG: Response contains %d contacts\n", len(responseData.Contacts))
		// Add contacts to routing table
		for _, contact := range responseData.Contacts {
			h.routingTable.AddContact(contact)
		}
	}

	// No response needed for a response message
	return Message{}, nil
}

func (h *KademliaMessageHandler) handleStore(msg Message) (Message, error) {
	fmt.Printf("DEBUG: Processing STORE request\n")

	// Parse the STORE data
	dataBytes, err := json.Marshal(msg.Data)
	if err != nil {
		return Message{}, fmt.Errorf("failed to marshal message data: %v", err)
	}

	var storeData StoreData
	err = json.Unmarshal(dataBytes, &storeData)
	if err != nil {
		return Message{}, fmt.Errorf("failed to unmarshal STORE data: %v", err)
	}

	key := storeData.Key
	value := storeData.Value

	//fmt.Printf("DEBUG: Storing key=%s, value_length=%d\n", key, len(value))

	// Store the key-value pair
	h.dataStore.Store(key, value)

	// Create success response
	myContact := NewContact(h.nodeID, h.nodeAddress)
	response := Message{
		Type:      STORE_RESPONSE,
		MessageID: msg.MessageID,
		Sender:    myContact,
		Timestamp: time.Now().Unix(),
		Data: StoreResponse{
			Status: "success",
		},
	}

	fmt.Printf("SUCCESS: Stored key-value pair, sending response\n")
	return response, nil
}

func (h *KademliaMessageHandler) handleStoreResponse(msg Message) (Message, error) {
	fmt.Printf("SUCCESS: Received STORE_RESPONSE\n")

	// Extract response data to check status
	dataBytes, err := json.Marshal(msg.Data)
	if err != nil {
		return Message{}, fmt.Errorf("failed to marshal response data: %v", err)
	}

	var responseData StoreResponse
	err = json.Unmarshal(dataBytes, &responseData)
	if err != nil {
		return Message{}, fmt.Errorf("failed to unmarshal STORE response: %v", err)
	}

	//fmt.Printf("DEBUG: Store operation status: %s\n", responseData.Status)

	// No response needed for a response message
	return Message{}, nil
}
