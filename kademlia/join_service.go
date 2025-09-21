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
    
    // Contact bootstrap node
    err := js.Node.SendPing(bootstrapAddr)
    if err != nil {
        return fmt.Errorf("failed to contact bootstrap node: %v", err)
    }
    
    fmt.Printf("Successfully contacted bootstrap node\n")
    
    time.Sleep(3 * time.Second)
    
    fmt.Printf("Performing self-lookup to discover nearby nodes\n")
    err = js.Node.SendFindNode(bootstrapAddr, js.Node.ID)
    if err != nil {
        fmt.Printf("Failed to perform self-lookup: %v\n", err)
        // Don't fail completely, just log the error
    } else {
        fmt.Printf("Self-lookup request sent\n")
    }
    
    go js.startPeriodicMaintenance(bootstrapAddr)
    
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