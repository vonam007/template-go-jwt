### multi-stage Dockerfile for building and running the Go app
FROM golang:1.20-alpine AS builder
WORKDIR /src

# Install necessary packages for building
RUN apk add --no-cache git ca-certificates tzdata build-base

# Copy go.mod and go.sum to leverage module cache
COPY go.mod go.sum ./
RUN go mod download

# If you vendor dependencies (recommended for faster, reproducible builds),
# copy the vendor directory into the image so build can use it.
COPY vendor ./vendor

# Copy the rest of the source
COPY . .

# Build the static binary using vendor if available
# If vendor/ exists, use -mod=vendor to avoid network fetches.
RUN if [ -d ./vendor ]; then \
			CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -mod=vendor -o /app/app ./; \
		else \
			CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /app/app ./; \
		fi

### final image
FROM alpine:3.18
RUN apk add --no-cache ca-certificates tzdata
WORKDIR /app
COPY --from=builder /app/app /app/app

ENV PORT=8080
EXPOSE 8080

ENTRYPOINT ["/app/app"]
