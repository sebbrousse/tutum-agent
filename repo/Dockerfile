FROM ubuntu:trusty

RUN apt-get update && apt-get install -yq reprepro awscli
ADD distributions /repo/conf/distributions
ADD run.sh /run.sh
ADD package.deb /package.deb
ADD gpg_private_key /gpg_private_key
ADD gpg_public_key /gpg_public_key

ENV AWS_ACCESS_KEY_ID changeme
ENV AWS_SECRET_ACCESS_KEY changeme
ENV S3_BUCKET repo.tutum.co

CMD ["/run.sh"]
