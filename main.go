package main

import (
	"bufio"
	"fmt"
	"kademlia"
	"log"
	"os"
	"strings"
	"time"
)

func main() {
	fmt.Println("Docker node is running. Type 'put' followed by your command:")
	address := os.Getenv("ADDRESS")
	KademliaNode1 := kademlia.CreateKademliaNode(address)
	KademliaNode1.Start()
	time.Sleep(1 * time.Second)
	log.Printf("my address is: %s", KademliaNode1.RoutingTable.GetMe().Address)

	if address == "127.0.0.1:8001" {
		log.Printf("Bootstrap node started at %s", KademliaNode1.RoutingTable.GetMe().Address)
	} else {
		contact := kademlia.NewContact(kademlia.NewKademliaID("AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA"), "127.0.0.1:8001")
		KademliaNode1.RoutingTable.AddContact(contact)
		KademliaNode1.StartTask(&KademliaNode1.Network.ID, "LookupContact", "")
	}

	// Check if we're running in a terminal (interactive mode)
	if fileInfo, _ := os.Stdin.Stat(); (fileInfo.Mode() & os.ModeCharDevice) != 0 {
		// Interactive mode - use scanner
		log.Println("Running in interactive mode")
		scanner := bufio.NewScanner(os.Stdin)
		
		for {
			fmt.Print("> ")
			if scanner.Scan() {
				input := strings.TrimSpace(scanner.Text())
				parts := strings.Fields(input)

				if len(parts) == 2 {
					command := parts[0]
					argument := parts[1]

					if command == "put" {
						fileToStore := argument
						hashedFile := kademlia.HashKademliaID(fileToStore)
						KademliaNode1.StartTask(&hashedFile, "StoreValue", fileToStore)
					} else if command == "show" {
						fileToShow := argument
						hashedFile := kademlia.HashKademliaID(fileToShow)
						value, found := KademliaNode1.Storage.GetValue(hashedFile)
						log.Printf("value found? %s%t", value, found)
					} else if command == "get" {
						hashedFile := kademlia.NewKademliaID(argument)
						KademliaNode1.StartTask(hashedFile, "FindValue", "")
					} else {
						fmt.Println("Unknown command. Type command followed by an argument.")
					}
				} else if strings.HasPrefix(input, "exit") {
					break
				} else {
					fmt.Println("Unknown command. Type command followed by an argument.")
				}
			} else {
				break
			}
		}
	} else {
		// Non-interactive mode (Docker) - keep running indefinitely
		log.Println("Running in daemon mode (Docker)")
		log.Printf("Node %s is ready and listening...", address)
		
		// Keep the process alive forever
		select {} // This blocks forever
	}
}