# Use Go image directly (larger but simpler)
FROM golang:1.23.5

WORKDIR /app

# Copy source code
COPY . .

# Download dependencies
RUN go mod download

# Build the application
RUN go build -o main .

# Make sure binary is executable
RUN chmod +x ./main

# Verify the binary exists
RUN ls -la ./main

# Run the application
CMD ["./main"]