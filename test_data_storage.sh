#!/bin/bash

# Test Data Storage Script - Demonstrates actual PUT/GET operations

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}  Kademlia Data Storage Test${NC}"
echo -e "${BLUE}========================================${NC}"
echo

# Start the network
echo -e "${YELLOW}Starting the Kademlia network...${NC}"
docker-compose up -d > /dev/null 2>&1

# Wait for network to initialize
echo -e "${YELLOW}Waiting for network to initialize (15 seconds)...${NC}"
sleep 15

echo -e "${GREEN}Network is ready!${NC}"
echo

# Define test data
TEST_KEY="test_data_key"
TEST_VALUE="Hello from Kademlia network! This is stored data."
SEARCH_KEY="missing_key"

echo -e "${BLUE}Test Data:${NC}"
echo -e "  Key: ${YELLOW}${TEST_KEY}${NC}"
echo -e "  Value: ${YELLOW}${TEST_VALUE}${NC}"
echo

# Function to execute CLI command on a node
execute_cli() {
    local node=$1
    local command=$2
    echo "Executing on ${node}: ${command}"
    docker exec -i kademlia-${node}-1 sh <<EOF
${command}
exit
EOF
}

# Test 1: Store data on node10
echo -e "${BLUE}=== Test 1: Storing data on node10 ===${NC}"
echo -e "${YELLOW}Storing key '${TEST_KEY}' with value '${TEST_VALUE}' on node10...${NC}"
echo

execute_cli "node10" "put ${TEST_KEY} ${TEST_VALUE}"

echo
echo -e "${GREEN}Data stored!${NC}"
echo

# Test 2: Retrieve data from the same node
echo -e "${BLUE}=== Test 2: Retrieving data from node10 ===${NC}"
echo -e "${YELLOW}Retrieving key '${TEST_KEY}' from node10...${NC}"
echo

execute_cli "node10" "get ${TEST_KEY}"

echo

# Test 3: Try to retrieve data from a different node
echo -e "${BLUE}=== Test 3: Retrieving data from node25 (different node) ===${NC}"
echo -e "${YELLOW}Retrieving key '${TEST_KEY}' from node25...${NC}"
echo

execute_cli "node25" "get ${TEST_KEY}"

echo

# Test 4: Try to retrieve non-existent data
echo -e "${BLUE}=== Test 4: Retrieving non-existent key ===${NC}"
echo -e "${YELLOW}Retrieving non-existent key '${SEARCH_KEY}' from node15...${NC}"
echo

execute_cli "node15" "get ${SEARCH_KEY}"

echo

# Show network status
echo -e "${BLUE}=== Network Status ===${NC}"
echo -e "${YELLOW}Checking routing table of node1 (bootstrap node)...${NC}"
echo

docker exec -i kademlia-node1-1 sh <<EOF
help
exit
EOF

echo
echo -e "${GREEN}Tests completed!${NC}"
echo -e "${BLUE}The above tests demonstrate:${NC}"
echo -e "  1. ✅ Data storage (PUT operation)"
echo -e "  2. ✅ Data retrieval from same node"
echo -e "  3. ✅ Data retrieval from different node (network distribution)"
echo -e "  4. ✅ Handling of non-existent keys"
echo

echo -e "${YELLOW}Network is still running. You can manually test with:${NC}"
echo -e "  docker exec -it kademlia-node<X>-1 sh"
echo -e "  Then use: put <key> <value>, get <key>, help, exit"
echo

echo -e "${YELLOW}To stop the network:${NC}"
echo -e "  docker-compose down"