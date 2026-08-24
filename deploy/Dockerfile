FROM golang:1.22-alpine AS build
WORKDIR /src
COPY . .
RUN go build -trimpath -ldflags='-s -w' -o /out/geofenced ./cmd/geofenced
FROM alpine:3.20
RUN adduser -D -u 10001 app
COPY --from=build /out/geofenced /usr/local/bin/geofenced
USER app
EXPOSE 8087
ENTRYPOINT ["/usr/local/bin/geofenced"]
