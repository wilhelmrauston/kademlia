#!/bin/bash

echo "🚀 Kademlia Interactive CLI Demo"
echo "================================="
echo "This demo starts the network in interactive mode for REAL PUT/GET operations"
echo ""

# Check if containers are already running in detached mode
running_count=$(docker compose ps -q | wc -l)
if [ "$running_count" -gt 0 ]; then
    echo "📋 Containers are already running in detached mode."
    echo "   🔄 Stopping them to restart in interactive mode..."
    docker-compose down
    echo "   ✅ Stopped detached containers"
fi

echo ""
echo "🚀 Starting Kademlia network in INTERACTIVE mode..."
echo "   📝 This allows direct CLI interaction with containers"
echo "   ⚠️  You'll see container logs - this is normal"
echo "   🎯 Once started, you can interact directly with any node"
echo ""
echo "📋 Instructions for testing:"
echo "   1. Wait for 'Node is running. Type 'help' for available commands.' messages"
echo "   2. Open a new terminal and run:"
echo "      docker exec -it kademlia-node10-1 sh"
echo "   3. In the container, test commands:"
echo "      put mykey myvalue"
echo "      get mykey"
echo "      exit"
echo "   4. Press Ctrl+C here to stop the network when done"
echo ""
echo "🔑 Test hashes you can use:"
echo "   Key 'demo_key_node1' → Hash: $(echo -n 'demo_key_node1' | sha1sum | cut -d' ' -f1)"
echo "   Key 'demo_key_node10' → Hash: $(echo -n 'demo_key_node10' | sha1sum | cut -d' ' -f1)"
echo "   Key 'testkey' → Hash: $(echo -n 'testkey' | sha1sum | cut -d' ' -f1)"
echo ""
echo "⏳ Starting in 3 seconds..."
sleep 3

# Start in foreground mode
echo "🌟 Network starting... (Press Ctrl+C to stop)"
docker-compose up