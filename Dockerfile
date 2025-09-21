FROM golang:alpine
WORKDIR /app
COPY . .
RUN go build -o kademlia-node
EXPOSE 8000
CMD ["./kademlia-node"]