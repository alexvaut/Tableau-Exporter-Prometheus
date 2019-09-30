FROM golang:1.13-alpine AS builder

# Add Maintainer Info
LABEL maintainer="Alexandre Vautier <alex@vautier.biz>"

# Set the Current Working Directory inside the container
WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download all dependencies. Dependencies will be cached if the go.mod and go.sum files are not changed
RUN go mod download

# Copy the source from the current directory to the Working Directory inside the container
COPY . .

# Build the Go app
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo -o main .

# Expose port 9030 to the outside world
EXPOSE 9030

# Command to run the executable
CMD ["./main"]

FROM scratch
COPY --from=builder /app/main /app/main
COPY --from=builder /app/config.yml /app/config.yml
WORKDIR /app
ENTRYPOINT ["./main"]