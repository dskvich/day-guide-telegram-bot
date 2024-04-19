FROM golang:1.22-alpine as builder

RUN apk update && apk add --no-cache git

WORKDIR /app

COPY . ./
ENV CGO_ENABLED=0 GOOS=linux GOARCH=amd64
RUN go mod download
RUN go build -ldflags="-s -w" -o main main.go


FROM alpine

WORKDIR /app

COPY --from=builder /app/main ./

CMD ["./main"]