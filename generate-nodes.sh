#!/bin/bash

cat > docker-compose.generated.yml << 'EOF'
version: '3.8'
networks:
  kademlia-net:
    driver: bridge

services:
  node1:
    build: .
    networks:
      - kademlia-net
    stdin_open: true
    tty: true
    environment:
      - RUN_CLI=true
    command: ["./kademlia-node", "-port=8001"]
    ports:
      - "8001:8001"

EOF

# Generate nodes 2-50 with proper ports
for i in {2..50}; do
    port=$((8000 + i))  # This gives 8002, 8003... 8050
cat << EOF >> docker-compose.generated.yml
  node${i}:
    build: .
    networks:
      - kademlia-net
    stdin_open: true
    tty: true
    environment:
      - RUN_CLI=true
    command: ["./kademlia-node", "-port=${port}", "-target=node1:8001"]
    ports:
      - "${port}:${port}"
    depends_on:
      - node1

EOF
done