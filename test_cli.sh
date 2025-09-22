#!/bin/bash

# Start a CLI node that connects to the Docker network
echo "Starting CLI node..."
cd /home/alexander/kademlia

# Give the node time to connect and build routing table
timeout 60 go run main.go -port 8051 -target localhost:8001 <<EOF
help
put 1234567890abcdef1234567890abcdef12345678 Hello from distributed network test 1!
put abcdef1234567890abcdef1234567890abcdef12 Second test message in the DHT!
put fedcba0987654321fedcba0987654321fedcba09 Third distributed storage example!
get 1234567890abcdef1234567890abcdef12345678
get abcdef1234567890abcdef1234567890abcdef12
get fedcba0987654321fedcba0987654321fedcba09
exit
EOF