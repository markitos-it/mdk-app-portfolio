FROM golang:1.26-alpine AS builder

WORKDIR /build


ENV CGO_ENABLED=0
ENV GOOS=linux
ENV GO111MODULE=on

RUN apk add --no-cache make git bash

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN echo "Building service at path: cmd/app/main.go"
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o app cmd/app/main.go

FROM alpine:3.22

RUN apk --no-cache add ca-certificates \
 && addgroup -S app \
 && adduser -S -G app app

WORKDIR /app

COPY --from=builder /build/app .
COPY --from=builder /build/pages .
COPY --from=builder /build/templates .
COPY --from=builder /build/assets .

HEALTHCHECK --interval=30s --timeout=5s --start-period=20s --retries=3 CMD pidof app >/dev/null || exit 1

ARG SERVER_ADDRESS
RUN echo "SERVER_ADDRESS environment variable is set to: $SERVER_ADDRESS"
EXPOSE $SERVER_ADDRESS
USER app

ENV EVENT_RELAY_DATABASE_PATH=""
ENV SERVER_ADDRESS=$SERVER_ADDRESS

CMD ["./app"]