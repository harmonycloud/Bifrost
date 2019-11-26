#!/bin/bash

ETCD_URL=$1

bashpath=$(cd `dirname $0`; pwd)

cp -rf $bashpath/hccni.template.yaml $bashpath/hccni.yaml

#todo

etcd_https=""
master_names=`kubectl get nodes --selector=node-role.kubernetes.io/master | awk 'NR!=1 {print $1}'`
for node in $master_names; do
ip=`kubectl describe node $node  | grep InternalIP: | awk '{print $NF}'`
etcd_https=$etcd_https"https://"$ip":2379,"
done
etcd_https=${etcd_https%?}


etcd_client_key_base64=`cat /etc/kubernetes/pki/apiserver-etcd-client.key | base64 -w 0`
etcd_client_crt_base64=`cat /etc/kubernetes/pki/apiserver-etcd-client.crt | base64 -w 0`
etcd_crt_base64=`cat /etc/kubernetes/pki/etcd/ca.crt | base64 -w 0`


sed -i "s|.*etcd-key:.*|  etcd-key: $etcd_client_key_base64|" $bashpath/hccni.yaml
sed -i "s|.*etcd-cert:.*|  etcd-cert: $etcd_client_crt_base64|" $bashpath/hccni.yaml
sed -i "s|.*etcd-ca:.*|  etcd-ca: $etcd_crt_base64|" $bashpath/hccni.yaml

#sed -i "s|.*etcd_endpoints:.*|  etcd_endpoints: \"${etcd_https}\"|" $bashpath/hccni.yaml
sed -i "s|.*etcd_endpoints:.*|  etcd_endpoints: \"${ETCD_URL}\"|" $bashpath/hccni.yaml

/usr/bin/kubectl create -f $bashpath/hccni.yaml

