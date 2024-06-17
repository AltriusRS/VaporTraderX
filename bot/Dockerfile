# Build stage
FROM golang:1.22.4-alpine3.20 AS build

WORKDIR /app

COPY go.mod .
COPY go.sum .
COPY ./src ./src


RUN go mod download

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o vaportrader ./src

# Final stage
FROM alpine:3.17

RUN apk add --no-cache ca-certificates

RUN addgroup -S vaportrader && adduser -S vaportrader -G vaportrader

USER vaportrader

COPY --from=build /app/vaportrader /vaportrader

ENTRYPOINT ["/vaportrader"]