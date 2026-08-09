# Build image
FROM golang:1.22-alpine AS build
WORKDIR /usr/src/app
COPY . .
RUN CGO_ENABLED=0 go build -o /go/bin/app

# Copy binary to a lean distroless image
FROM gcr.io/distroless/static-debian13:nonroot
COPY --from=build /go/bin/app /
ENV PORT=8080
CMD ["/app"]
