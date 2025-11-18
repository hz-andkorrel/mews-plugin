FROM golang:1.23-alpine

WORKDIR /app

# Copy go mod files
COPY go.* ./

# Download dependencies and create go.sum
RUN go mod download && go mod verify

# Copy source code
COPY main.go ./

# Build the application
RUN go build -o mews-plugin main.go

# Run the binary
CMD ["./mews-plugin"]
