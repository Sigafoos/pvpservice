FROM golang:1.13 AS build

WORKDIR /app
COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=arm GOARM=7 go build -a -ldflags '-extldflags "-static"'

FROM alpine:latest AS certs
RUN apk --no-cache add ca-certificates

FROM scratch
COPY --from=build /app/pvpservice .
COPY --from=certs /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY migrations migrations
ENTRYPOINT ["./pvpservice"]
