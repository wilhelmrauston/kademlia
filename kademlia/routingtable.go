package kademlia

const bucketSize int = 20

// RoutingTable definition
// keeps a refrence contact of me and an array of buckets
type RoutingTable struct {
	me      Contact
	buckets [IDLength * 8]*bucket
}

// NewRoutingTable returns a new instance of a RoutingTable
func NewRoutingTable(me Contact) *RoutingTable {
	routingTable := &RoutingTable{}
	for i := 0; i < IDLength*8; i++ {
		routingTable.buckets[i] = newBucket()
	}
	routingTable.me = me
	return routingTable
}

// GetMe returns the self contact
func (routingTable *RoutingTable) GetMe() *Contact {
	return &routingTable.me
}

// AddContact add a new contact to the correct Bucket (or move to front if it already exists)
func (routingTable *RoutingTable) AddContact(contact Contact) {
	//should the node be added? look at the bucket it should belong to
	// and see if the next node is alive.
	// add check if oldest contact is alive (should this be here or somewhere else?)
	bucketIndex := routingTable.getBucketIndex(contact.ID)
	bucket := routingTable.buckets[bucketIndex]
	bucket.AddContact(contact)
}

// FindClosestContacts finds the count closest Contacts to the target in the RoutingTable
func (routingTable *RoutingTable) FindClosestContacts(target *KademliaID, count int) []Contact {
	var candidates ContactCandidates
	bucketIndex := routingTable.getBucketIndex(target)
	bucket := routingTable.buckets[bucketIndex]

	candidates.Append(bucket.GetContactAndCalcDistance(target))

	for i := 1; (bucketIndex-i >= 0 || bucketIndex+i < IDLength*8) && candidates.Len() < count; i++ {
		if bucketIndex-i >= 0 {
			bucket = routingTable.buckets[bucketIndex-i]
			candidates.Append(bucket.GetContactAndCalcDistance(target))
		}
		if bucketIndex+i < IDLength*8 {
			bucket = routingTable.buckets[bucketIndex+i]
			candidates.Append(bucket.GetContactAndCalcDistance(target))
		}
	}

	candidates.Sort()

	if count > candidates.Len() {
		count = candidates.Len()
	}

	return candidates.GetContacts(count)
}

// getBucketIndex get the correct Bucket index for the KademliaID
func (routingTable *RoutingTable) getBucketIndex(id *KademliaID) int {
	distance := id.CalcDistance(routingTable.me.ID)
	for i := 0; i < IDLength; i++ {
		for j := 0; j < 8; j++ {
			if (distance[i]>>uint8(7-j))&0x1 != 0 {
				return i*8 + j
			}
		}
	}

	return IDLength*8 - 1
}
func (routingTable *RoutingTable) getBucket(index int) bucket {
	return *routingTable.buckets[index]
}

// IsContactInRoutingTable checks if a contact is in the routing table
func (routingTable *RoutingTable) IsContactInRoutingTable(contact *Contact) bool {
	index := routingTable.getBucketIndex(contact.ID)
	bucket := routingTable.buckets[index]
	return bucket.IsContactInBucket(contact)
}

// IsBucketFull checks if a bucket is full
func (routingTable *RoutingTable) IsBucketFull(bucket *bucket) bool {
	return bucket.Len() >= bucketSize
}