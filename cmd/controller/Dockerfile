FROM amd64/alpine:3.8
ADD hcipamRestServer /root/
RUN chmod +x /root/hcipamRestServer

ENTRYPOINT /root/hcipamRestServer  -cacert /etcd-secrets/etcd-ca -cert /etcd-secrets/etcd-cert  -key /etcd-secrets/etcd-key
