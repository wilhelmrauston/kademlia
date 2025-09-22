package kademlia

func LookupNode(rt RoutingTableManager, targetID *KademliaID) []Contact {
	return rt.FindClosestContacts(targetID, DefaultConfig().Alpha)
}

// HashToKademliaID converts a hash string to a KademliaID
func HashToKademliaID(hash string) *KademliaID {
	return NewKademliaID(hash)
}
