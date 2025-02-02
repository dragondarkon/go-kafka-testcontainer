# FROM golang:1.13.6-alpine as app

# WORKDIR /app

# COPY . /app

# RUN apk update && apk add git


# RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build 

# FROM alpine
# COPY --from=app /app/go-kafka-testcontainer .
# # COPY --from=app /app/config.yml .
# EXPOSE 8080
# ENTRYPOINT [ "./go-kafka-testcontainer" ]


# Start from golang base image
FROM golang:alpine as builder

ENV GO111MODULE=on

# Install git.
# Git is required for fetching the dependencies.
RUN apk update && apk add --no-cache git

# Set the current working directory inside the container 
WORKDIR /app

# Copy go mod and sum files 
COPY go.mod go.sum ./

# Download all dependencies. Dependencies will be cached if the go.mod and the go.sum files are not changed 
RUN go mod download 

# Copy the source from the current directory to the working Directory inside the container 
COPY . .
RUN ls -lrt
# Build the Go app
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o go-kafka-testcontainer .

# Start a new stage from scratch
FROM alpine:latest
RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copy the Pre-built binary file from the previous stage. Observe we also copied the .env file
COPY --from=builder /app/go-kafka-testcontainer .
COPY --from=builder /app/.env .       
RUN ls -lrt
# Expose port 8080 to the outside world
EXPOSE 8080

#Command to run the executable
CMD ["./go-kafka-testcontainer"]