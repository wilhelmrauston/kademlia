#!/bin/bash

# Build the binary
go build -o kademlia-node

# Start 50 nodes in background
for i in {8001..8050}; do
    ./kademlia-node -port=$i &
    echo "Started node on port $i"
    sleep 0.1
done

# Wait a bit, then test connectivity
sleep 5

# Send some test pings
echo "Testing connectivity..."
./kademlia-node -port=9999 -target=localhost:8001 &
./kademlia-node -port=9998 -target=localhost:8025 &
./kademlia-node -port=9997 -target=localhost:8050 &

echo "Network is running. Check logs for message exchanges."