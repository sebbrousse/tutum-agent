FROM ubuntu:trusty

RUN apt-get update && DEBIAN_FRONTEND=noninteractive apt-get install -qy automake build-essential curl git ruby ruby-dev

# Install Go
RUN curl -sSL https://go.googlecode.com/files/go1.2.1.src.tar.gz | tar -v -C /usr/local -xz
ENV PATH /usr/local/go/bin:$PATH
ENV GOROOT /usr/local/go
ENV GOPATH /go:/go/src/github.com/tutumcloud/tutum-agent

# Build go toolchain
RUN cd $GOROOT/src && GOOS=linux GOARCH=amd64 ./make.bash --no-clean 2>&1

# Install FPM for packaging
RUN gem install --no-rdoc --no-ri fpm --version 1.0.2

WORKDIR /go/src/github.com/tutumcloud/tutum-agent
ADD . /go/src/github.com/tutumcloud/tutum-agent
RUN go get
RUN go build

CMD ["/go/bin/tutum-agent"]
