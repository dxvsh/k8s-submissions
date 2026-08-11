# Build image
FROM golang:1.22-alpine AS build
WORKDIR /usr/src/
COPY . .
RUN CGO_ENABLED=0 go build -o /usr/src/app

# Copy binary to a lean distroless image
FROM gcr.io/distroless/static-debian13:nonroot
WORKDIR /usr/src
COPY --from=build /usr/src/ .
ENV PORT=8080
CMD ["./app"]
