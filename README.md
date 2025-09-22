# Kademlia DHT Implementation

A distributed hash table (DHT) implementation based on the Kademlia protocol, written in Go. This implementation provides peer-to-peer networking, object storage and retrieval, and supports large-scale distributed networks.

## Features

- **M1: Network Formation** ✅ - Bootstrap and join Kademlia networks
- **M2: Object Distribution** ✅ - Store and retrieve key-value pairs across the network
- **M4: Unit Testing** ✅ - Comprehensive test suite for all components
- **M5: Containerization** ✅ - Docker support for 50-node networks
- **M7: Thread Safety** ✅ - Concurrent operations with proper synchronization

## Project Structure

```
kademlia/
├── main.go                           # Application entry point
├── go.mod                            # Main module dependencies
├── go.sum                            # Dependency checksums
├── kademlia-node                     # Compiled binary
├── docker-compose.yml                # 50-node Docker network configuration
├── Dockerfile                        # Container build configuration
├── docker-network.sh                 # Docker network management script
├── start-network.sh                  # Network startup script
├── README.md                         # Project documentation
├── .gitignore                        # Git ignore rules
├── .github/                          # GitHub configuration
│   └── workflows/                    # CI/CD workflows
│       ├── docker_master.yaml        # Docker build workflow
│       ├── go.yml                    # Go test workflow
│       └── releaser.yaml             # Release workflow
├── pkg/                              # Public packages
│   └── build/
│       └── build.go                  # Build information utilities
└── kademlia/                         # Core Kademlia implementation
    ├── go.mod                        # Kademlia module definition
    ├── node.go                       # Main Node implementation with storage
    ├── node_test.go                  # Node functionality tests
    ├── routingtable.go               # Kademlia routing table (k-buckets)
    ├── routingtable_test.go          # Routing table tests
    ├── contact.go                    # Network contact management
    ├── contact_test.go               # Contact management tests
    ├── message.go                    # Protocol message definitions
    ├── message_handler.go            # Message processing logic
    ├── message_handler_test.go       # Message handler tests
    ├── transport.go                  # UDP network transport
    ├── kademliaid.go                 # 160-bit identifier implementation
    ├── kademliaid_test.go            # KademliaID tests
    ├── bucket.go                     # K-bucket implementation
    ├── config.go                     # Network configuration
    ├── config_test.go                # Configuration tests
    ├── interfaces.go                 # Interface definitions
    ├── join_service.go               # Network joining logic
    ├── helpers.go                    # Utility functions
    ├── helpers_test.go               # Helper function tests
    ├── storage_test.go               # M2 storage functionality tests
    └── benchmark_network_test.go     # Network performance benchmarks
```

## Requirements

- **Go 1.23.5 or later**
- **Docker** (for containerized deployment)
- **Docker Compose** (for multi-node networks)

## Quick Start

### 1. Build the Project

```bash
# Navigate to project directory
cd kademlia

# Download dependencies and build
go mod tidy
go build -o kademlia-node
```

### 2. Run a Single Node

```bash
# Start a bootstrap node
./kademlia-node -port=8001

# In another terminal, start a second node that joins the first
./kademlia-node -port=8002 -target=127.0.0.1:8001
```

### 3. Available Command Line Options

```bash
./kademlia-node -h
```

Options:
- `-port=8001` - Set the UDP port for this node
- `-target=127.0.0.1:8001` - Bootstrap node address to join
- `-id=<hex_string>` - Use specific node ID (optional)

## Testing

### Run All Tests

```bash
# Navigate to kademlia package
cd kademlia

# Run complete test suite
go test -v

# Run tests with race detection
go test -v -race
```

### Run Specific Test Categories

```bash
# Test storage functionality (M2)
go test -v -run "TestStore|TestFindValue|TestConcurrent"

# Test network functionality (M1)
go test -v -run "TestPing|TestFindNode|TestNetwork"

# Test routing table
go test -v -run "TestRoutingTable"

# Test message handling
go test -v -run "TestMessage"
```

### Test Output Example

```console
=== RUN   TestNodeStoreAndRetrieve
DEBUG: Stored value for key a3bb6783fd4cf3a7274f5a5d623e6353e19f031e locally
--- PASS: TestNodeStoreAndRetrieve (0.00s)
=== RUN   TestStoreMessageHandler
DEBUG: Processing STORE request
DEBUG: Storing key=test_key, value=test_value
--- PASS: TestStoreMessageHandler (0.00s)
=== RUN   TestFindValueMessageHandler
DEBUG: Processing FIND_VALUE request
DEBUG: Found value locally for key find_test_key
--- PASS: TestFindValueMessageHandler (0.00s)
PASS
ok      kademlia        0.009s
```

## Docker Deployment

### Single Container Test

```bash
# Build Docker image
docker build -t kademlia-node .

# Run single node
docker run -p 8001:8001 kademlia-node ./kademlia-node -port=8001
```

### 50-Node Network (M5 Requirement)

```bash
# Start the complete 50-node network
docker compose up

# Start in background
docker compose up -d

# View logs
docker compose logs -f

# Stop the network
docker compose down
```

### Docker Network Configuration

The `docker-compose.yml` configures:
- **50 nodes** (node1 through node50)
- **Port range:** 8001-8050
- **Bootstrap node:** node1 (all others connect to it)
- **Automatic container dependencies**

## API Usage

### Core Node Operations

```go
import "kademlia"

// Create a new node
node := kademlia.NewNode("127.0.0.1:8001")

// Store a value
node.StoreValue("my-key", "my-value")

// Retrieve a value
value, found := node.GetValue("my-key")
if found {
    fmt.Printf("Retrieved: %s\n", value)
}

// Send store message to network
node.SendStore("network-key", "network-value")

// Find value in network
result := node.SendFindValue("network-key")
```

### Message Types

The implementation supports these Kademlia protocol messages:
- **PING/PONG** - Liveness check
- **FIND_NODE** - Locate nodes closest to a target ID
- **STORE** - Store key-value pairs
- **FIND_VALUE** - Retrieve values or find closest nodes

## Development

### Adding New Features

1. **Implement the feature** in the appropriate `.go` file
2. **Add tests** in corresponding `*_test.go` file
3. **Run tests** to ensure functionality
4. **Update this README** if needed

### Code Quality

```bash
# Format code
go fmt ./...

# Vet code for issues
go vet ./...

# Run tests with coverage
go test -cover -v ./...
```

### Debugging

Enable debug output by looking for `DEBUG:` prefixed log messages in test output. The implementation includes extensive debugging information for:
- Message processing
- Storage operations
- Network events
- Routing table updates

## Performance

### Concurrent Operations

The implementation supports concurrent operations with:
- **Thread-safe storage** using `sync.RWMutex`
- **Concurrent network operations**
- **Race condition protection**

Test concurrent functionality:
```bash
go test -v -run TestConcurrentStorage
```

### Network Scale

- **Supports up to 50 nodes** in Docker deployment
- **Optimized for local testing** and development
- **UDP-based networking** for low latency
- **JSON message serialization** for readability

## Troubleshooting

### Common Issues

1. **Port already in use**
   ```bash
   # Check for running processes
   lsof -i :8001
   # Kill if necessary
   pkill -f kademlia-node
   ```

2. **Module dependency errors**
   ```bash
   # Clean and rebuild
   go mod tidy
   go clean -cache
   go build
   ```

3. **Docker issues**
   ```bash
   # Rebuild containers
   docker compose down
   docker compose build --no-cache
   docker compose up
   ```

### Verification Commands

```bash
# Verify build
go build -o kademlia-node && echo "Build successful"

# Verify tests
go test ./kademlia && echo "All tests passed"

# Verify Docker
docker compose config --quiet && echo "Docker config valid"
```

## Architecture

### Protocol Implementation

This Kademlia implementation follows the original paper specifications:
- **160-bit node identifiers** using SHA-1
- **K-bucket routing tables** with k=20
- **XOR distance metric** for node proximity
- **Iterative lookup algorithms**
- **UDP transport protocol**

### Thread Safety

All storage operations are protected by:
- `sync.RWMutex` for read/write operations
- Atomic operations where appropriate
- Goroutine-safe message handling

## Contributing

1. Fork the repository
2. Create a feature branch
3. Add tests for new functionality
4. Ensure all tests pass: `go test -v ./...`
5. Submit a pull request

## License

This project is part of an academic implementation of the Kademlia distributed hash table protocol.
