FROM golang:1.24.2 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY app/main.go ./
RUN GOOS=linux CGO_ENABLED=0 go build -o /app/main --trimpath --ldflags '-w -s' .


FROM gcr.io/distroless/static-debian12:latest

COPY --from=builder /app/main /app/main
EXPOSE 8080
ENTRYPOINT ["/app/main"]