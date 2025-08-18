FROM golang:1.25.0-alpine AS builder

WORKDIR /build

COPY go.mod go.sum ./

RUN go mod download

COPY . ./

RUN go build -o /usr/bin/f2 ./cmd/f2...

FROM alpine:3.22 AS final

RUN apk add --no-cache exiftool

WORKDIR /app

COPY --from=builder /usr/bin/f2 /usr/bin/f2

# Run the f2 command when the container starts
ENTRYPOINT ["f2"]
