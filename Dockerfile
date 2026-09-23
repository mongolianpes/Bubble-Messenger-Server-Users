FROM golang:1.25.10-alpine AS builder

WORKDIR /users

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN GOOS=linux GOARCH=amd64 go build -o users ./cmd/users

FROM alpine:latest

WORKDIR /users

COPY --from=builder /users/users .

EXPOSE 8086

CMD ["./users"]