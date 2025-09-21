package kademlia

func LookupNode(rt RoutingTableManager, targetID *KademliaID) []Contact {
    return rt.FindClosestContacts(targetID, DefaultConfig().Alpha)
}