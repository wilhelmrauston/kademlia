#!/bin/bash

echo "=== COMMAND INJECTION VIA DOCKER EXEC ==="
echo "This method uses running containers and injects commands"
echo ""

# Start a small network
echo "Step 1: Starting network..."
docker-compose up -d node1 node2 node3

echo "Waiting for network to stabilize..."
sleep 5

echo ""
echo "Step 2: Injecting commands into running containers..."

# Method 3A: Using echo and pipes
echo "Method 3A: Using echo and pipes"
echo "put test_key test_value" | docker exec -i kademlia-node1-1 sh -c 'cat > /tmp/cmd && timeout 2 cat /tmp/cmd | ./kademlia-node --port 8001 || true'

sleep 2

echo "get test_key" | docker exec -i kademlia-node2-1 sh -c 'cat > /tmp/cmd && timeout 2 cat /tmp/cmd | ./kademlia-node --port 8002 --contact node1:8001 || true'

echo ""
echo "Method 3B: Using expect (if available)"
if command -v expect &> /dev/null; then
    echo "Using expect for automated interaction..."
    expect << 'EOF'
spawn docker exec -it kademlia-node3-1 ./kademlia-node --port 8003 --contact node1:8001
expect ">"
send "put automated_key automated_value\r"
expect ">"
send "get automated_key\r"
expect ">"
send "exit\r"
expect eof
EOF
else
    echo "Expect not available - would need to install it"
fi

echo ""
echo "Method 3C: Using timeout and background processes"
(echo "put background_key background_value"; sleep 1; echo "get background_key"; sleep 1; echo "exit") | \
timeout 10 docker exec -i kademlia-node3-1 ./kademlia-node --port 8003 --contact node1:8001 || true

echo ""
echo "Cleaning up..."
docker-compose down

echo ""
echo "=== SUMMARY ==="
echo "Method 1: docker run -it --rm kademlia-node --port 8010"
echo "Method 2: Start network + interactive node"
echo "Method 3: Use docker exec with command injection"