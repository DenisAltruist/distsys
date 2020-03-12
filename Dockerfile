FROM golang:latest

EXPOSE 12345

WORKDIR /go/src/github.com/DenisAltruist/distsys
COPY . .

RUN go get -v .

CMD go run main.go