package kademlia

import (
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sort"
	"sync"
	"time"
)

type DataStore struct {
    data map[string]string
    mutex sync.RWMutex
}

type Node struct {
	ID           *KademliaID
	Address      string
	transport    NetworkTransport
	routingTable RoutingTableManager
	handler      MessageHandler
	config       *Config
	dataStore    *DataStore
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

	// Create data store
	dataStore := NewDataStore()

	// Create message handler
	handler := NewKademliaMessageHandler(rt, id, address, config, dataStore)

	// Create transport with handler
	transport := NewUDPTransport(config, handler)

	return &Node{
		ID:           id,
		Address:      address,
		transport:    transport,
		routingTable: rt,
		handler:      handler,
		config:       config,
		dataStore:    dataStore,
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

// IterativeFindNode performs the core Kademlia lookup algorithm
func (n *Node) IterativeFindNode(targetID *KademliaID) []Contact {
    alpha := n.config.Alpha // 3
    k := n.config.K         // 20
    
    fmt.Printf("Starting iterative lookup for target %s\n", targetID.String()[:8])
    
    // Step 1: Initialize with k closest contacts from our routing table
    candidates := n.routingTable.FindClosestContacts(targetID, k)
    if len(candidates) == 0 {
        fmt.Printf("No initial candidates found in routing table\n")
        return []Contact{}
    }
    
    // Calculate distances for all candidates
    for i := range candidates {
        candidates[i].CalcDistance(targetID)
    }
    
    // Sort by distance to target
    sort.Slice(candidates, func(i, j int) bool {
        return candidates[i].distance.Less(candidates[j].distance)
    })
    
    fmt.Printf("Starting with %d candidates from routing table\n", len(candidates))
    
    // Track which nodes we've queried
    queried := make(map[string]bool)
    
    // Track the closest distance seen so far
    var closestDistance *KademliaID
    if len(candidates) > 0 {
        closestDistance = candidates[0].distance
    }
    
    for round := 1; round <= 10; round++ { // Max 10 rounds to prevent infinite loops
        fmt.Printf("--- Lookup round %d ---\n", round)
        
        // Step 2: Select α closest unqueried nodes
        var toQuery []Contact
        for _, contact := range candidates {
            contactKey := contact.ID.String()
            if !queried[contactKey] && len(toQuery) < alpha {
                toQuery = append(toQuery, contact)
                queried[contactKey] = true
            }
        }
        
        if len(toQuery) == 0 {
            fmt.Printf("No more unqueried nodes available\n")
            break
        }
        
        fmt.Printf("Querying %d nodes in parallel\n", len(toQuery))
        
        // Step 3: Query selected nodes in parallel
        responseChan := make(chan []Contact, len(toQuery))
        
        for _, contact := range toQuery {
            go func(c Contact) {
                contacts, err := n.SendFindNodeAndWait(c.Address, targetID, 5*time.Second)
                if err != nil {
                    fmt.Printf("Query to %s failed: %v\n", c.Address, err)
                    responseChan <- []Contact{}
                } else {
                    responseChan <- contacts
                }
            }(contact)
        }
        
        // Collect all responses
        var newContacts []Contact
        for i := 0; i < len(toQuery); i++ {
            select {
            case contacts := <-responseChan:
                newContacts = append(newContacts, contacts...)
            case <-time.After(6 * time.Second):
                fmt.Printf("Timeout waiting for response\n")
            }
        }
        
        fmt.Printf("Received %d new contacts from queries\n", len(newContacts))
        
        // Step 4: Add new contacts to candidate list
        contactMap := make(map[string]Contact)
        
        // Add existing candidates
        for _, contact := range candidates {
            contactMap[contact.ID.String()] = contact
        }
        
        // Add new contacts (avoiding duplicates)
        for _, contact := range newContacts {
            if _, exists := contactMap[contact.ID.String()]; !exists {
                contact.CalcDistance(targetID)
                contactMap[contact.ID.String()] = contact
            }
        }
        
        // Convert back to slice and sort by distance
        candidates = make([]Contact, 0, len(contactMap))
        for _, contact := range contactMap {
            candidates = append(candidates, contact)
        }
        
        sort.Slice(candidates, func(i, j int) bool {
            return candidates[i].distance.Less(candidates[j].distance)
        })
        
        // Keep only k closest candidates
        if len(candidates) > k {
            candidates = candidates[:k]
        }
        
        // Step 5: Check termination condition
        // Terminate if we haven't found any closer nodes
        if len(candidates) > 0 && candidates[0].distance.Less(closestDistance) {
            closestDistance = candidates[0].distance
            fmt.Printf("Found closer node: distance %s\n", closestDistance.String()[:8])
        } else {
            fmt.Printf("No closer nodes found, terminating lookup\n")
            break
        }
    }
    
    // Step 6: Ensure we've queried the k closest nodes we know about
    // This is the paper's requirement: "query and get responses from the k closest nodes"
    unqueriedClosest := 0
    for i, contact := range candidates {
        if i >= k {
            break
        }
        if !queried[contact.ID.String()] {
            unqueriedClosest++
        }
    }
    
    // If we have unqueried nodes among the k closest, query them
    if unqueriedClosest > 0 {
        fmt.Printf("Querying %d remaining closest nodes\n", unqueriedClosest)
        
        for i, contact := range candidates {
            if i >= k {
                break
            }
            contactKey := contact.ID.String()
            if !queried[contactKey] {
                queried[contactKey] = true
                contacts, err := n.SendFindNodeAndWait(contact.Address, targetID, 5*time.Second)
                if err == nil {
                    // Add any new contacts discovered
                    for _, newContact := range contacts {
                        newContact.CalcDistance(targetID)
                        // Only add if it would be in the k closest
                        if len(candidates) < k || newContact.distance.Less(candidates[k-1].distance) {
                            candidates = append(candidates, newContact)
                            sort.Slice(candidates, func(i, j int) bool {
                                return candidates[i].distance.Less(candidates[j].distance)
                            })
                            if len(candidates) > k {
                                candidates = candidates[:k]
                            }
                        }
                    }
                }
            }
        }
    }
    
    // Return the k closest contacts
    result := candidates
    if len(result) > k {
        result = result[:k]
    }
    
    fmt.Printf("Iterative lookup complete, returning %d contacts\n", len(result))
    return result
}

// parallelQuery sends FIND_NODE requests to multiple nodes concurrently
func (n *Node) parallelQuery(contacts []Contact, targetID *KademliaID) []Contact {
	resultChan := make(chan []Contact, len(contacts))
	timeout := 5 * time.Second
	
	// Send queries in parallel
	for _, contact := range contacts {
		go func(c Contact) {
			result := n.queryNodeWithTimeout(c, targetID, timeout)
			resultChan <- result
		}(contact)
	}
	
	// Collect results
	var allNewContacts []Contact
	for i := 0; i < len(contacts); i++ {
		select {
		case contacts := <-resultChan:
			allNewContacts = append(allNewContacts, contacts...)
		case <-time.After(timeout + time.Second):
			// Individual queries have their own timeout, this is just cleanup
			fmt.Printf("Warning: Query result not received in time\n")
		}
	}
	
	return allNewContacts
}

// queryNodeWithTimeout sends a FIND_NODE to a single node with timeout
func (n *Node) queryNodeWithTimeout(contact Contact, targetID *KademliaID, timeout time.Duration) []Contact {
    fmt.Printf("Querying node %s for target %s\n", contact.ID.String()[:8], targetID.String()[:8])
    
    contacts, err := n.SendFindNodeAndWait(contact.Address, targetID, timeout)
    if err != nil {
        fmt.Printf("Failed to query node %s: %v\n", contact.Address, err)
        return []Contact{}
    }
    
    fmt.Printf("Successfully received %d contacts from %s\n", len(contacts), contact.Address)
    return contacts
}

// mergeAndSortContacts combines candidate lists and sorts by distance to target
func (n *Node) mergeAndSortContacts(existing []Contact, new []Contact, targetID *KademliaID, maxCount int) []Contact {
	// Create a map to avoid duplicates
	contactMap := make(map[string]Contact)
	
	// Add existing contacts
	for _, contact := range existing {
		contactMap[contact.ID.String()] = contact
	}
	
	// Add new contacts
	for _, contact := range new {
		contactMap[contact.ID.String()] = contact
	}
	
	// Convert back to slice and calculate distances
	var allContacts []Contact
	for _, contact := range contactMap {
		contact.CalcDistance(targetID)
		allContacts = append(allContacts, contact)
	}
	
	// Sort by distance to target
	candidates := ContactCandidates{contacts: allContacts}
	candidates.Sort()
	
	// Return up to maxCount contacts
	result := candidates.GetContacts(min(maxCount, len(allContacts)))
	return result
}

// Add a method to use iterative lookup in your join process
func (n *Node) PerformSelfLookup() {
	fmt.Printf("Performing iterative self-lookup to discover network\n")
	contacts := n.IterativeFindNode(n.ID)
	
	// Add discovered contacts to routing table
	for _, contact := range contacts {
		n.routingTable.AddContact(contact)
	}
	
	fmt.Printf("Self-lookup complete, discovered %d contacts\n", len(contacts))
}

func (n *Node) SendFindNodeAndWait(targetAddr string, targetID *KademliaID, timeout time.Duration) ([]Contact, error) {
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
    
    // Cast transport to UDPTransport to access SendAndWaitForResponse
    udpTransport, ok := n.transport.(*UDPTransport)
    if !ok {
        return nil, fmt.Errorf("transport is not UDPTransport")
    }
    
    response, err := udpTransport.SendAndWaitForResponse(msg, targetAddr, timeout)
    if err != nil {
        return nil, fmt.Errorf("failed to get response: %v", err)
    }
    
    if response.Type != FIND_NODE_RESPONSE {
        return nil, fmt.Errorf("unexpected response type: %s", response.Type)
    }
    
    // Extract contacts from response using JSON marshaling approach
    dataBytes, err := json.Marshal(response.Data)
    if err != nil {
        return nil, fmt.Errorf("failed to marshal response data: %v", err)
    }
    
    var responseData FindNodeResponse
    err = json.Unmarshal(dataBytes, &responseData)
    if err != nil {
        return nil, fmt.Errorf("failed to unmarshal response data: %v", err)
    }
    
    fmt.Printf("DEBUG: Received %d contacts in FIND_NODE response\n", len(responseData.Contacts))
    return responseData.Contacts, nil
}

func NewDataStore() *DataStore {
	return &DataStore{
		data: make(map[string]string),
	}
}

// Store saves a key-value pair
func (ds *DataStore) Store(key, value string) {
	ds.mutex.Lock()
	defer ds.mutex.Unlock()
	ds.data[key] = value
	fmt.Printf("DEBUG: Stored key %s with value (length: %d)\n", key, len(value))
}

// Get retrieves a value by key
func (ds *DataStore) Get(key string) (string, bool) {
	ds.mutex.RLock()
	defer ds.mutex.RUnlock()
	value, exists := ds.data[key]
	return value, exists
}

// GetAllKeys returns all stored keys
func (ds *DataStore) GetAllKeys() []string {
	ds.mutex.RLock()
	defer ds.mutex.RUnlock()
	
	keys := make([]string, 0, len(ds.data))
	for key := range ds.data {
		keys = append(keys, key)
	}
	return keys
}

// HashKey creates a Kademlia ID from a string key
func HashKey(key string) *KademliaID {
	// Use SHA-1 to create a 160-bit hash (20 bytes)
	hasher := sha1.New()
	hasher.Write([]byte(key))
	hash := hasher.Sum(nil)
	
	// Convert to hex string and create KademliaID
	hexStr := hex.EncodeToString(hash)
	return NewKademliaID(hexStr)
}

// Add these methods to the Node struct (add to your existing node.go)

// StoreValue stores a key-value pair in the k closest nodes
func (n *Node) StoreValue(key, value string) error {
	fmt.Printf("Storing key-value pair: key=%s, value_length=%d\n", key, len(value))
	
	// Hash the key to get target ID
	targetID := HashKey(key)
	fmt.Printf("Target ID for key '%s': %s\n", key, targetID.String())
	
	// Find k closest nodes to the target
	closestNodes := n.IterativeFindNode(targetID)
	if len(closestNodes) == 0 {
		fmt.Printf("No nodes found to store the value\n")
		return fmt.Errorf("no nodes available for storage")
	}
	
	fmt.Printf("Found %d nodes for storage\n", len(closestNodes))
	
	// Store on ourselves if we're among the closest
	myDistance := n.ID.CalcDistance(targetID)
	shouldStoreLocally := true
	
	for _, contact := range closestNodes {
		if contact.distance.Less(myDistance) {
			shouldStoreLocally = false
			break
		}
	}
	
	successCount := 0
	
	// Store locally if we're close enough
	if shouldStoreLocally {
		n.dataStore.Store(key, value)
		successCount++
		fmt.Printf("Stored locally on node %s\n", n.ID.String()[:8])
	}
	
	// Send STORE RPCs to closest nodes (up to k nodes)
	maxNodes := min(len(closestNodes), n.config.K)
	for i := 0; i < maxNodes; i++ {
		contact := closestNodes[i]
		// Skip ourselves
		if contact.ID.Equals(n.ID) {
			continue
		}
		
		err := n.SendStore(contact.Address, key, value)
		if err != nil {
			fmt.Printf("Failed to store at %s: %v\n", contact.Address, err)
		} else {
			successCount++
			fmt.Printf("Successfully stored at node %s\n", contact.ID.String()[:8])
		}
	}
	
	fmt.Printf("Storage complete: %d successful stores\n", successCount)
	
	if successCount == 0 {
		return fmt.Errorf("failed to store value at any node")
	}
	
	return nil
}

// FindValue performs an iterative search for a value
func (n *Node) FindValue(key string) (string, error) {
	fmt.Printf("Finding value for key: %s\n", key)
	
	// Hash the key to get target ID
	targetID := HashKey(key)
	fmt.Printf("Target ID for key '%s': %s\n", key, targetID.String())
	
	// Check if we have it locally first
	if value, exists := n.dataStore.Get(key); exists {
		fmt.Printf("Found value locally\n")
		return value, nil
	}
	
	// Perform iterative lookup using FIND_VALUE
	return n.IterativeFindValue(targetID, key)
}

// SendStore sends a STORE RPC to a target address
func (n *Node) SendStore(targetAddr, key, value string) error {
	myContact := NewContact(n.ID, n.Address)
	
	msg := Message{
		Type:      STORE,
		MessageID: generateMessageID(),
		Sender:    myContact,
		Timestamp: time.Now().Unix(),
		Data: StoreData{
			Key:   key,
			Value: value,
		},
	}
	
	// Cast transport to UDPTransport to access SendAndWaitForResponse
	udpTransport, ok := n.transport.(*UDPTransport)
	if !ok {
		return fmt.Errorf("transport is not UDPTransport")
	}
	
	response, err := udpTransport.SendAndWaitForResponse(msg, targetAddr, 5*time.Second)
	if err != nil {
		return fmt.Errorf("failed to get STORE response: %v", err)
	}
	
	if response.Type != STORE_RESPONSE {
		return fmt.Errorf("unexpected response type: %s", response.Type)
	}
	
	fmt.Printf("DEBUG: Received STORE response from %s\n", targetAddr)
	return nil
}

// SendFindValue sends a FIND_VALUE RPC to a target address
func (n *Node) SendFindValue(targetAddr, key string) (*FindValueResponse, error) {
	myContact := NewContact(n.ID, n.Address)
	
	msg := Message{
		Type:      FIND_VALUE,
		MessageID: generateMessageID(),
		Sender:    myContact,
		Timestamp: time.Now().Unix(),
		Data: FindValueData{
			Key: key,
		},
	}
	
	// Cast transport to UDPTransport to access SendAndWaitForResponse
	udpTransport, ok := n.transport.(*UDPTransport)
	if !ok {
		return nil, fmt.Errorf("transport is not UDPTransport")
	}
	
	response, err := udpTransport.SendAndWaitForResponse(msg, targetAddr, 5*time.Second)
	if err != nil {
		return nil, fmt.Errorf("failed to get FIND_VALUE response: %v", err)
	}
	
	if response.Type != FIND_VALUE_RESPONSE {
		return nil, fmt.Errorf("unexpected response type: %s", response.Type)
	}
	
	// Parse response data
	responseData, err := parseFindValueResponse(response.Data)
	if err != nil {
		return nil, fmt.Errorf("failed to parse FIND_VALUE response: %v", err)
	}
	
	return responseData, nil
}

// IterativeFindValue performs iterative lookup for a value
func (n *Node) IterativeFindValue(targetID *KademliaID, key string) (string, error) {
	alpha := n.config.Alpha
	k := n.config.K
	
	fmt.Printf("Starting iterative FIND_VALUE for key %s (target: %s)\n", key, targetID.String()[:8])
	
	// Start with closest known contacts from our routing table
	candidates := n.routingTable.FindClosestContacts(targetID, k)
	if len(candidates) == 0 {
		fmt.Printf("No initial candidates found in routing table\n")
		return "", fmt.Errorf("no nodes available for lookup")
	}
	
	fmt.Printf("Starting with %d candidates from routing table\n", len(candidates))
	
	// Track which nodes we've already queried
	queried := make(map[string]bool)
	var queriedMutex sync.Mutex
	
	round := 0
	for {
		round++
		fmt.Printf("--- FIND_VALUE round %d ---\n", round)
		
		// Select up to alpha unqueried nodes that are closest to target
		var toQuery []Contact
		queriedMutex.Lock()
		for _, contact := range candidates {
			contactKey := contact.ID.String()
			if !queried[contactKey] && len(toQuery) < alpha {
				toQuery = append(toQuery, contact)
				queried[contactKey] = true
			}
		}
		queriedMutex.Unlock()
		
		if len(toQuery) == 0 {
			fmt.Printf("No more nodes to query, value not found\n")
			break
		}
		
		fmt.Printf("Querying %d nodes in parallel for value\n", len(toQuery))
		
		// Query nodes in parallel for the value
		value, found, newContacts := n.parallelFindValue(toQuery, key)
		
		if found {
			fmt.Printf("Value found! Returning result\n")
			return value, nil
		}
		
		fmt.Printf("Value not found, received %d new contacts\n", len(newContacts))
		
		// Merge new contacts with existing candidates
		candidates = n.mergeAndSortContacts(candidates, newContacts, targetID, k)
		
		// Prevent infinite loops
		if round > 10 {
			fmt.Printf("Max rounds reached, value not found\n")
			break
		}
	}
	
	return "", fmt.Errorf("value not found for key: %s", key)
}

// FindValueResult represents the result of a FIND_VALUE query
type FindValueResult struct {
	Value    string
	Found    bool
	Contacts []Contact
}

// parallelFindValue sends FIND_VALUE requests to multiple nodes concurrently
func (n *Node) parallelFindValue(contacts []Contact, key string) (string, bool, []Contact) {
	
	resultChan := make(chan FindValueResult, len(contacts))
	timeout := 5 * time.Second
	
	// Send queries in parallel
	for _, contact := range contacts {
		go func(c Contact) {
			result := n.queryNodeForValue(c, key, timeout)
			resultChan <- result
		}(contact)
	}
	
	// Collect results
	var allNewContacts []Contact
	for i := 0; i < len(contacts); i++ {
		select {
		case result := <-resultChan:
			if result.Found {
				// Found the value! Return immediately
				return result.Value, true, nil
			}
			allNewContacts = append(allNewContacts, result.Contacts...)
		case <-time.After(timeout + time.Second):
			fmt.Printf("Warning: FIND_VALUE result not received in time\n")
		}
	}
	
	return "", false, allNewContacts
}

// queryNodeForValue sends a FIND_VALUE to a single node with timeout
func (n *Node) queryNodeForValue(contact Contact, key string, timeout time.Duration) FindValueResult {
	fmt.Printf("Querying node %s for value (key: %s)\n", contact.ID.String()[:8], key)
	
	response, err := n.SendFindValue(contact.Address, key)
	if err != nil {
		fmt.Printf("Failed to query node %s for value: %v\n", contact.Address, err)
		return FindValueResult{Found: false}
	}
	
	if response.Found {
		fmt.Printf("SUCCESS: Found value at node %s\n", contact.Address)
		return FindValueResult{
			Value: response.Value,
			Found: true,
		}
	}
	
	fmt.Printf("Value not found at %s, received %d contacts\n", contact.Address, len(response.Contacts))
	return FindValueResult{
		Found:    false,
		Contacts: response.Contacts,
	}
}

// Helper function to parse FindValueResponse from interface{}
func parseFindValueResponse(data interface{}) (*FindValueResponse, error) {
	dataBytes, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal response data: %v", err)
	}
	
	var responseData FindValueResponse
	err = json.Unmarshal(dataBytes, &responseData)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal response data: %v", err)
	}
	
	return &responseData, nil
}

// Helper function for min
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func (n *Node) GetRoutingTable() RoutingTableManager {
    return n.routingTable
}

func (n *Node) GetDataStore() *DataStore {
    return n.dataStore
}