FROM golang:1.4.2

# Install FPM for packaging
RUN apt-get update && DEBIAN_FRONTEND=noninteractive apt-get install -qy ruby ruby-dev && \
	gem install --no-rdoc --no-ri fpm --version 1.0.2

WORKDIR /usr/src/tutum-agent
ADD . /usr/src/tutum-agent
RUN go get -d -v && go build -v

CMD ["/usr/src/tutum-agent/tutum-agent"]
