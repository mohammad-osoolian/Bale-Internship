FROM golang:1.22-alpine

WORKDIR /fileserver

COPY go.mod go.sum ./
RUN go mod download

COPY . .
WORKDIR /fileserver/main
EXPOSE 8000
CMD ["go", "run", "main.go"]