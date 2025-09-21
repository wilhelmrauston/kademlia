package main

import (
	"bufio"
	"flag"
	"fmt"
	"kademlia"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

type AppConfig struct {
	Port          int
	BootstrapAddr string
	NodeID        string
	IP            string
	CLI           bool  // Add this field
}

func parseFlags() *AppConfig {
	config := &AppConfig{}
	flag.IntVar(&config.Port, "port", 8001, "Port to listen on")
	flag.StringVar(&config.BootstrapAddr, "target", "", "Bootstrap node address (host:port)")
	flag.StringVar(&config.NodeID, "id", "", "Node ID (optional, random if empty)")
	flag.BoolVar(&config.CLI, "cli", false, "Run interactive CLI")  // Add this line
	flag.Parse()
	config.IP = "0.0.0.0" // Listen on all interfaces for Docker
	return config
}

func main() {
	// Parse command line arguments
	appConfig := parseFlags()

	// Create Kademlia configuration
	kademliaConfig := kademlia.DefaultConfig()

	var address string
	if appConfig.BootstrapAddr == "" {
		// Bootstrap node - use node1 for external communication
		address = "node1:8001"  // This is what other nodes will use to contact us
	} else {
		// Regular nodes keep their current logic
		address = fmt.Sprintf("node%d:8000", appConfig.Port-8000)
	}

	// Initialize node with data store support
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

		// Use the join service for better network joining
		joinService := &kademlia.JoinService{
			Node:   node,
			Config: kademliaConfig,
		}

		err = joinService.JoinNetwork(appConfig.BootstrapAddr)
		if err != nil {
			fmt.Printf("Failed to join network: %v\n", err)
			fmt.Printf("Continuing as standalone node...\n")
		} else {
			fmt.Printf("Successfully joined network!\n")
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
			contacts := node.GetRoutingTable().GetAllContacts()
			keys := node.GetDataStore().GetAllKeys()
			
			fmt.Printf("INFO: Node %s has %d contacts, storing %d keys\n", 
				node.ID.String()[:8], len(contacts), len(keys))
		}
	}()

	/* // Demo: Store some test data periodically (only for non-bootstrap nodes)
	if appConfig.BootstrapAddr != "" {
		go func() {
			time.Sleep(10 * time.Second) // Wait for network to stabilize
			
			// Store a test value
			testKey := fmt.Sprintf("node-%d-message", appConfig.Port)
			testValue := fmt.Sprintf("Hello from node %d at %s", appConfig.Port, time.Now().Format("15:04:05"))
			
			fmt.Printf("Storing test data: %s -> %s\n", testKey, testValue)
			err := node.StoreValue(testKey, testValue)
			if err != nil {
				fmt.Printf("Failed to store test data: %v\n", err)
			} else {
				fmt.Printf("Successfully stored test data\n")
				
				// Try to retrieve it after a moment
				time.Sleep(2 * time.Second)
				value, err := node.FindValue(testKey)
				if err != nil {
					fmt.Printf("Failed to retrieve test data: %v\n", err)
				} else {
					fmt.Printf("Successfully retrieved: %s\n", value)
				}
			}
		}()
	} */
	// Check if we should run CLI interactively
	if appConfig.CLI || os.Getenv("RUN_CLI") == "true" {
		// Run CLI in background goroutine, not main thread
		fmt.Printf("Starting interactive CLI for node %s\n", node.ID.String()[:8])
		go startCLI(node)
	} else {
		fmt.Printf("Node %s is running. No CLI requested.\n", node.ID.String()[:8])
	}
	
	// Main thread always waits for shutdown signal
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

func startCLI(node *kademlia.Node) {
    scanner := bufio.NewScanner(os.Stdin)
    fmt.Println("Kademlia CLI ready. Commands: put <content>, get <hash>, exit")
    
    for {
        fmt.Print("> ")
        if !scanner.Scan() {
            break
        }
        
        input := strings.TrimSpace(scanner.Text())
        if input == "" {
            continue
        }
        
        parts := strings.SplitN(input, " ", 2)
        command := strings.ToLower(parts[0])
        
        switch command {
        case "put":
            if len(parts) < 2 {
                fmt.Println("Usage: put <content>")
                fmt.Println("Example: put \"Hello, World!\"")
                continue
            }
            
            content := parts[1]
            
            // Create hash of the content (this is the key)
            hash := kademlia.HashKey(content)
            
            fmt.Printf("Storing content with hash: %s\n", hash.String())
            
            // Store the content in the network
            err := node.StoreValue(hash.String(), content)
            if err != nil {
                fmt.Printf("Error storing content: %v\n", err)
            } else {
                fmt.Printf("Successfully stored content!\n")
                fmt.Printf("Hash: %s\n", hash.String())
            }
            
        case "get":
            if len(parts) < 2 {
                fmt.Println("Usage: get <hash>")
                fmt.Println("Example: get 1a2b3c4d5e...")
                continue
            }
            
            hash := strings.TrimSpace(parts[1])
            
            fmt.Printf("Looking up content for hash: %s\n", hash)
            
            // Retrieve the content from the network
            value, err := node.FindValue(hash)
            if err != nil {
                fmt.Printf("Error retrieving content: %v\n", err)
            } else {
                fmt.Printf("Successfully retrieved content!\n")
                fmt.Printf("Content: %s\n", value)
                
                // Show which node we retrieved it from (this would require modifying FindValue)
                // For now, we'll just show that it was found
                fmt.Printf("Retrieved from network\n")
            }
            
        case "exit", "quit", "q":
            fmt.Println("Exiting CLI... (node will continue running)")
            go func() {
                // Wait for shutdown signal in background
                c := make(chan os.Signal, 1)
                signal.Notify(c, os.Interrupt, syscall.SIGTERM)
                <-c
                fmt.Printf("\nShutdown signal received, cleaning up...\n")
                node.Stop()
                fmt.Printf("Shutdown complete\n")
                os.Exit(0)
            }()
            return
            
        case "help", "h":
            fmt.Println("Available commands:")
            fmt.Println("  put <content>  - Store content in the network and get its hash")
            fmt.Println("  get <hash>     - Retrieve content by its hash")
            fmt.Println("  status         - Show node status and routing table info")
            fmt.Println("  exit           - Exit the CLI")
            fmt.Println("  help           - Show this help message")
            
        case "status":
            // Show node information
            contacts := node.GetRoutingTable().GetAllContacts()
            keys := node.GetDataStore().GetAllKeys()
            
            fmt.Printf("=== Node Status ===\n")
            fmt.Printf("Node ID: %s\n", node.ID.String())
            fmt.Printf("Address: %s\n", node.Address)
            fmt.Printf("Known contacts: %d\n", len(contacts))
            fmt.Printf("Stored keys: %d\n", len(keys))
            
            if len(contacts) > 0 {
                fmt.Printf("\nFirst 5 contacts:\n")
                for i, contact := range contacts {
                    if i >= 5 {
                        break
                    }
                    fmt.Printf("  %s (%s)\n", contact.ID.String()[:16]+"...", contact.Address)
                }
            }
            
            if len(keys) > 0 {
                fmt.Printf("\nStored keys:\n")
                for i, key := range keys {
                    if i >= 5 {
                        fmt.Printf("  ... and %d more\n", len(keys)-5)
                        break
                    }
                    fmt.Printf("  %s\n", key)
                }
            }
            
        default:
            fmt.Printf("Unknown command: %s\n", command)
            fmt.Println("Type 'help' for available commands")
        }
    }
    
    if err := scanner.Err(); err != nil {
        fmt.Printf("Error reading input: %v\n", err)
    }
}