FROM golang:1.22-alpine

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY . .
RUN mkdir bin
RUN go build -v -o ./bin ./cmd/scheduler ./cmd/worker
