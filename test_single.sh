#!/bin/bash

echo "=== SINGLE INTERACTIVE CONTAINER DEMO ==="
echo "This method gives you REAL command line interaction"
echo ""
echo "Starting a standalone Kademlia node in interactive mode..."
echo "You can type 'put key value' and 'get key' commands directly!"
echo ""
echo "Commands you can use:"
echo "  put hello world"
echo "  get hello"
echo "  exit"
echo ""
echo "Press Ctrl+C to stop the demo, then try it yourself:"
echo "docker run -it --rm kademlia-node --port 8010"
echo ""

# Start the container in truly interactive mode
docker run -it --rm kademlia-node --port 8010