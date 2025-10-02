package kademlia

import (
	"log"
	"net"
	"strings"
)

// Network is a node in the network
type Network struct {
	ID        KademliaID
	Conn      *net.UDPConn
	MessageCh chan Message
}

// Message is a simple struct to hold a message and its sender's address.
// Message is a simple struct to hold a command, command info, sender's ID, and sender's address.
type Message struct {
	Command       string
	CommandInfo   string
	CommandID     string
	SenderID      string
	SenderAddress string
}

// Listen listens on a UDP address and stores incoming messages in the channel.
func (network *Network) Listen(contact Contact) error {
	// Create a UDPAddr based on the Network's IP and Port
	udpAddr, err := net.ResolveUDPAddr("udp", contact.Address)
	if err != nil {
		log.Printf("Error resolving UDP address: %v\n", err)
		return err
	}

	// Start listening on the provided address
	conn, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		log.Printf("Error starting UDP listener: %v\n", err)
		return err
	}
	network.Conn = conn // Store the connection for sending messages
	defer conn.Close()

	// Loop to continuously listen for incoming messages
	for {
		buffer := make([]byte, 1024)
		n, remoteAddr, err := conn.ReadFromUDP(buffer)
		if err != nil {
			log.Printf("Error reading from UDP connection: %v", err)
			continue
		}

		// Convert the buffer into a string message
		message := string(buffer[:n])
		contactAddress := remoteAddr.String() // This is the sender's IP:Port

		// Print diagnostics for the message received
		//log.Printf("Message received: %s from %s", message, contactAddress)

		// Split the message by ":" to extract command, senderID, and command info
		parts := strings.SplitN(message, ":", 4) // Expect 4 parts now: <command>, <senderID>,<commandID>, <commandInfo>

		// Log the split parts for diagnosis
		// log.Printf("Message split into %d parts", len(parts))
		// for i, part := range parts {
		// 	log.Printf("Part[%d]: %s", i, part)
		// }

		// Validate the split result
		// if len(parts) != 4 {
		// 	log.Printf("Invalid message format received from %s: %s", contactAddress, message)
		// 	continue
		// }
		// for _, i := range parts {
		// 	log.Printf("printing part of parts %s:  ", i)
		// }
		// Extract the message parts
		command := parts[0]
		senderID := parts[1]
		commandID := parts[2]
		commandInfo := parts[3]

		// Log and send the message to the channel for processing
		//log.Printf("Network received message: %s from %s (ID: %s, commandInfo: %s)", command, contactAddress, senderID, commandInfo)
		network.MessageCh <- Message{
			Command:       command,
			CommandInfo:   commandInfo,
			CommandID:     commandID,
			SenderID:      senderID,
			SenderAddress: contactAddress, // Use the remote address (UDP sender address)
		}
	}
}

// SendMessage sends a message to a contact using the existing UDP connection.
// The message string is the final message to be sent, already formatted.
func (network *Network) SendMessage(contact *Contact, message string) {
	// Resolve the contact's address (expected to be in "IP:Port" format)
	// log.Printf("Network: sendmessage")
	// log.Printf("messegeString: %s", message)
	addr, err := net.ResolveUDPAddr("udp", contact.Address)
	if err != nil {
		log.Printf("Error resolving UDP address: %v", err)
		return
	}

	// Use the existing UDP connection to send the message
	_, err = network.Conn.WriteToUDP([]byte(message), addr)
	if err != nil {
		log.Printf("Error sending message to %s: %v", contact.Address, err)
		return
	}

	//log.Printf("Network sent message '%s' to %s", message, contact.Address)
}
func (network *Network) SendMessageWithAddress(Address string, message string) {
	log.Printf("Network: sendmessagewithAddress")

	addr, err := net.ResolveUDPAddr("udp", Address)
	if err != nil {
		log.Printf("Error resolving UDP address: %v", err)
		return
	}

	// Use the existing UDP connection to send the message
	_, err = network.Conn.WriteToUDP([]byte(message), addr)
	if err != nil {
		log.Printf("Error sending message to %s: %v", Address, err)
		return
	}

	//log.Printf("Network sent message '%s' to %s", message, contact.Address)
}

// SendPingMessage sends a "ping" message to the contact.
func (network *Network) SendPingMessage(contact *Contact, message string) {
	network.SendMessage(contact, message)
}

// SendPongMessage sends a "pong" message to the contact.
func (network *Network) SendPongMessage(contact *Contact, message string) {
	network.SendMessage(contact, message)
}