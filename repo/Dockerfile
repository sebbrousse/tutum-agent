FROM ubuntu:trusty

RUN apt-get update && apt-get install -yq reprepro awscli createrepo
ADD distributions /repo/conf/distributions
ADD . /

ENV AWS_ACCESS_KEY_ID changeme
ENV AWS_SECRET_ACCESS_KEY changeme
ENV S3_BUCKET repo.tutum.co

CMD ["/run.sh"]
