# Kademlia DHT Implementation

A distributed hash table (DHT) implementation based on the Kademlia protocol, written in Go. This implementation provides peer-to-peer networking, object storage and retrieval, and supports large-scale distributed networks.

## Features

- **M1: Network Formation** ✅ - Bootstrap and join Kademlia networks
- **M2: Object Distribution** ✅ - Store and retrieve immutable string data across the network
- **M3: Command Line Interface** ✅ - Interactive CLI with put, get, and exit commands
- **M4: Network Emulation & Testing** ✅ - Large-scale network testing (1000+ nodes) with packet dropping simulation
- **M5: Containerization** ✅ - Docker support for 50-node networks
- **M7: Thread Safety** ✅ - Concurrent operations with proper synchronization

## Project Delimitations

This implementation follows specific delimitations to maintain simplicity:

- **Immutable Data**: Objects cannot be deleted or modified once stored
- **Hash-Only Access**: Data objects are accessed only by their SHA-1 hash (40-character hex)
- **UTF-8 Strings**: All data objects are UTF-8 strings (no binary data)
- **In-Memory Storage**: No disk persistence - data is lost when nodes stop
- **No Encryption**: All network communication is unencrypted JSON over UDP
- **No Permissions**: All nodes can access all stored data objects
- **Simplified Routing**: Uses flat k-bucket structure (k=20, 160 buckets)
- **Simplified Lookups**: Basic FIND_NODE operations without full iterative/recursive algorithms

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
    ├── m4_network_emulation_test.go  # M4 large-scale network emulation tests
    └── benchmark_network_test_test.go # Network performance benchmarks
```

## Requirements

- **Go 1.23.5 or later**
- **Docker** (for containerized deployment)
- **Docker Compose** (for multi-node networks)
- **GCC** (optional, for race detection testing - `sudo apt install build-essential` on Ubuntu/Debian)

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

# The node will start and display a CLI prompt
Node is running. Type 'help' for available commands.
> 
```

### 3. Interactive CLI Commands

Once the node is running, you can use these commands:

```bash
# Store immutable string content in the DHT
> put hello world
Storing content: hello world
Content stored successfully!
Hash: 2aae6c35c94fcfb415dbe95f408b9ce91ee846ed

# Retrieve content by its SHA-1 hash (exactly 40 hex characters)
> get 2aae6c35c94fcfb415dbe95f408b9ce91ee846ed
Looking for content with hash: 2aae6c35c94fcfb415dbe95f408b9ce91ee846ed
Content found locally!
Content: hello world
Retrieved from: local node (0.0.0.0:8001)

# View available commands
> help
Available commands:
  put <content>  - Store UTF-8 string content and return its SHA-1 hash
  get <hash>     - Retrieve content by 40-character hex hash
  exit           - Shutdown the node and exit
  help           - Show this help message

# Exit the node
> exit
Shutting down node...
Goodbye!
```

### 4. Join an Existing Network

```bash
# In another terminal, start a second node that joins the first
./kademlia-node -port=8002 -target=127.0.0.1:8001

# Now you can store content on one node and retrieve it from another
```

### 5. Daemon Mode (for Containers)

```bash
# Run node in daemon mode (no CLI, for containers)
./kademlia-node -port=8001 -daemon

# Join existing network in daemon mode
./kademlia-node -port=8002 -target=127.0.0.1:8001 -daemon
```

### 6. Available Command Line Options

```bash
./kademlia-node -h
```

Options:
- `-port=8001` - Set the UDP port for this node
- `-target=127.0.0.1:8001` - Bootstrap node address to join
- `-id=<hex_string>` - Use specific node ID (optional)
- `-daemon` - Run in daemon mode (no interactive CLI, for containers)

### CLI Commands

Once the node is running, the following interactive commands are available:

- **`put <content>`** - Store immutable UTF-8 string content and return its SHA-1 hash
- **`get <hash>`** - Retrieve content by exact 40-character hex hash from the DHT  
- **`exit`** - Shutdown the node gracefully
- **`help`** - Display available commands

**Note**: Data objects are immutable (cannot be modified or deleted) and are accessed only by their SHA-1 hash.

## Testing

### Run All Tests

```bash
# Navigate to kademlia package
cd kademlia

# Run complete test suite (recommended for most users)
go test -v

# Run tests with race detection (requires GCC compiler)
# Install GCC first: sudo apt install build-essential
CGO_ENABLED=1 go test -v -race

# Alternative: Run tests without race detection if GCC unavailable
go test -v
```

**Note for VS Code Users:** If VS Code test runner shows "setup failed" errors with module path `main/kademlia`, run tests from terminal instead. The issue is with module path resolution in the IDE.

### Run Specific Test Categories

```bash
# Test storage functionality (M2)
go test -v -run "TestStore|TestFindValue|TestConcurrent"

# Test network functionality (M1)
go test -v -run "TestPing|TestFindNode|TestNetwork"

# Test large-scale network emulation (M4)
go test -v -run "TestLargeScale|TestPacketDropping|TestConfigurable|TestNetworkResilience"

# Test routing table
go test -v -run "TestRoutingTable"

# Test message handling
go test -v -run "TestMessage"

# Test KademliaID functionality
go test -v -run "TestKademliaID"

# Test contact management
go test -v -run "TestContact"
```

**VS Code Testing:** If using VS Code test runner and encountering module path issues, use the terminal commands above instead. All tests will run correctly from the `kademlia/` directory.

### Test CLI Commands

```bash
# Build and test CLI functionality
cd kademlia
go build -o kademlia-node

# Test with scripted input (basic commands)
echo -e "help\nput test content\nexit" | ./kademlia-node -port=8001

# Test with a complete put/get cycle (use hash returned by PUT)
echo -e "put hello world\nget <USE_HASH_FROM_PUT_OUTPUT>\nexit" | ./kademlia-node -port=8001

# Create a test script for repeated testing
cat > test_cli.sh << 'EOF'
#!/bin/bash
echo -e "help\nput my immutable data\nput another immutable value\nexit" | ./kademlia-node -port=8001
EOF
chmod +x test_cli.sh
./test_cli.sh
```

**Note:** The hash `2aae6c35c94fcfb415dbe95f408b9ce91ee846ed` is the SHA-1 hash of "hello world". In real usage, use the hash returned by your `put` command.

### M4 Network Emulation Testing

The implementation includes comprehensive M4 testing for large-scale network emulation:

```bash
# Run all M4 network emulation tests
go test -v -run "TestLargeScale|TestPacketDropping|TestConfigurable|TestNetworkResilience"

# Test large-scale network (1000+ nodes)
go test -v -run "TestLargeScaleNetworkEmulation"

# Test packet dropping functionality
go test -v -run "TestPacketDroppingFunctionality"

# Test configurable parameters
go test -v -run "TestConfigurableParameters"

# Test network resilience under adverse conditions
go test -v -run "TestNetworkResilience"

# Run M4 benchmarks
go test -bench="BenchmarkNetworkEmulation" -v
```

#### M4 Testing Features

- **Large-Scale Emulation**: Tests with 1000+ nodes to validate scalability
- **Packet Dropping Simulation**: Configurable packet loss (0-80%) to test network resilience
- **Configurable Parameters**: Easy-to-modify constants for testing different scenarios
- **Network Resilience**: Validates DHT operations under adverse network conditions
- **Concurrent Operations**: Tests thread safety under packet loss
- **Performance Benchmarking**: Measures network setup and operation performance

#### M4 Configuration Constants

The M4 tests use easily configurable constants:

```go
const (
    TestNodeCount         = 1000 // Large-scale test (1000+ nodes)
    TestNodeCountSmall    = 100  // Smaller tests for faster execution
    DefaultPacketDropRate = 0.1  // 10% packet loss
    HighPacketDropRate    = 0.3  // 30% packet loss for stress testing
    NoPacketDrop          = 0.0  // No packet loss for baseline
)
```

### Test Output Example

```console
=== RUN   TestNodeStoreAndRetrieve
DEBUG: Stored value for key a3bb6783fd4cf3a7274f5a5d623e6353e19f031e locally
--- PASS: TestNodeStoreAndRetrieve (0.00s)
=== RUN   TestLargeScaleNetworkEmulation
--- PASS: TestLargeScaleNetworkEmulation (3.78s)
=== RUN   TestPacketDroppingFunctionality
--- PASS: TestPacketDroppingFunctionality (1.42s)
=== RUN   TestNetworkResilience
--- PASS: TestNetworkResilience (0.13s)
PASS
ok      kademlia        6.803s
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

# Start in background (daemon mode)
docker compose up -d

# Start specific nodes for testing
docker compose up -d node1 node2 node3

# View logs from specific node
docker compose logs -f node1

# Stop the network
docker compose down
```

### Docker Network Configuration

The `docker-compose.yml` configures:
- **50 nodes** (node1 through node50)
- **Port range:** 8001-8050
- **Bootstrap node:** node1 (all others connect to it)
- **Automatic container dependencies**
- **Daemon mode:** All containers run in non-interactive mode

Each node runs with the `-daemon` flag to avoid CLI conflicts in containerized environments.

## API Usage

### Command Line Interface (M3)

The node provides an interactive CLI for storing and retrieving content:

```bash
# Start the node in interactive mode
./kademlia-node -port=8001

# Interactive commands:
> put "my important data"     # Store content, returns hash
> get <hash>                  # Retrieve content by hash
> help                        # Show available commands
> exit                        # Shutdown node
```

### Connecting CLI to Docker Network

You can run an interactive CLI that connects to a Docker network:

```bash
# Start Docker nodes in daemon mode
docker compose up -d node1 node2 node3

# Connect interactive CLI to the Docker network
./kademlia-node -port=8051 -target=127.0.0.1:8001

# Now use CLI commands to interact with the distributed network:
> put hello world distributed
> get <returned_hash>
```

### Core Node Operations

```go
import "kademlia"

// Create a new node
node := kademlia.NewNode("127.0.0.1:8001", "", kademlia.DefaultConfig())

// Store immutable UTF-8 string data (generates SHA-1 hash)
hash, err := node.SendStore("my immutable data")
if err == nil {
    fmt.Printf("Stored with hash: %s\n", hash)
}

// Retrieve by exact hash only
value, found, err := node.SendFindValue(hash)
if found {
    fmt.Printf("Retrieved: %s\n", value)
}

// Local storage (hash-based)
node.StoreValue(hash, "my-value")
value, exists := node.GetValue(hash)
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

#test coverage precentage
go test -cover -coverprofile=coverage.out
go tool cover -func=coverage.out

# Run race detection (if CGO available)
CGO_ENABLED=1 go test -race -v ./...
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
- **M4 Network Emulation**: Tests with 1000+ nodes for scalability validation
- **Packet Loss Simulation**: Configurable packet dropping (0-80%) for resilience testing
- **Optimized for local testing** and development
- **UDP-based networking** for low latency
- **JSON message serialization** for readability

#### M4 Performance Metrics

- **Large Network Creation**: ~1-2 seconds for 1000 nodes
- **Network Bootstrap**: ~1-2 seconds for full connectivity
- **Packet Drop Simulation**: Accurate loss rates with minimal overhead
- **Concurrent Operations**: Thread-safe operations under packet loss
- **Memory Efficiency**: Optimized allocation for large-scale testing

## Troubleshooting

### Common Issues

1. **Port already in use**
   ```bash
   # Check for running processes
   lsof -i :8001
   # Kill if necessary
   pkill -f kademlia-node
   ```

2. **CLI not responding**
   ```bash
   # Ensure you press Enter after typing commands
   # Use Ctrl+C to force shutdown if needed
   # For daemon mode, use: pkill -f kademlia-node
   ```

3. **Content not found with GET**
   ```bash
   # Ensure the hash is exactly 40 hex characters (SHA-1)
   # Example: 2aae6c35c94fcfb415dbe95f408b9ce91ee846ed
   # Hash is case-sensitive
   # Content may be lost if storing node was terminated
   # Try storing content first with PUT
   ```

4. **Module dependency errors**
   ```bash
   # Clean and rebuild
   go mod tidy
   go clean -cache
   go build
   ```

5. **Docker issues**
   ```bash
   # Rebuild containers
   docker compose down
   docker compose build --no-cache
   docker compose up
   
   # Check container status
   docker ps
   docker compose logs node1
   ```

6. **Hash validation errors**
   ```bash
   # SHA-1 hash must be exactly 40 hex characters
   # Example valid: 2aae6c35c94fcfb415dbe95f408b9ce91ee846ed
   # Invalid: short hashes, non-hex characters, uppercase/lowercase mix
   # Hashes are generated automatically by PUT command
   ```

7. **Race detection requires CGO and GCC**
   ```bash
   # If you get "go: -race requires cgo; enable cgo by setting CGO_ENABLED=1"
   CGO_ENABLED=1 go test -v -race
   
   # If you get 'cgo: C compiler "gcc" not found'
   # Install GCC first:
   sudo apt install build-essential  # Ubuntu/Debian
   # or
   sudo yum install gcc              # CentOS/RHEL
   # or
   brew install gcc                  # macOS
   
   # Then run with CGO enabled:
   CGO_ENABLED=1 go test -v -race
   
   # Or simply run tests without race detection:
   go test -v
   ```

8. **VS Code test runner issues**
   ```bash
   # If VS Code shows "FAIL main/kademlia [setup failed]"
   # This is a module path issue - use terminal instead:
   cd kademlia
   go test -v
   
   # Or run specific tests:
   go test -v -run "TestNodeStore"
   ```

### Verification Commands

```bash
# Verify build
go build -o kademlia-node && echo "Build successful"

# Verify tests (recommended)
go test ./kademlia && echo "All tests passed"

# Verify Docker config
docker compose config --quiet && echo "Docker config valid"

# Test CLI functionality with immutable data
echo -e "help\nput immutable test content\nexit" | ./kademlia-node -port=8001

# Test complete put/get cycle (note: use actual hash from PUT output)
echo -e "put hello world\nget <USE_ACTUAL_HASH>\nexit" | ./kademlia-node -port=8001

# Test daemon mode
timeout 5s ./kademlia-node -port=8001 -daemon && echo "Daemon mode works"

# Test with race detection (only if GCC is installed)
command -v gcc >/dev/null 2>&1 && CGO_ENABLED=1 go test -race ./kademlia || echo "Skipped race detection (GCC not available)"
```

## Architecture

### Protocol Implementation

This simplified Kademlia implementation includes:
- **160-bit node identifiers** using SHA-1
- **K-bucket routing tables** with k=20 (simplified flat structure)
- **XOR distance metric** for node proximity
- **Basic lookup operations** (simplified single-node queries, not full iterative algorithm)
- **UDP transport protocol** with JSON messages
- **Immutable string storage** with hash-only access

**Intentional Simplifications** (per course delimitations):
- No recursive/iterative lookup algorithms
- No parallel alpha queries to multiple nodes
- No response aggregation or state management
- Single-node FIND_NODE operations instead of distributed lookup

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
