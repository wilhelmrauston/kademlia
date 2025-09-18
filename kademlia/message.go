package kademlia

type MessageType int

const (
    PING MessageType = iota
    PONG
    FIND_NODE
    FIND_NODE_RESPONSE
    FIND_VALUE
    STORE
    JOIN_REQUEST
    JOIN_RESPONSE
)

func (mt MessageType) String() string {
	switch mt {
	case PING:
		return "PING"
	case PONG:
		return "PONG"
	case FIND_NODE:
		return "FIND_NODE"
	case FIND_NODE_RESPONSE:
		return "FIND_NODE_RESPONSE"
	case FIND_VALUE:
		return "FIND_VALUE"
	case STORE:
		return "STORE"
	case JOIN_REQUEST:
		return "JOIN_REQUEST"
	case JOIN_RESPONSE:
		return "JOIN_RESPONSE"
	default:
		return "UNKNOWN"
	}
}

type Message struct {
	Type      MessageType `json:"type"`
	MessageID string      `json:"message_id"`
	Sender    Contact     `json:"sender"`
	Target    *KademliaID `json:"target,omitempty"`
	Data      interface{} `json:"data,omitempty"`
	Timestamp int64       `json:"timestamp"`
}

type PingData struct {
	Message string `json:"message"`
}

type PongData struct {
	Message string `json:"message"`
}

type FindNodeData struct {
	TargetID *KademliaID `json:"target_id"`
}

type FindNodeResponse struct {
	Contacts []Contact `json:"contacts"`
}

type FindValueData struct {
	Key string `json:"key"`
}

type FindValueResponse struct {
	Found    bool      `json:"found"`
	Value    string    `json:"value,omitempty"`
	Contacts []Contact `json:"contacts,omitempty"`
}

type StoreData struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type JoinRequestData struct {
	NewNode Contact `json:"new_node"`
}

type JoinResponseData struct {
	Status   string    `json:"status"`
	Contacts []Contact `json:"contacts"`
}