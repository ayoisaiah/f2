# Build stage
ARG GO_VERSION=1.26.2
FROM golang:${GO_VERSION}-alpine AS builder

# Set the working directory
WORKDIR /build

# Install build dependencies
RUN apk add --no-cache git

# Copy and download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy the source code
COPY . .

# Build the binary with optimizations
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -trimpath -o /usr/bin/f2 ./cmd/f2

# Final stage
FROM alpine:3.21 AS final

# Metadata arguments
ARG REPO_OWNER=ayoisaiah
ARG REPO_BINARY_NAME=f2
ARG REPO_DESCRIPTION="F2 is a cross-platform command-line tool for batch renaming files and directories quickly and safely"

# Metadata labels
LABEL org.opencontainers.image.source="https://github.com/${REPO_OWNER}/${REPO_BINARY_NAME}"
LABEL org.opencontainers.image.description="${REPO_DESCRIPTION}"
LABEL org.opencontainers.image.licenses="MIT"

# Install runtime dependencies (exiftool)
RUN apk add --no-cache exiftool

# Set the working directory
WORKDIR /app

# Copy the binary from the builder
COPY --from=builder /usr/bin/f2 /usr/bin/f2

# Set the entrypoint
ENTRYPOINT ["f2"]
