#!/bin/bash

echo "=== NETWORK WITH INTERACTIVE NODE DEMO ==="
echo "This method starts a full network + one interactive node"
echo ""

# Start the background network first (detached)
echo "Step 1: Starting background network..."
docker-compose up -d node1 node2 node3 node4 node5

echo "Waiting for network to initialize..."
sleep 3

echo ""
echo "Step 2: Starting interactive node that can join the network..."
echo "You can now use commands:"
echo "  put distributed_key some_value"
echo "  get distributed_key"
echo "  exit"
echo ""
echo "Press Ctrl+C to stop, then run:"
echo "docker-compose down"
echo ""

# Start one interactive node that connects to the network
docker run -it --rm --network kademlia_default \
  kademlia-node --port 8060 --contact node1:8001

echo ""
echo "Cleaning up network..."
docker-compose down