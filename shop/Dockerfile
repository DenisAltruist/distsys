FROM golang:latest

EXPOSE 12345

WORKDIR /go/src/github.com/DenisAltruist/distsys/shop
COPY . .

# Would download external packages from remote github.com site
RUN go get -v . 

CMD go run ./main.go