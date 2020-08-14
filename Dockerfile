FROM amd64/alpine:3.8

COPY ./scripts/install-cni.sh /
COPY ./scripts/hcmacvlan.config.default /
COPY ./hcipam /
COPY ./hcmacvlan /
COPY ./hcipvlan /