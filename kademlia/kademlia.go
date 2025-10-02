package kademlia

import (
	"fmt"
	"log"
	"math/rand"
	"strconv"
	"strings"
	"time"
)

const alpha = 3 //parallell contact nodes
const TTL = 200 //ms

//node
type Kademlia struct {
	Network                     *Network
	RoutingTable                *RoutingTable
	Tasks                       []Task
	Storage                     *Storage
	HandlePing                	func(contact *Contact, msg Message) // Inject the  function here
	HandlePong               	func(contact *Contact, msg Message)
	HandleLookup            	func(contact *Contact, msg Message)
	HandleReturnLookupContact 	func(contact *Contact, msg Message)
	HandleFindValue         	func(contact *Contact, msg Message)
	HandleReturnFindValue		func(contact *Contact, msg Message)
	HandleStoreValue			func(contact *Contact, msg Message)
	HandleReturnStoreValue  	func(contact *Contact, msg Message)
}

// NewKademlia creates and initializes a new Kademlia node
func NewKademlia(network *Network, routingTable *RoutingTable, taskList []Task, storage *Storage) *Kademlia {
	return &Kademlia{
		Network:      network,
		RoutingTable: routingTable,
		Tasks:        taskList,
		Storage:      storage,
	}
}

// CreateKademliaNode make a new kademlia node
func CreateKademliaNode(address string) *Kademlia {

	ID := NewRandomKademliaID()
	if address == "127.0.0.1:8001" {
		log.Printf("changing kademliaID")
		ID = NewKademliaID("AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA")
	}
	contact := NewContact(ID, address)
	routingTable := NewRoutingTable(contact)
	messageCh := make(chan Message)
	network := &Network{
		ID:        *ID,
		MessageCh: messageCh,
	}
	taskList := make([]Task, 0)
	storage := NewStorage()
	kademliaNode := NewKademlia(network, routingTable, taskList, storage)
	return kademliaNode
}

// Start starts the Kademlia node, processing incoming messages from the network channel
func (kademlia *Kademlia) Start() {
	// Start processing messages from the network's channel
	go func() {
		err := kademlia.Network.Listen(kademlia.RoutingTable.me)
		if err != nil {
			//log.Printf("Error in network listener: %v", err)
		}
	}()
	go kademlia.processMessages()
}

// StartLookupContact starts the lookup process
func (kademlia *Kademlia) StartTask(lookupTarget *KademliaID, commandType string, File string) {
	//Ska inte ta en recipient utan vilka som ska skickas till räknas ut av routingtable i guess
	commandID := NewCommandID()
	// log.Printf("(file: kademlia function: StartTask) Command id: %d ", commandID)
	task := kademlia.CreateTask(commandType, commandID, lookupTarget)
	task.File = File
	kademlia.Tasks = append(kademlia.Tasks, *task)
	///the task is also appended to the task list
	task.ClosestContacts = kademlia.RoutingTable.FindClosestContacts(lookupTarget, bucketSize)
	task.SortContactsByDistance()
	// lägger till de 20 närmsta noderna till closestContacts och sortera

	//skicka lookupmsg till de 3 närmsta och lägg till dessa 3 i ContactedNodes
	//och WaitingForReturns (med tiden msg skickades)
	limit := alpha
	if len(task.ClosestContacts) < alpha {
		limit = len(task.ClosestContacts)
	}
	if commandType == "StoreValue" {
		commandType = "LookupContact"
	}
	for i := 0; i < limit; i++ {
		waitingContact := WaitingContact{
			SentTime: time.Now(),              // Set the current time as SentTime
			Contact:  task.ClosestContacts[i], // Use the contact struct
		}
		task.WaitingForReturns = append(task.WaitingForReturns, waitingContact)
		task.ContactedNodes = append(task.ContactedNodes, task.ClosestContacts[i])

		//log.Printf("(file: kademlia function: StartTask) sending message %s:%s:%d:%s", commandType, kademlia.Network.ID.String(), commandID, lookupTarget.String())
		Message := fmt.Sprintf("%s:%s:%d:%s", commandType, kademlia.Network.ID.String(), commandID, lookupTarget.String())
		kademlia.Network.SendMessage(&task.ClosestContacts[i], Message)

		// //log.Printf("(File: kademlia: Function: StartLookupContact) lookupmessage:%s", lookupMessage)
		////log.Printf("(File: kademlia: Function: StartLookupContact) task.closestcontact[i].adress:%d", task.ClosestContacts[i].Address)

	}
}

func (kademlia *Kademlia) processMessages() {
	for {
		select {
		case msg := <-kademlia.Network.MessageCh: // If there's a message in the channel
			// Create a contact using the sender's ID and address
			contact := &Contact{
				ID:      NewKademliaID(msg.SenderID), // Convert the sender's ID to a KademliaID
				Address: msg.SenderAddress,           // The sender's IP and port
			}

			// Handle different message types based on the "Command" field
			switch msg.Command {
			case "ping":
				if kademlia.HandlePing != nil {
					kademlia.HandlePing(contact, msg)
				} else {
					kademlia.handlePing(contact, msg) // Default to actual function if no 
				}

			case "pong":
				if kademlia.HandlePong != nil {
					kademlia.HandlePong(contact, msg)
				} else {
					kademlia.handlePongMessage(contact, msg)
				}

			case "LookupContact":
				if kademlia.HandleLookup != nil {
					kademlia.HandleLookup(contact, msg)
				} else {
					kademlia.handleLookupContact(contact, msg)
				}

			case "returnLookupContact":
				if kademlia.HandleReturnLookupContact != nil {
					kademlia.HandleReturnLookupContact(contact, msg)
				} else {
					kademlia.handleReturnLookupContact(contact, msg)
				}

			case "FindValue":
				if kademlia.HandleFindValue != nil {
					kademlia.HandleFindValue(contact, msg)
				} else {
					kademlia.handleFindValue(contact, msg)
				}

			case "returnFindValue":
				if kademlia.HandleReturnFindValue != nil {
					kademlia.HandleReturnFindValue(contact, msg)
				} else {
					kademlia.handleReturnFindValue(contact, msg)
				}

			case "StoreValue":
				if kademlia.HandleStoreValue != nil {
					kademlia.HandleStoreValue(contact, msg)
				} else {
					kademlia.handleStoreValue(contact, msg)
				}

			case "returnStoreValue":
				if kademlia.HandleReturnStoreValue != nil {
					kademlia.HandleReturnStoreValue(contact, msg)
				} else {
					kademlia.handleReturnStoreValue(contact, msg)
				}

			default:
				// Handle unknown command types if necessary
			}

		case <-time.After(TTL * time.Millisecond): // If no message is received after 100 ms
			// Perform some other action when no messages are received
			kademlia.checkTTLs()
		}
	}
}

// checkTTLs checks the task list for contacts that haven't responded within the TTL limit
func (kademlia *Kademlia) checkTTLs() {
	ttl := TTL * time.Millisecond // Set the TTL limit

	for _, task := range kademlia.Tasks {
		var updatedWaitingForReturns []WaitingContact // To store contacts that are still within TTL

		for _, waitingContact := range task.WaitingForReturns {
			// Check if the contact has exceeded the TTL
			if time.Since(waitingContact.SentTime) > ttl {
				// Contact has timed out, remove it from WaitingForReturns
				//log.Printf("Contact %s in task %d has timed out", waitingContact.Contact.ID.String(), task.CommandID)
				task.RemoveContactFromWaitingForReturnsByTask(*waitingContact.Contact.ID)
				if task.CommandType == "ping" {
					bucketIndex := kademlia.RoutingTable.getBucketIndex(task.TargetID)
					bucket := kademlia.RoutingTable.buckets[bucketIndex]
					bucket.AddContact(task.ReplaceContact)

					kademlia.RemoveTask(task.CommandID)
					continue
				}
				// Try to find the next closest contact that hasn't been contacted yet
				for _, closestContact := range task.ClosestContacts {
					if !kademlia.isContactInList(task.ContactedNodes, closestContact) {
						// Send a new lookup to this uncontacted node
						task.ContactedNodes = append(task.ContactedNodes, closestContact)
						task.WaitingForReturns = append(task.WaitingForReturns, WaitingContact{
							SentTime: time.Now(),
							Contact:  closestContact,
						})
						lookupMessage := fmt.Sprintf("LookupContact:%s:%d:%s", kademlia.RoutingTable.me.ID.String(), task.CommandID, task.TargetID.String())
						kademlia.Network.SendMessage(&closestContact, lookupMessage)

						// Mark this contact as contacted and add it to WaitingForReturns

						//log.Printf("Sent LookupContact to %s", closestContact.Address)
						break // Stop looking for the next contact once one is found
					}
				}
			} else {
				// Contact is still within TTL, keep it in the list
				updatedWaitingForReturns = append(updatedWaitingForReturns, waitingContact)
			}
		}

		// Update the task's WaitingForReturns list to only include those that haven't timed out
		task.WaitingForReturns = updatedWaitingForReturns

		// Optionally, you could check if the task is now complete and remove it
		if len(task.WaitingForReturns) == 0 && task.AreClosestContactsContacted() {

			log.Printf("doing marktaskascompleted on the row below")
			kademlia.handleTaskCompletion(&task)
			log.Printf("%d", task.CommandID)

			kademlia.MarkTaskAsCompleted(task.CommandID)

		}
	}
}

// Helper function to check if a contact is in a given list
func (kademlia *Kademlia) isContactInList(contacts []Contact, contact Contact) bool {
	for _, c := range contacts {
		if c.ID.Equals(contact.ID) {
			return true
		}
	}
	return false
}

// handlePing processes a "ping" message
func (kademlia *Kademlia) handlePing(contact *Contact, msg Message) {
	//log.Printf("Received ping from %s", contact.Address)

	// Prepare the pong message with the appropriate format
	// The format will be "pong:<senderID>:<senderAddress>:pong"
	id := kademlia.RoutingTable.me.ID.String()

	pongMessage := fmt.Sprintf("pong:%s:%s:pong", id, msg.CommandID)

	// Send the pong message back to the contact
	kademlia.Network.SendMessage(contact, pongMessage)

	//log.Printf("Sent pong to %s", contact.Address)
}

func (kademlia *Kademlia) handlePongMessage(contact *Contact, msg Message) {
	// Parse the CommandID from the message
	commandID, err := strconv.Atoi(msg.CommandID)
	if err != nil {
		//log.Printf("Invalid CommandID in pong message: %s", msg.CommandID)
		return
	}

	// Try to find the task with the matching CommandID (you can omit `task` if it's not needed)
	if _, err := kademlia.FindTaskByCommandID(commandID); err != nil {
		//log.Printf("Task with CommandID %d not found", commandID)
		return
	}

	kademlia.RemoveTask(commandID)
	//log.Printf("Task with CommandID %d removed after pong received", commandID)

	// Use AddContact from the routing table to handle updating or adding the contact
	kademlia.RoutingTable.AddContact(*contact)
	//log.Printf("Contact %s added or updated in the routing table", contact.Address)
}

func (kademlia *Kademlia) handleLookupContact(contact *Contact, msg Message) {
	//log.Printf("(File: kademlia: Function: HandleLookupContact) Handling LookupContact from %s with target ID: %s and commandID: %s", contact.Address, msg.Command, msg.CommandID)

	// Find the bucketSize closest contacts to the target ID in the routing table
	closestContacts := kademlia.RoutingTable.FindClosestContacts(NewKademliaID(msg.CommandInfo), bucketSize)

	// Prepare the response message by concatenating the three closest contacts
	var responseMessage string
	myID := kademlia.RoutingTable.me.ID.String()
	for i, c := range closestContacts {

		contactStr := fmt.Sprintf("%s:%s", c.ID.String(), c.Address)

		// Append to the response message
		responseMessage += contactStr

		// Add a comma after each contact except the last one
		if i < len(closestContacts)-1 {
			responseMessage += ","
		}
	}

	// Send the response message back to the requesting contact
	// The command for the response is 'returnLookupContact'
	message := fmt.Sprintf("return%s:%s:%s:%s", msg.Command, myID, msg.CommandID, responseMessage)
	kademlia.Network.SendMessage(contact, message)
	//log.Printf("%s", message)

	kademlia.updateRoutingTable(contact)

	//log.Printf("(File: kademlia: Function: HandleLookupContact) Sent returnLookupContact to %s with contacts: %s", contact.Address, responseMessage)
}

// handleReturnLookupContact processes a "returnLookupContact" message
func (kademlia *Kademlia) handleReturnLookupContact(contact *Contact, msg Message) {
	contactStrings := strings.Split(msg.CommandInfo, ",")
	commandID, err := strconv.Atoi(msg.CommandID)
	if err != nil {
		log.Printf("Invalid command ID: %s", msg.CommandID)
		return
	}

	task, err := kademlia.FindTaskByCommandID(commandID)
	if task == nil {
		// log.Printf("Task not found for CommandID: %d", commandID)
		return
	}
	if err != nil {
		log.Printf("error: %e", err)
	}

	senderID := NewKademliaID(msg.SenderID)
	kademlia.processReturnedContacts(task, contactStrings)
	kademlia.RemoveContactFromWaitingForReturns(task.CommandID, *senderID)

	if len(task.WaitingForReturns) == 0 && task.AreClosestContactsContacted() {
		kademlia.handleTaskCompletion(task)
	} else {
		kademlia.sendNextLookup(task)
	}
}

// processReturnedContacts processes each returned contact and updates the routing table and task.
func (kademlia *Kademlia) processReturnedContacts(task *Task, contactStrings []string) {
	for _, contactStr := range contactStrings {
		parts := strings.Split(contactStr, ":")
		if len(parts) != 3 || parts[0] == "" {
			continue
		}

		newContact := NewContact(NewKademliaID(parts[0]), parts[1]+":"+parts[2])
		kademlia.updateRoutingTable(&newContact)

		if !task.IsContactInClosestContacts(newContact) {
			task.ClosestContacts = append(task.ClosestContacts, newContact)
			task.SortContactsByDistance()
		}
	}
}

// handleTaskCompletion handles the completion of a task.
func (kademlia *Kademlia) handleTaskCompletion(task *Task) {
	// log.Printf("(File: kademlia, Function: handleTaskCompletion) handleTastCompletion is running")
	// log.Printf("(File: kademlia, Function: handleTaskCompletion) task.commandtype: %s", task.CommandType)

	if task.CommandType == "StoreValue" {
		limit := min(bucketSize, len(task.ClosestContacts))
		log.Printf("value is stored now %s: ", task.TargetID.String())
		for i := 0; i < limit; i++ {

			storeMessage := fmt.Sprintf("StoreValue:%s:%d:%s", kademlia.Network.ID.String(), task.CommandID, task.File)
			//log.Printf("(File: kademlia: Function: handleTaskCompletion): with hash: %s file to store: %s", task.TargetID.String(), task.File)
			kademlia.Network.SendMessage(&task.ClosestContacts[i], storeMessage)
		}
	} else if task.CommandType == "FindValue" {
		log.Printf("(file: kademlia function: handleTaskCompletion) file not found: %s", task.TargetID)
	}
}

// sendNextLookup sends the next lookup message to an uncontacted node, if available.
func (kademlia *Kademlia) sendNextLookup(task *Task) {
	index := task.FindFirstNotContactedNodeIndex()
	if index == -1 {
		return // No uncontacted nodes available
	}

	nextContact := task.ClosestContacts[index]

	// Check if the contact has already been contacted
	if task.ContactIsContacted(nextContact) {
		log.Printf("Contact %s already contacted, skipping...", nextContact.Address)
		return // Skip this contact if it has already been contacted
	}

	// Mark the contact as being waited for
	waitingContact := WaitingContact{SentTime: time.Now(), Contact: nextContact}
	task.WaitingForReturns = append(task.WaitingForReturns, waitingContact)
	task.ContactedNodes = append(task.ContactedNodes, nextContact)

	// Prepare and send the lookup message
	if task.CommandType == "StoreValue" {
		lookupMessage := fmt.Sprintf("LookupContact:%s:%d:%s", kademlia.Network.ID.String(), task.CommandID, task.TargetID)
		kademlia.Network.SendMessage(&nextContact, lookupMessage)
		return
	}

	lookupMessage := fmt.Sprintf("%s:%s:%d:%s", task.CommandType, kademlia.Network.ID.String(), task.CommandID, task.TargetID)
	kademlia.Network.SendMessage(&nextContact, lookupMessage)
}

// min returns the smaller of two integers.
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// handleFindValue processes a "findValue" message
func (kademlia *Kademlia) handleFindValue(contact *Contact, msg Message) {
	// Parse the CommandID from the message
	commandID, err := strconv.Atoi(msg.CommandID)
	if err != nil {
		log.Printf("Invalid CommandID in findValue message: %s", msg.CommandID)
		return
	}

	// Compute the KademliaID for the requested value using the CommandInfo as the key
	hashKey := NewKademliaID(msg.CommandInfo)

	// Attempt to retrieve the value from storage
	value, found := kademlia.Storage.GetValue(*hashKey)
	//log.Printf("(file: kademlia function: handleFindValue) findValue recived looking for filewith hash: %s", hashKey)
	if found {
		// Value is found, send it back to the requesting contact
		response := fmt.Sprintf("returnFindValue:%s:%d:%s", kademlia.Network.ID.String(), commandID, value)
		kademlia.Network.SendMessage(contact, response)
		//log.Printf("Value found for %s, sending response to %s with the value: %s", msg.CommandInfo, contact.Address, value)
	} else {
		// Value not found, find the closest contacts to the requested value
		//log.Printf("(file: kademlia function: handleFindValue) file not found sending contacts insted")
		closestContacts := kademlia.RoutingTable.FindClosestContacts(hashKey, bucketSize)

		// Prepare the list of closest contacts in the format "<ID>:<Address>"
		var responseMessage string
		for i, c := range closestContacts {
			contactStr := fmt.Sprintf("%s:%s", c.ID.String(), c.Address)
			responseMessage += contactStr

			// Add a comma after each contact except the last one
			if i < len(closestContacts)-1 {
				responseMessage += ","
			}
		}

		// Send the closest contacts back to the requesting contact
		response := fmt.Sprintf("returnLookupContact:%s:%d:%s", kademlia.Network.ID.String(), commandID, responseMessage)
		kademlia.Network.SendMessage(contact, response)
	}
}

// handleReturnFindValue processes a "returnFindValue" message
func (kademlia *Kademlia) handleReturnFindValue(contact *Contact, msg Message) {
	commandID, err := strconv.Atoi(msg.CommandID)
	if err != nil {
		log.Printf("Invalid CommandID in findValue message: %s", msg.CommandID)
		return
	}

	task, err := kademlia.FindTaskByCommandID(commandID)
	if task == nil {
		log.Printf("Task not found for CommandID: %d", commandID)
		return
	}
	if err != nil {
		log.Printf("error: %e", err)
	}

	if *task.TargetID == HashKademliaID(msg.CommandInfo) {
		log.Printf("(File: kademlia: Function: handleReturnFindValue): The file was found: %s", msg.CommandInfo)
		kademlia.MarkTaskAsCompleted(task.CommandID)
	}
}

// handleStoreValue processes a "StoreValue" message
func (kademlia *Kademlia) handleStoreValue(contact *Contact, msg Message) {
	//log.Printf("(File: kademlia: Function: HandleStoreValue) Handling StoreValue from %s", contact.Address)
	// log.Printf("(File: kademlia, Function: handleStoreValue) handlestorevalue is running: %s", msg.CommandInfo)
	hashKey := HashKademliaID(msg.CommandInfo)
	// Extract the value to be stored from the CommandInfo field
	value := msg.CommandInfo

	// Store the value using the storage class
	//log.Printf("(File: kademlia: Function: handleStoreValue) msg.commandInfo: %s", msg.CommandInfo)
	//log.Printf("(File: kademlia: Function: handleStoreValue)hash value %x: actual value %s", hashKey, value)
	kademlia.Storage.StoreValue(hashKey, value)

	// Log the successful storage operation
	//log.Printf("(File: kademlia: Function: handleStoreValue)Stored value for KademliaID %s: %s", hashblabla.String(), value)
	message := fmt.Sprintf("returnStoreValue:%s:%s:%s", kademlia.Network.ID.String(), msg.CommandID, hashKey.String())
	kademlia.Network.SendMessage(contact, message)
}

// handleReturnStoreValue processes a "returnStoreValue" message
func (kademlia *Kademlia) handleReturnStoreValue(contact *Contact, msg Message) {
	commandID, err := strconv.Atoi(msg.CommandID)
	if err != nil {
		log.Printf("(File: kademlia: Function: handleReturnStoreValue error strconv )")
	}
	task, err := kademlia.FindTaskByCommandID(commandID)
	if err != nil {
		log.Printf("(File: kademlia: Function: handleReturnStoreValue) task not found")
	} else {
		task.NrNodesToStore++

		if task.NrNodesToStore == 1 {
			log.Printf("(File: kademlia: Function: handleReturnStoreValue) Hash: %s", msg.CommandInfo)
		}
		if task.NrNodesToStore == bucketSize {
			kademlia.MarkTaskAsCompleted(task.CommandID)
			log.Printf("(File: kademlia: Function: handleReturnStoreValue) bucketsize nodes has stored the value")
		}

	}
}

// func (kademlia *Kademlia) UpdateRoutingTable(newContact *Contact) {
// 	kademlia.updateRoutingTable(newContact)
// }

// TODO: check this
// If this doesn't work use the updateRoutingTable below
func (kademlia *Kademlia) updateRoutingTable(newContact *Contact) {
	if kademlia.RoutingTable.me.ID.Equals(newContact.ID) {
		return // Don't add yourself to the routing table.
	}

	bucketIndex := kademlia.RoutingTable.getBucketIndex(newContact.ID)
	bucket := kademlia.RoutingTable.buckets[bucketIndex]

	if kademlia.RoutingTable.IsContactInRoutingTable(newContact) {
		// Move the existing contact to the front (most recently used)
		bucket.AddContact(*newContact)
		return
	}

	if kademlia.RoutingTable.IsBucketFull(bucket) {
		// Ping the oldest contact if the bucket is full
		oldestElement := bucket.list.Back()
		if oldestElement != nil {
			oldContact := oldestElement.Value.(Contact)
			kademlia.CheckContactStatus(&oldContact, newContact)
		}
	} else {
		// Add the new contact since the bucket isn't full
		bucket.AddContact(*newContact)
	}
}

// CheckContactStatus creates a ping task for a contact and adds it to the task list.
func (kademlia *Kademlia) CheckContactStatus(oldContact *Contact, newContact *Contact) {
	// Generate a new command ID for the ping task
	commandID := NewCommandID()

	// Create a message string for the ping
	messageString := fmt.Sprintf("ping:%s:%d:ping", kademlia.RoutingTable.me.ID.String(), commandID)

	// Create a new task for this ping operation (without ClosestContacts or WaitingForReturns)
	task := Task{
		CommandType: "ping",
		CommandID:   commandID,
		TargetID:    oldContact.ID,
		StartTime:   time.Now(),
		// We omit ClosestContacts, ContactedNodes, and WaitingForReturns since ping doesn't need them.
		ReplaceContact: *newContact,
	}

	// Add the task to the node's task list
	kademlia.Tasks = append(kademlia.Tasks, task)

	// Send the ping message to the contact
	kademlia.Network.SendPingMessage(oldContact, messageString)

	//log.Printf("Ping task added for contact %s with CommandID %d", oldContact.Address, commandID)
}

// NewCommandID give a new command ID random int
func NewCommandID() int {
	return rand.Int()
}

// FindTaskByCommandID looks for a matching Task with the same CommandID in the task list.
func (kademlia *Kademlia) FindTaskByCommandID(commandID int) (*Task, error) {
	for i := range kademlia.Tasks {
		if kademlia.Tasks[i].CommandID == commandID {
			return &kademlia.Tasks[i], nil
		}
	}
	return nil, fmt.Errorf("task with CommandID %d not found", commandID)
}