
FROM golang:1.25-alpine AS builder

RUN apk add --no-cache git

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .


RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo -o /app/main .


FROM alpine:3.19

WORKDIR /app

RUN apk --no-cache add ca-certificates


COPY --from=builder /app/main .

COPY --from=builder /app/dist ./dist

EXPOSE 8080

# Запускаем
CMD ["./main"]