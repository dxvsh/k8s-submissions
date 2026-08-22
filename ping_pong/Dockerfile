# Build image
FROM golang:1.22-alpine AS build

WORKDIR /usr/src/app

COPY go.mod go.sum ./
RUN go mod download

COPY main.go ./
RUN CGO_ENABLED=0 go build -o /usr/src/app/pong

# Copy binary to a lean runtime image
FROM alpine:3.14

WORKDIR /usr/src/app

COPY --from=build /usr/src/app/pong ./pong

ENTRYPOINT ["./pong"]
