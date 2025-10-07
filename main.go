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
	NoInteractive bool
}

func parseFlags() *AppConfig {
	config := &AppConfig{}
	flag.IntVar(&config.Port, "port", 8001, "Port to listen on")
	flag.StringVar(&config.BootstrapAddr, "target", "", "Bootstrap node address (host:port)")
	flag.StringVar(&config.NodeID, "id", "", "Node ID (optional, random if empty)")
	flag.BoolVar(&config.NoInteractive, "daemon", false, "Run in daemon mode (no interactive CLI)")
	flag.Parse()

	config.IP = "0.0.0.0" // Listen on all interfaces for Docker
	return config
}

// startCLI starts the interactive command line interface
func startCLI(node *kademlia.Node) {
	// Set up graceful shutdown handling
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	// Channel to signal CLI exit
	cliExit := make(chan bool, 1)

	// Handle shutdown signals
	go func() {
		<-c
		fmt.Printf("\nShutdown signal received, cleaning up...\n")
		node.Stop()
		fmt.Printf("Shutdown complete\n")
		cliExit <- true
	}()

	// Start CLI loop
	go func() {
		scanner := bufio.NewScanner(os.Stdin)
		fmt.Print("> ")

		for scanner.Scan() {
			input := strings.TrimSpace(scanner.Text())
			if input == "" {
				fmt.Print("> ")
				continue
			}

			parts := strings.Split(input, " ")
			command := strings.ToLower(parts[0])

			switch command {
			case "help":
				showHelp()
			case "put":
				handlePutCommand(node, parts)
			case "get":
				handleGetCommand(node, parts)
			case "exit":
				fmt.Printf("Shutting down node...\n")
				node.Stop()
				fmt.Printf("Goodbye!\n")
				cliExit <- true
				return
			default:
				fmt.Printf("Unknown command: %s. Type 'help' for available commands.\n", command)
			}

			fmt.Print("> ")
		}
	}()

	// Wait for exit signal
	<-cliExit
}

// showHelp displays available CLI commands
func showHelp() {
	fmt.Printf("Available commands:\n")
	fmt.Printf("  put <content>  - Store content in the DHT and return its hash\n")
	fmt.Printf("  get <hash>     - Retrieve content by hash from the DHT\n")
	fmt.Printf("  exit           - Shutdown the node and exit\n")
	fmt.Printf("  help           - Show this help message\n")
}

// handlePutCommand processes the 'put' command
func handlePutCommand(node *kademlia.Node, parts []string) {
	if len(parts) < 2 {
		fmt.Printf("Error: put command requires content. Usage: put <content>\n")
		return
	}

	// Join all parts after 'put' as the content (handles spaces in content)
	content := strings.Join(parts[1:], " ")

	fmt.Printf("Storing content: %s\n", content)
	hash, err := node.SendStoreNetworkOnly(content)
	if err != nil {
		fmt.Printf("Error storing content: %v\n", err)
	} else {
		fmt.Printf("Content stored successfully!\n")
		fmt.Printf("Hash: %s\n", hash)
	}
}

// handleGetCommand processes the 'get' command
func handleGetCommand(node *kademlia.Node, parts []string) {
	if len(parts) != 2 {
		fmt.Printf("Error: get command requires exactly one hash. Usage: get <hash>\n")
		return
	}

	hash := parts[1]

	// Validate hash format (should be 40 character hex string for SHA-1)
	if len(hash) != 40 {
		fmt.Printf("Error: invalid hash format. Hash must be 40 characters long.\n")
		return
	}

	// Check if hash contains only valid hex characters
	for _, c := range hash {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')) {
			fmt.Printf("Error: invalid hash format. Hash must contain only hexadecimal characters.\n")
			return
		}
	}

	fmt.Printf("Looking for content with hash: %s\n", hash)

	// Try network first to test distribution (skip local for testing)
	content, found, err := node.SendFindValueNetworkOnly(hash)
	if err != nil {
		fmt.Printf("Error retrieving content: %v\n", err)
	} else if found {
		fmt.Printf("Content found in network!\n")
		fmt.Printf("Content: %s\n", content)
		fmt.Printf("Retrieved from: network\n")
		return
	}

	// Fall back to local if not found in network
	if content, found := node.GetValue(hash); found {
		fmt.Printf("Content found locally!\n")
		fmt.Printf("Content: %s\n", content)
		fmt.Printf("Retrieved from: local node (%s)\n", node.Address)
	} else {
		fmt.Printf("Content not found in DHT\n")
	}
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

	// Choose between interactive CLI or daemon mode
	if appConfig.NoInteractive {
		// Daemon mode for Docker containers
		fmt.Printf("Node is running in daemon mode. Press Ctrl+C to shutdown\n")
		waitForShutdown(node)
	} else {
		// Interactive CLI mode
		fmt.Printf("Node is running. Type 'help' for available commands.\n")
		startCLI(node)
	}
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
