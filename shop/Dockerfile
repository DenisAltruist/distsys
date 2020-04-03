FROM golang:1.13.8-alpine3.11 AS builder
RUN apk update && apk add git
WORKDIR /go/src/github.com/DenisAltruist/distsys
COPY . .
RUN ls
WORKDIR /go/src/github.com/DenisAltruist/distsys/shop
RUN ls
RUN go get -d -v
RUN GOOS=linux go build -o /shop

FROM alpine:3.11
# Adding ssl
RUN apk update && apk add ca-certificates
COPY --from=builder /shop ./

CMD ["./shop"]