package kademlia

import "time"

type Config struct {
    K           int    // bucket size
    Alpha       int    // concurrency parameter
    IDLength    int    // 160 bits
    PingTimeout time.Duration
    JoinTimeout time.Duration
}

func DefaultConfig() *Config {
    return &Config{
        K:           20,
        Alpha:       3,
        IDLength:    20,
        PingTimeout: 5 * time.Second,
        JoinTimeout: 30 * time.Second,
    }
}