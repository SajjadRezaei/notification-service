FROM golang:1.23 AS builder


WORKDIR /app

COPY go.* ./
RUN go mod download

COPY . ./

# Build the application
RUN go build -v -o server ./src/cmd/server/main.go




FROM debian:bookworm

COPY --from=builder /app/server /app/server

CMD ["./app/server"]