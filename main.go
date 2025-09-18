package main

import (
	"flag"
	"fmt"
	"kademlia"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type AppConfig struct {
	Port          int
	BootstrapAddr string
	NodeID        string
	IP            string
}

func parseFlags() *AppConfig {
	config := &AppConfig{}
	flag.IntVar(&config.Port, "port", 8001, "Port to listen on")
	flag.StringVar(&config.BootstrapAddr, "target", "", "Bootstrap node address (host:port)")
	flag.StringVar(&config.NodeID, "id", "", "Node ID (optional, random if empty)")
	flag.Parse()
	
	config.IP = "0.0.0.0" // Listen on all interfaces for Docker
	return config
}

func main() {
	// Parse command line arguments
	appConfig := parseFlags()
	
	// Create Kademlia configuration
	kademliaConfig := kademlia.DefaultConfig()
	
	// Create node address
	address := fmt.Sprintf("%s:%d", appConfig.IP, appConfig.Port)
	
	// Initialize node
	fmt.Printf("Initializing Kademlia node...\n")
	node := kademlia.NewNode(address, appConfig.NodeID, kademliaConfig)
	fmt.Printf("Node ID: %s\n", node.ID.String())
	fmt.Printf("Node Address: %s\n", node.Address)
	
	// Start the node
	err := node.Start()
	if err != nil {
		log.Fatalf("Failed to start node: %v", err)
	}
	
	// Join network if bootstrap address provided
	if appConfig.BootstrapAddr != "" {
		fmt.Printf("Joining network via %s\n", appConfig.BootstrapAddr)
		
		// Give the node a moment to fully start
		time.Sleep(2 * time.Second)
		
		// Initial ping to bootstrap node
		err = node.SendPing(appConfig.BootstrapAddr)
		if err != nil {
			fmt.Printf("Failed to ping bootstrap node: %v\n", err)
		} else {
			fmt.Printf("Successfully contacted bootstrap node\n")
		}
		
		// Wait for ping/pong exchange, then do self-lookup
		time.Sleep(3 * time.Second)
		
		fmt.Printf("Performing self-lookup to discover nearby nodes\n")
		err = node.SendFindNode(appConfig.BootstrapAddr, node.ID)
		if err != nil {
			fmt.Printf("Failed to perform self-lookup: %v\n", err)
		} else {
			fmt.Printf("Self-lookup request sent\n")
		}
		
		// Start periodic maintenance
		go startPeriodicMaintenance(node, appConfig.BootstrapAddr)
	} else {
		fmt.Printf("No bootstrap address provided, running as bootstrap node\n")
	}
	
	// Print routing table info periodically
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()
		for range ticker.C {
			fmt.Printf("INFO: %s\n", node.GetRoutingTableInfo())
		}
	}()
	
	// Wait for shutdown signal
	fmt.Printf("Node is running. Press Ctrl+C to shutdown\n")
	waitForShutdown(node)
}

func startPeriodicMaintenance(node *kademlia.Node, bootstrapAddr string) {
	ticker := time.NewTicker(60 * time.Second)
	defer ticker.Stop()
	
	for range ticker.C {
		fmt.Printf("DEBUG: Sending periodic ping to bootstrap\n")
		err := node.SendPing(bootstrapAddr)
		if err != nil {
			fmt.Printf("Periodic ping failed: %v\n", err)
		}
	}
}

func waitForShutdown(node *kademlia.Node) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c
	
	fmt.Printf("\nShutdown signal received, cleaning up...\n")
	node.Stop()
	fmt.Printf("Shutdown complete\n")
}