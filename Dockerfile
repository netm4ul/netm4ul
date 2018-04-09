FROM golang:1.9

ARG PROJECT="github.com/netm4ul/netm4ul"
ARG FULL_PATH=/go/src/${PROJECT}
ARG EXECUTABLE=${FULL_PATH}/netm4ul
RUN useradd netm4ul

# RUN go get -u github.com/golang/dep/...
# RUN make
WORKDIR ${FULL_PATH}
RUN mkdir -p ${FULL_PATH}
COPY . ${FULL_PATH}
RUN mv netm4ul.conf.docker netm4ul.conf
RUN go build .

USER netm4ul
CMD [${EXECUTABLE}]
