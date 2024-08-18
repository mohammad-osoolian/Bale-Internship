FROM golang:1.22-alpine

WORKDIR /message-broker

COPY go.mod go.sum ./
RUN go mod download

COPY . .

EXPOSE 8000
CMD ["go", "run", "main.go"]