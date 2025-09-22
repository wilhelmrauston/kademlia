#!/bin/bash

# Start a CLI node that connects to the Docker network
echo "Starting CLI node and waiting for network connection..."
cd /home/alexander/kademlia

# Create a more interactive script that waits for network setup
{
  echo "help"
  sleep 5  # Give time for network discovery
  echo "put 1234567890abcdef1234567890abcdef12345678 Hello from distributed network test 1!"
  sleep 2
  echo "put abcdef1234567890abcdef1234567890abcdef12 Second test message in the DHT!"
  sleep 2  
  echo "put fedcba0987654321fedcba0987654321fedcba09 Third distributed storage example!"
  sleep 2
  echo "get c03c7680644bfabd7a233fa6b207663c8a444e1d"
  sleep 2
  echo "get ee32f7123d0b9fac408b11bea0ab027cdeed9759"
  sleep 2
  echo "get 779d463aca551a4a0c7b6e0be4e65b22ad138728"
  sleep 2
  echo "exit"
} | timeout 60 go run main.go -port 8051 -target localhost:8001