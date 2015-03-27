default: all

all: image
	mkdir -p ./build
	docker run --name=agentbuild tutum-agent contrib/make-all.sh
	docker cp agentbuild:/build .
	docker rm -f agentbuild

clean:
	rm -fr build/
	docker rm -f agentbuild || true
	docker rmi tutum-agent || true

image:
	docker build --force-rm --rm -t tutum-agent .

upload:
	s3cmd sync -P build s3://files.tutum.co/tutum-agent/