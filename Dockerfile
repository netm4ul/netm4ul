FROM golang:1.9

RUN mkdir /go/src/netm4ul

COPY . /go/src/netm4ul
WORKDIR /go/src/netm4ul

RUN go get -u github.com/golang/dep/...

RUN make

CMD ["/go/src/netm4ul/netm4ul"]
