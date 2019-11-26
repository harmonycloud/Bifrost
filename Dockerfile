FROM amd64/alpine:3.8

COPY ./scripts/install-cni.sh /
COPY ./scripts/hcbridge.config.default /
COPY ./hcipam /
COPY ./hcbridge /