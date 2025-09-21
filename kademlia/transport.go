package kademlia

import (
	"encoding/json"
	"fmt"
	"net"
)

type MessageEnvelope struct {
	Message Message
	Sender  string
}

type UDPTransport struct {
	conn         *net.UDPConn
	messageQueue chan MessageEnvelope
	config       *Config
	handler      MessageHandler
}

func NewUDPTransport(config *Config, handler MessageHandler) *UDPTransport {
	return &UDPTransport{
		messageQueue: make(chan MessageEnvelope, 100),
		config:       config,
		handler:      handler,
	}
}

func (t *UDPTransport) Listen(address string) error {
	addr, err := net.ResolveUDPAddr("udp", address)
	if err != nil {
		return fmt.Errorf("failed to resolve UDP address: %v", err)
	}

	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		return fmt.Errorf("failed to listen on UDP: %v", err)
	}

	t.conn = conn
	fmt.Printf("UDP transport listening on %s\n", addr.String())

	// Start message processing goroutine
	go t.processMessages()

	// Handle incoming messages
	buffer := make([]byte, 4096)
	for {
		n, clientAddr, err := conn.ReadFromUDP(buffer)
		if err != nil {
			//fmt.Printf("Error reading UDP message: %v\n", err)
			continue
		}

		go t.handleIncomingMessage(buffer[:n], clientAddr)
	}
}

func (t *UDPTransport) handleIncomingMessage(data []byte, clientAddr *net.UDPAddr) {
	fmt.Printf("DEBUG: Received %d bytes from %s\n", len(data), clientAddr.String())
	fmt.Printf("DEBUG: Raw message data: %s\n", string(data))

	var msg Message
	if err := json.Unmarshal(data, &msg); err != nil {
		fmt.Printf("ERROR: Failed to unmarshal message: %v\n", err)
		return
	}

	fmt.Printf("SUCCESS: Received %s from %s (MessageID: %s)\n", msg.Type, clientAddr.String(), msg.MessageID)

	// Send to message queue for processing
	envelope := MessageEnvelope{
		Message: msg,
		Sender:  clientAddr.String(),
	}

	select {
	case t.messageQueue <- envelope:
		// Message queued successfully
	default:
		fmt.Printf("WARNING: Message queue full, dropping message\n")
	}
}

func (t *UDPTransport) processMessages() {
	for envelope := range t.messageQueue {
		response, err := t.handler.HandleMessage(envelope.Message, envelope.Sender)
		if err != nil {
			fmt.Printf("ERROR: Message handling failed: %v\n", err)
			continue
		}

		// Send response if one was generated
		if response.Type != 0 { // Assuming 0 is not a valid message type
			err = t.sendResponse(response, envelope.Sender)
			if err != nil {
				fmt.Printf("ERROR: Failed to send response: %v\n", err)
			}
		}
	}
}

func (t *UDPTransport) Send(msg Message, addr string) error {
	if t.conn == nil {
		return fmt.Errorf("transport not initialized")
	}

	udpAddr, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		return fmt.Errorf("failed to resolve address %s: %v", addr, err)
	}

	fmt.Printf("DEBUG: Sending %s to %s (MessageID: %s)\n", msg.Type, addr, msg.MessageID)

	data, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %v", err)
	}

	fmt.Printf("DEBUG: Marshaled message: %s\n", string(data))

	bytesWritten, err := t.conn.WriteToUDP(data, udpAddr)
	if err != nil {
		return fmt.Errorf("failed to send message: %v", err)
	}

	fmt.Printf("SUCCESS: Sent %d bytes to %s\n", bytesWritten, addr)
	return nil
}

func (t *UDPTransport) sendResponse(msg Message, addr string) error {
	udpAddr, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		return fmt.Errorf("failed to resolve address %s: %v", addr, err)
	}

	fmt.Printf("DEBUG: Sending %s response to %s\n", msg.Type, addr)

	data, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal response: %v", err)
	}

	bytesWritten, err := t.conn.WriteToUDP(data, udpAddr)
	if err != nil {
		return fmt.Errorf("failed to send response: %v", err)
	}

	fmt.Printf("SUCCESS: Sent %s response (%d bytes) to %s\n", msg.Type, bytesWritten, addr)
	return nil
}

func (t *UDPTransport) Close() error {
	close(t.messageQueue)
	if t.conn != nil {
		return t.conn.Close()
	}
	return nil
}