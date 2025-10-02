package kademlia

import (
	"log"
	"sort"
	"strconv"
	"time"
)

// WaitingContact represents a contact we're waiting for a response from, along with the time the message was sent.
type WaitingContact struct {
	SentTime time.Time // The time when the message was sent
	Contact  Contact   // The contact we are waiting for a response from
}

// Task represents an ongoing task for the Kademlia node.
type Task struct {
	CommandType       string           // The type of task (e.g., "lookUpContact", "ping","findValue", "storevalue")
	CommandID         int              // The unique identifier for the command
	TargetID          *KademliaID      // The ID we're looking for (if applicable)
	StartTime         time.Time        // When the job was started (for timeouts)
	ClosestContacts   []Contact        // for storing nodes that are closest
	ContactedNodes    []Contact        // All the contacted nodes that we shouldn't contact again
	WaitingForReturns []WaitingContact // Alpha number of nodes that we're waiting for returns from
	ReplaceContact    Contact          // Contact to replace if no pong was recived
	File              string           //file/string that is to be stored
	NrNodesToStore    int              // NUmber of nodes that should store a value (bucket size)
}

// UpdateTaskFromMessage takes a Message struct and a contact, parses the information,
// and updates the task to the Kademlia node's task list.
func (kademlia *Kademlia) UpdateTaskFromMessage(msg Message, contact *Contact) {
	// Parse the message fields from the Message struct
	commandType := msg.Command
	commandID, err := strconv.Atoi(msg.CommandID)
	if err != nil {
		//log.Printf("Invalid command ID: %s", msg.CommandID)
		return
	}
	targetID := NewKademliaID(msg.CommandInfo) // Assuming CommandInfo is the targetID in this case

	// Create a new task and set its fields
	task := Task{
		CommandType:       commandType,
		CommandID:         commandID,
		TargetID:          targetID,
		StartTime:         time.Now(),
		ClosestContacts:   make([]Contact, 0),        // Initialize as an empty slice
		ContactedNodes:    make([]Contact, 0),        // Initialize as an empty slice
		WaitingForReturns: make([]WaitingContact, 0), // Initialize as an empty slice
	}

	// Log the creation of the task for debugging purposes
	//log.Printf("Task added: CommandType=%s, CommandID=%d, TargetID=%s", commandType, commandID, targetID.String())

	// Add the task to the node's task list (assuming you have a task list)
	kademlia.Tasks = append(kademlia.Tasks, task)
}

// Checks if a specific contact is in contacted nodes
func (task *Task) ContactIsContacted(contact Contact) bool {
	for _, c := range task.ContactedNodes {
		if c.ID == contact.ID {
			return true
		}
	}
	return false
}
func (task *Task) FindFirstNotContactedNodeIndex() int {
	limit := bucketSize

	// Ensure we don't go out of bounds if the list has fewer than 20 elements
	if len(task.ClosestContacts) < bucketSize {
		limit = len(task.ClosestContacts)
	}

	for i := 0; i < limit; i++ {
		if !task.ContactIsContacted(task.ClosestContacts[i]) {
			return i
		}
	}
	return -1
}
func (task *Task) AreClosestContactsContacted() bool {
	limit := bucketSize

	// Ensure we don't go out of bounds if the list has fewer than 20 elements
	if len(task.ClosestContacts) < bucketSize {
		limit = len(task.ClosestContacts)
	}

	// Check if each of the first 20 (or fewer) elements are in ContactedNodes
	for i := 0; i < limit; i++ {
		if !task.ContactIsContacted(task.ClosestContacts[i]) {
			return false
		}
	}
	return true
}

func (kademlia *Kademlia) CreateTask(commandType string, commandID int, targetID *KademliaID) *Task {

	//log.Printf("(File: task: Function: CreateTask)hash value %s: and commandType: %s", targetID, commandType)
	task := Task{
		CommandType:       commandType,
		CommandID:         commandID,
		TargetID:          targetID,
		StartTime:         time.Now(),
		ClosestContacts:   make([]Contact, 0),        // Initialize as an empty slice
		ContactedNodes:    make([]Contact, 0),        // Initialize as an empty slice
		WaitingForReturns: make([]WaitingContact, 0), // Initialize as an empty slice
	}

	// Log the creation of the task for debugging purposes
	//log.Printf("(File: Task, function: CreateTask) Task added: CommandType=%s, CommandID=%d, TargetID=%s", commandType, commandID, targetID.String())

	// Add the task to the node's task list (assuming you have a task list)
	return &task
}
func (task *Task) RemoveContactFromWaitingForReturnsByTask(contactID KademliaID) {
	// Remove contact from WaitingForReturns
	for i, waitingContact := range task.WaitingForReturns {
		if waitingContact.Contact.ID.Equals(&contactID) {
			// Remove the contact from the list
			task.WaitingForReturns = append(task.WaitingForReturns[:i], task.WaitingForReturns[i+1:]...)
			//log.Printf("Contact %s removed from WaitingForReturns in task %d", contactID.String(), commandID)
			break
		}
	}
}

// RemoveContactFromTask removes a contact from the WaitingForReturns list when they respond
func (kademlia *Kademlia) RemoveContactFromWaitingForReturns(commandID int, contactID KademliaID) {
	task, err := kademlia.FindTaskByCommandID(commandID)
	if err != nil {
		//log.Printf("Task with CommandID %d not found, cannot remove contact", commandID)
		return
	}

	// Remove contact from WaitingForReturns
	for i, waitingContact := range task.WaitingForReturns {
		if waitingContact.Contact.ID.Equals(&contactID) {
			// Remove the contact from the list
			task.WaitingForReturns = append(task.WaitingForReturns[:i], task.WaitingForReturns[i+1:]...)
			//log.Printf("Contact %s removed from WaitingForReturns in task %d", contactID.String(), commandID)
			break
		}
	}
}

// IsContactInClosestContacts checks if a given contact is in the ClosestContacts list for the task
func (task *Task) IsContactInClosestContacts(contact Contact) bool {
	for _, closestContact := range task.ClosestContacts {
		if closestContact.ID.Equals(contact.ID) {
			return true
		}
	}
	return false

}

// RemoveContactFromTask removes a contact from the WaitingForReturns list when they respond
func (kademlia *Kademlia) RemoveContactFromTask(commandID int, contact Contact) {
	task, err := kademlia.FindTaskByCommandID(commandID)
	if err != nil {
		//log.Printf("Task with CommandID %d not found, cannot remove contact", commandID)
		return
	}

	// Remove contact from WaitingForReturns
	for i, waitingContact := range task.WaitingForReturns {
		if waitingContact.Contact.ID.Equals(contact.ID) {
			// Remove the contact from the list
			task.WaitingForReturns = append(task.WaitingForReturns[:i], task.WaitingForReturns[i+1:]...)
			//log.Printf("Contact %s removed from WaitingForReturns in task %d", contact.ID.String(), commandID)
			break
		}
	}
}

// MarkTaskAsCompleted updates the task when all contacts have responded or it is considered done
func (kademlia *Kademlia) MarkTaskAsCompleted(commandID int) int {
	task, err := kademlia.FindTaskByCommandID(commandID)
	if err != nil {
		log.Printf("Task with CommandID %d not found", commandID)
		return -1
	}

	// Task is completed when no more waiting for returns
	if len(task.WaitingForReturns) == 0 {
		if task.CommandType == "StoreValue" {
			// kademlia.handleTaskCompletion(task)
		}
	}
	kademlia.RemoveTask(commandID)
	return 1
}
func (kademlia *Kademlia) RemoveTask(commandID int) {
	var remainingTasks []Task
	for _, task := range kademlia.Tasks {
		if task.CommandID != commandID {
			remainingTasks = append(remainingTasks, task)

		}
	}
	kademlia.Tasks = remainingTasks

}

// SortContactsByDistance sorts the contacts in the task's ClosestContacts based on their XOR distance to the TargetID.
func (task *Task) SortContactsByDistance() {
	if task.TargetID == nil {
		//log.Printf("Task %d has no TargetID, unable to sort contacts", task.CommandID)
		return
	}

	// Define a custom sort function that compares XOR distances
	sort.Slice(task.ClosestContacts, func(i, j int) bool {
		distanceToI := task.ClosestContacts[i].ID.CalcDistance(task.TargetID)
		distanceToJ := task.ClosestContacts[j].ID.CalcDistance(task.TargetID)

		// Compare distances: return true if the distance to I is less than the distance to J
		return distanceToI.Less(distanceToJ)
	})

	//log.Printf("(File: task, Function: SortContactsByDistance) Contacts in Task %d sorted by distance to TargetID", task.CommandID)
}

// RemoveTask removes a completed task by commandID

// HasTaskTimedOut checks whether a task has timed out based on a given duration.
func (task *Task) HasTaskTimedOut(timeout time.Duration) bool {
	return time.Since(task.StartTime) > timeout
}