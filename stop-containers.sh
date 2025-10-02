#!/bin/bash

echo "Stopping Kademlia containers..."

NUM_CONTAINERS=50

# Stop all containers
echo "Stopping containers..."
for i in $(seq 1 $NUM_CONTAINERS); do
    NAME="container$i"
    echo "Stopping $NAME..."
    
    if docker stop "$NAME" >/dev/null 2>&1; then
        echo "  $NAME stopped"
    else
        echo "  $NAME not running or already stopped"
    fi
done

echo ""
echo "Cleanup complete!"
echo "All Kademlia containers have been stopped."