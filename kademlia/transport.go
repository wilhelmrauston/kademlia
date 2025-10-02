package kademlia

import (
	"encoding/json"
	"fmt"
	"net"
	"sync"
	"time"
)

type MessageEnvelope struct {
	Message Message
	Sender  string
}

type PendingRequest struct {
    MessageID    string
    ResponseChan chan Message
    Created      time.Time
}

type UDPTransport struct {
	conn         *net.UDPConn
	messageQueue chan MessageEnvelope
	config       *Config
	handler      MessageHandler
	pendingRequests map[string]*PendingRequest
    pendingMutex    sync.RWMutex
}

func NewUDPTransport(config *Config, handler MessageHandler) *UDPTransport {
    t := &UDPTransport{
        messageQueue:    make(chan MessageEnvelope, 100),
        config:          config,
        handler:         handler,
        pendingRequests: make(map[string]*PendingRequest),  // Add this
    }
    
    // Start cleanup goroutine for expired requests
    go t.cleanupExpiredRequests()  // Add this
    
    return t
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
	fmt.Printf("DEBUG: Received raw JSON: %s\n", string(data))
    
    var msg Message
    if err := json.Unmarshal(data, &msg); err != nil {
        fmt.Printf("ERROR: Failed to unmarshal message: %v\n", err)
        fmt.Printf("ERROR: Problematic JSON: %s\n", string(data))
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
        // Deliver to waiter if applicable
        _ = t.isResponse(envelope.Message) && t.handlePendingResponse(envelope.Message)
        // IMPORTANT: do NOT return/continue here — always fall through

        // Always let the handler see the message (requests *and* responses)
        response, err := t.handler.HandleMessage(envelope.Message, envelope.Sender)
        if err != nil {
            fmt.Printf("ERROR: Message handling failed: %v\n", err)
            continue
        }

        // Only requests should produce replies; your response-handlers should return zero-value Message
        if response.Type != 0 {
            if err := t.sendResponse(response, envelope.Sender); err != nil {
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

// Add these methods to your transport.go file:

func (t *UDPTransport) SendAndWaitForResponse(msg Message, addr string, timeout time.Duration) (Message, error) {
    responseChan := make(chan Message, 1)
    
    t.pendingMutex.Lock()
    t.pendingRequests[msg.MessageID] = &PendingRequest{
        MessageID:    msg.MessageID,
        ResponseChan: responseChan,
        Created:      time.Now(),
    }
    t.pendingMutex.Unlock()
    
    defer func() {
        t.pendingMutex.Lock()
        delete(t.pendingRequests, msg.MessageID)
        t.pendingMutex.Unlock()
        close(responseChan)
    }()
    
    err := t.Send(msg, addr)
    if err != nil {
        return Message{}, fmt.Errorf("failed to send message: %v", err)
    }
    
    select {
    case response := <-responseChan:
        return response, nil
    case <-time.After(timeout):
        return Message{}, fmt.Errorf("timeout waiting for response")
    }
}

func (t *UDPTransport) isResponse(msg Message) bool {
    return msg.Type == PONG || 
           msg.Type == FIND_NODE_RESPONSE || 
           msg.Type == FIND_VALUE_RESPONSE || 
           msg.Type == STORE_RESPONSE
}

func (t *UDPTransport) handlePendingResponse(msg Message) bool {
    t.pendingMutex.RLock()
    pending, exists := t.pendingRequests[msg.MessageID]
    t.pendingMutex.RUnlock()
    
    if !exists {
        return false
    }
    
    select {
    case pending.ResponseChan <- msg:
        fmt.Printf("DEBUG: Matched response %s to pending request\n", msg.MessageID)
        return true
    default:
        return false
    }
}

func (t *UDPTransport) cleanupExpiredRequests() {
    ticker := time.NewTicker(30 * time.Second)
    defer ticker.Stop()
    
    for range ticker.C {
        cutoff := time.Now().Add(-60 * time.Second)
        
        t.pendingMutex.Lock()
        for messageID, pending := range t.pendingRequests {
            if pending.Created.Before(cutoff) {
                delete(t.pendingRequests, messageID)
                close(pending.ResponseChan)
            }
        }
        t.pendingMutex.Unlock()
    }
}