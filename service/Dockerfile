## Stage to build backend go code
FROM golang:1.13 AS gobuilder

# Create and change to the app directory.
WORKDIR /app

# Retrieve application dependencies.
# This allows the container build to reuse cached dependencies.
COPY go.* ./
RUN go mod download

# Copy local code to the container image.
COPY . ./
# Build the app
RUN CGO_ENABLED=0 GOOS=linux go build -mod=readonly -v -o storage-events

FROM alpine:3
RUN apk add --no-cache ca-certificates

# Copy binary and assets
COPY --from=gobuilder /app/storage-events /storage-events

# Run
ENTRYPOINT ["./storage-events"]
