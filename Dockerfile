FROM golang:alpine as builder
RUN apk update && apk add --no-cache git ca-certificates && update-ca-certificates
ARG COMMIT=unknown
ARG DATE=unknown
ARG VERSION=unknown
WORKDIR /app
ADD go.mod go.sum ./
RUN go mod download
RUN go mod verify
ADD . .
RUN go build -ldflags "-X main.Version=$(VERSION) -X main.Commit=$(COMMIT) -X main.Date=$(DATE)" \
    -o restmigrate \
    ./cmd/restmigrate

FROM scratch
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /app/restmigrate /usr/bin/restmigrate
ENTRYPOINT ["/usr/bin/restmigrate"]
