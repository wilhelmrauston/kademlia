package kademlia

type NetworkTransport interface {
    Listen(address string) error
    Send(msg Message, addr string) error
    Close() error
}

type RoutingTableManager interface {
    AddContact(contact Contact)
    FindClosestContacts(target *KademliaID, count int) []Contact
    GetAllContacts() []Contact
}

type MessageHandler interface {
    HandleMessage(msg Message, sender string) (Message, error)
}