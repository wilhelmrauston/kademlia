#!/bin/bash

# Docker Compose Management Script for 50-Node Kademlia Network
# This script helps manage the 50-node Docker network for M5 requirement

echo "=== Kademlia 50-Node Docker Network Management ==="
echo "M5 Requirement: 50 nodes in containers"
echo ""

case "$1" in
    "start"|"up")
        echo "Starting 50-node Kademlia network..."
        echo "This will create 50 Docker containers (node1-node50)"
        echo "Ports: 8001-8050"
        echo ""
        docker-compose up -d
        echo ""
        echo "Network started! Checking status..."
        docker-compose ps | head -10
        echo "... (total 50 nodes)"
        echo ""
        echo "To see all nodes: docker-compose ps"
        echo "To see logs: docker-compose logs node1 (or any node)"
        ;;
    
    "stop"|"down")
        echo "Stopping 50-node Kademlia network..."
        docker-compose down
        echo "Network stopped."
        ;;
    
    "status"|"ps")
        echo "Kademlia Network Status:"
        echo "Total nodes configured: $(grep -c '^  node[0-9]' docker-compose.yml)"
        echo ""
        docker-compose ps
        ;;
    
    "logs")
        if [ -z "$2" ]; then
            echo "Showing logs for bootstrap node (node1):"
            docker-compose logs node1
        else
            echo "Showing logs for $2:"
            docker-compose logs "$2"
        fi
        ;;
    
    "test")
        echo "Testing network connectivity..."
        echo "Checking if node1 (bootstrap) is responding:"
        curl -m 5 http://localhost:8001/ 2>/dev/null || echo "Node1 not responding (this is expected - no HTTP endpoint)"
        echo ""
        echo "Checking Docker containers:"
        docker-compose ps | grep -c "Up" | xargs -I {} echo "{} containers are running"
        ;;
    
    "count")
        echo "Node Configuration Count:"
        echo "Nodes in docker-compose.yml: $(grep -c '^  node[0-9]' docker-compose.yml)"
        echo "Running containers: $(docker-compose ps -q | wc -l)"
        ;;
    
    *)
        echo "Usage: $0 {start|stop|status|logs [node]|test|count}"
        echo ""
        echo "Commands:"
        echo "  start   - Start the 50-node network"
        echo "  stop    - Stop the network"
        echo "  status  - Show container status"
        echo "  logs    - Show logs (default: node1, or specify node)"
        echo "  test    - Test network connectivity"
        echo "  count   - Show node counts"
        echo ""
        echo "Examples:"
        echo "  $0 start"
        echo "  $0 logs node5"
        echo "  $0 status"
        exit 1
        ;;
esac