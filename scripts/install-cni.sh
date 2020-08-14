#!/bin/sh

# Script to install hcmacvlan on a Kubernetes host.
# - Expects the host CNI binary path to be mounted at /host/opt/cni/bin.
# - Expects the host CNI network config path to be mounted at /host/etc/cni/net.d.

# Ensure all variables are defined, and that the script fails when an error is hit.
set -u -e

# Capture the usual signals and exit from the script
trap 'echo "SIGINT received, simply exiting..."; exit 0' SIGINT
trap 'echo "SIGTERM received, simply exiting..."; exit 0' SIGTERM
trap 'echo "SIGHUP received, simply exiting..."; exit 0' SIGHUP

# Helper function for raising errors
# Usage:
# some_command || exit_with_error "some_command_failed: maybe try..."
exit_with_error(){
  echo $1
  exit 1
}
INTERFACE_MASTER=$1

# Hcipam & hcmacvlan binary path
path=/
# Default config path
confpath=/
# Default cert path
keypath=/etcd-secrets
# Target CNI directory
dir=/host/opt/cni/bin
# Target config directory
confdir=/host/etc/cni/net.d
# Target ssl directory
ssldir=$confdir/ssl

# Clean up any existing binaries / config / assets.
rm -f $dir/hcipam
rm -f $dir/hcmacvlan
rm -rf $confdir/*

# Create ssl directory
mkdir $ssldir

# Copy key&cert
cp $keypath/* $ssldir || exit_with_error "Failed to copy $keypath to $ssldir. This may be caused by selinux configuration on the host, or something else."

# If the TLS assets actually exist, update the variables to populate into the
# CNI network config.  Otherwise, we'll just fill that in with blanks.
CNI_CONF_ETCD_CA=/etc/cni/net.d/ssl/etcd-ca
CNI_CONF_ETCD_KEY=/etc/cni/net.d/ssl/etcd-key
CNI_CONF_ETCD_CERT=/etc/cni/net.d/ssl/etcd-cert

# Load template config
TMP_CONF='/hcmacvlan.config.default'

# Load env args
# Use alternative command character "~", since these include a "/".
sed -i s~__ETCD_CERT_FILE__~${CNI_CONF_ETCD_CERT:-}~g $TMP_CONF
sed -i s~__ETCD_KEY_FILE__~${CNI_CONF_ETCD_KEY:-}~g $TMP_CONF
sed -i s~__ETCD_CA_CERT_FILE__~${CNI_CONF_ETCD_CA:-}~g $TMP_CONF
sed -i s~__ETCD_ENDPOINTS__~${ETCD_ENDPOINTS:-}~g $TMP_CONF
sed -i s~__INTERFACE_MASTER__~${INTERFACE_MASTER:-}~g $TMP_CONF

# Copy the binary and config
cp $path/hcipam $dir/hcipam || exit_with_error "Failed to copy $path to $dir. This may be caused by selinux configuration on the host, or something else."
cp $path/hcmacvlan $dir/hcmacvlan || exit_with_error "Failed to copy $path to $dir. This may be caused by selinux configuration on the host, or something else."
cp $TMP_CONF $confdir/10-my-net-ssl.conflist || exit_with_error "Failed to copy $confpath to $confdir. This may be caused by selinux configuration on the host, or something else."

chmod +x $dir/hcipam
chmod +x $dir/hcmacvlan

SERVICE_ACCOUNT_PATH=/var/run/secrets/kubernetes.io/serviceaccount
KUBE_CA_FILE=${KUBE_CA_FILE:-$SERVICE_ACCOUNT_PATH/ca.crt}
SKIP_TLS_VERIFY=${SKIP_TLS_VERIFY:-false}
# Pull out service account token.
SERVICEACCOUNT_TOKEN=$(cat $SERVICE_ACCOUNT_PATH/token)

# Check if we're running as a k8s pod.
if [ -f "$SERVICE_ACCOUNT_PATH/token" ]; then
  # We're running as a k8d pod - expect some variables.
  if [ -z ${KUBERNETES_SERVICE_HOST} ]; then
    echo "KUBERNETES_SERVICE_HOST not set"; exit 1;
  fi
  if [ -z ${KUBERNETES_SERVICE_PORT} ]; then
    echo "KUBERNETES_SERVICE_PORT not set"; exit 1;
  fi

  if [ "$SKIP_TLS_VERIFY" == "true" ]; then
    TLS_CFG="insecure-skip-tls-verify: true"
  elif [ -f "$KUBE_CA_FILE" ]; then
    TLS_CFG="certificate-authority-data: $(cat $KUBE_CA_FILE | base64 | tr -d '\n')"
  fi

  # Write a kubeconfig file for the CNI plugin.  Do this
  # to skip TLS verification for now.  We should eventually support
  # writing more complete kubeconfig files. This is only used
  # if the provided CNI network config references it.
  touch /host/etc/cni/net.d/hcmacvlan-kubeconfig
  chmod ${KUBECONFIG_MODE:-600} /host/etc/cni/net.d/hcmacvlan-kubeconfig
  cat > /host/etc/cni/net.d/hcmacvlan-kubeconfig <<EOF
# Kubeconfig file for hcmacvlan plugin.
apiVersion: v1
kind: Config
clusters:
- name: local
  cluster:
    server: ${KUBERNETES_SERVICE_PROTOCOL:-https}://[${KUBERNETES_SERVICE_HOST}]:${KUBERNETES_SERVICE_PORT}
    $TLS_CFG
users:
- name: hcmacvlan
  user:
    token: "${SERVICEACCOUNT_TOKEN}"
contexts:
- name: hcmacvlan-context
  context:
    cluster: local
    user: hcmacvlan
current-context: hcmacvlan-context
EOF

fi

while :
do
    sleep 10
done
