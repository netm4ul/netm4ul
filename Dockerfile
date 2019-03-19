FROM debian:buster

RUN groupadd -r netm4ul && useradd -r -g netm4ul netm4ul

ARG PROJECT="github.com/netm4ul/netm4ul"
ARG FULL_PATH=/home/netm4ul/go/src/${PROJECT}
ARG EXECUTABLE=${FULL_PATH}/netm4ul

ARG GOVERSION=1.11.1
ARG GOOS=linux
ARG GOARCH=amd64
ENV PATH=$PATH:/usr/local/go/bin:/home/netm4ul/go/bin
ENV GOPATH=/home/netm4ul/go

# Install golang and general stuff (clang gcc git make)
RUN apt-get update && apt-get upgrade -y \
    && apt-get install -y wget clang gcc git make curl \ 
    && wget https://dl.google.com/go/go${GOVERSION}.${GOOS}-${GOARCH}.tar.gz \
    && tar -C /usr/local -xzf go${GOVERSION}.${GOOS}-${GOARCH}.tar.gz \
    && mkdir -p /home/netm4ul/go/bin
    # && curl https://raw.githubusercontent.com/golang/dep/master/install.sh | sh


# Install nmap : https://github.com/jessfraz/dockerfiles/blob/master/nmap/Dockerfile
RUN apt-get install -y \
    nmap \
    --no-install-recommends \
    && rm -rf /var/lib/apt/lists/*

# Install masscan : https://github.com/jessfraz/dockerfiles/blob/master/masscan/Dockerfile
RUN apt-get install -y \
    ca-certificates \
    libpcap0.8 \
    && rm -rf /var/lib/apt/lists/*

RUN git clone --depth 1 https://github.com/robertdavidgraham/masscan.git /usr/src/masscan \
    && cd /usr/src/masscan \
    && make \
    && make install \
    && rm -rf /usr/src/masscan

# RUN make
WORKDIR ${FULL_PATH}

RUN mkdir -p ${FULL_PATH} && mkdir /data && chown -R netm4ul:netm4ul /data
COPY . ${FULL_PATH}

RUN cd ${FULL_PATH} \
    && make
RUN mv netm4ul.conf.docker netm4ul.conf \
    && chown -R netm4ul:netm4ul ${FULL_PATH}

USER netm4ul
CMD [${EXECUTABLE}]
