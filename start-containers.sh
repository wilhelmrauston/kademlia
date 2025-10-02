#!/bin/bash

echo "Starting Kademlia containers (with keep-alive fix)..."

# Build the image first
echo "Building Docker image..."
docker build -t kadlab .
if [ $? -ne 0 ]; then
    echo "ERROR: Failed to build Docker image!"
    exit 1
fi

# Configuration
START_PORT=8001
NUM_CONTAINERS=50

# Clean up existing containers first
echo "Cleaning up existing containers..."
for i in $(seq 1 $NUM_CONTAINERS); do
    docker stop container$i >/dev/null 2>&1
    docker rm container$i >/dev/null 2>&1
done

echo "Starting containers with keep-alive..."

# Create containers with a different approach to keep them alive
for i in $(seq 1 $NUM_CONTAINERS); do
    PORT=$((START_PORT + i - 1))
    NAME="container$i"
    echo "Starting $NAME on port $PORT"
    
    # Run with stdin kept open and tty
    docker run -d -it --name "$NAME" --network host -e ID="$NAME" -e ADDRESS="127.0.0.1:$PORT" kadlab
    
    if [ $? -ne 0 ]; then
        echo "ERROR: Failed to start $NAME"
        exit 1
    fi
    
    # Small delay
    sleep 0.1
done

# Wait a moment then check how many are still running
sleep 5
RUNNING=$(docker ps | grep container | wc -l)

echo ""
echo "Containers started: $NUM_CONTAINERS"
echo "Containers still running: $RUNNING"

if [ "$RUNNING" -gt 0 ]; then
    echo "SUCCESS: $RUNNING containers are running!"
    echo ""
    echo "Test commands:"
    echo "  docker exec -it container1 sh                 - Attach to container1"
    echo "  docker logs container1                        - View container1 logs"
    echo "  ./stop-containers.sh                          - Stop all containers"
else
    echo "ERROR: All containers exited. Check logs:"
    echo "  docker logs container1"
    echo "  docker logs container2"
fi
echo ""