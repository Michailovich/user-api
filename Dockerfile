FROM golang:1.23 AS builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

ENV GOOS=linux
ENV GOARCH=amd64
ENV CGO_ENABLED=0

RUN go build -o user-api .


FROM alpine:latest

RUN apk --no-cache add ca-certificates
COPY --from=builder /app/user-api /usr/local/bin/user-api
RUN chmod +x /usr/local/bin/user-api
RUN ls -l /usr/local/bin/
CMD ["user-api"]

EXPOSE 8080