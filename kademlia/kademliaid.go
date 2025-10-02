package kademlia

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"math/rand"
)

// IDLength the static number of bytes in a KademliaID
const IDLength = 20

// KademliaID type definition of a KademliaID
type KademliaID [IDLength]byte

// KademliaID type definition of a KademliaID

func HashKademliaID(input string) KademliaID {

	// Create a new SHA-1 hash
	hash := sha1.New()

	// Write the string to the hasher
	hash.Write([]byte(input))

	// Get the resulting 20-byte hash
	hashedBytes := hash.Sum(nil)

	return [IDLength]byte(hashedBytes)
}

func NewKademliaID(data string) *KademliaID {
	if len(data) != 40 { // Ensure the input is exactly 40 hex characters (20 bytes)
		error := fmt.Sprintf("Invalid KademliaID: input must be 40 hexadecimal characters: %s", data)
		panic(error)
	}

	decoded, err := hex.DecodeString(data)
	if err != nil {
		panic(fmt.Sprintf("Failed to decode KademliaID: %s", err))
	}

	newKademliaID := KademliaID{}
	for i := 0; i < IDLength; i++ {
		newKademliaID[i] = decoded[i]
	}

	return &newKademliaID
}

// NewRandomKademliaID returns a new instance of a random KademliaID,
// change this to a better version if you like
func NewRandomKademliaID() *KademliaID {
	newKademliaID := KademliaID{}
	for i := 0; i < IDLength; i++ {
		newKademliaID[i] = uint8(rand.Intn(256))
	}
	return &newKademliaID
}

// Less returns true if kademliaID < otherKademliaID (bitwise)
func (kademliaID KademliaID) Less(otherKademliaID *KademliaID) bool {
	for i := 0; i < IDLength; i++ {
		if kademliaID[i] != otherKademliaID[i] {
			return kademliaID[i] < otherKademliaID[i]
		}
	}
	return false
}

// Equals returns true if kademliaID == otherKademliaID (bitwise)
func (kademliaID KademliaID) Equals(otherKademliaID *KademliaID) bool {
	for i := 0; i < IDLength; i++ {
		if kademliaID[i] != otherKademliaID[i] {
			return false
		}
	}
	return true
}

// CalcDistance returns a new instance of a KademliaID that is built
// through a bitwise XOR operation betweeen kademliaID and target
func (kademliaID KademliaID) CalcDistance(target *KademliaID) *KademliaID {
	result := KademliaID{}
	for i := 0; i < IDLength; i++ {
		result[i] = kademliaID[i] ^ target[i]
	}
	return &result
}

// String returns a simple string representation of a KademliaID
func (kademliaID *KademliaID) String() string {
	return hex.EncodeToString(kademliaID[0:IDLength])
}