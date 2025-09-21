// join_service.go
package kademlia

import (
	"fmt"
	"time"
)

type JoinService struct {
    Node   *Node
    Config *Config
}

func (js *JoinService) JoinNetwork(bootstrapAddr string) error {
    fmt.Printf("Joining network via %s\n", bootstrapAddr)
    
    // Step 1: Contact bootstrap node
    err := js.Node.SendPing(bootstrapAddr)
    if err != nil {
        return fmt.Errorf("failed to contact bootstrap node: %v", err)
    }
    
    // Step 2: Add bootstrap to routing table (this happens automatically in ping handler)
    time.Sleep(2 * time.Second)
    
    // Step 3: Perform lookup for own node ID (paper requirement)
    fmt.Printf("Performing self-lookup to populate routing table\n")
    contacts := js.Node.IterativeFindNode(js.Node.ID)
    
    // Add discovered contacts to routing table
    for _, contact := range contacts {
        js.Node.GetRoutingTable().AddContact(contact)
    }
    
    fmt.Printf("Self-lookup complete, discovered %d contacts\n", len(contacts))
    return nil
}

func (js *JoinService) startPeriodicMaintenance(bootstrapAddr string) {
    ticker := time.NewTicker(60 * time.Second)
    defer ticker.Stop()
    
    for range ticker.C {
        fmt.Printf("DEBUG: Sending periodic ping to bootstrap\n")
        err := js.Node.SendPing(bootstrapAddr)
        if err != nil {
            fmt.Printf("Periodic ping failed: %v\n", err)
        }
    }
}