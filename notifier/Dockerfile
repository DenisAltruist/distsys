FROM golang:1.13.8-alpine3.11 AS builder
RUN apk update && apk add git
WORKDIR /go/src/github.com/DenisAltruist/distsys
COPY utils utils
COPY db db
COPY notifier notifier
WORKDIR /go/src/github.com/DenisAltruist/distsys/notifier
RUN go get -d -v
RUN GOOS=linux go build -o /notifier

FROM alpine:3.11
# Adding ssl
RUN apk update && apk add ca-certificates
COPY --from=builder /notifier ./

CMD ["./notifier"]