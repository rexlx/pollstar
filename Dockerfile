FROM golang:1.21.0 as builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download
COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /app/main .

FROM alpine:latest
RUN apk --no-cache add ca-certificates
COPY --from=builder /app/main /app/main
COPY --from=builder /app/questions.json /questions.json
RUN chmod +x /app/main

EXPOSE 8080
CMD ["/app/main", "-questions", "/questions.json"]