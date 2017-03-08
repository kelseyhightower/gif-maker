gcloud container clusters get-credentials k0 \
  --zone us-west1-a
C1_SERVER=$(gcloud container clusters describe k0 \
  --format 'value(endpoint)')
C1_CERTIFICATE_AUTHORITY_DATA=$(gcloud container clusters describe k0 \
  --format 'value(masterAuth.clusterCaCertificate)')
C1_CLIENT_CERTIFICATE_DATA=$(gcloud container clusters describe k0 \
  --format 'value(masterAuth.clientCertificate)')
C1_CLIENT_KEY_DATA=$(gcloud container clusters describe k0 \
  --format 'value(masterAuth.clientKey)')
kubectl config set-cluster gke --kubeconfig k0-kubeconfig
kubectl config set clusters.gke.server \
  "https://${C1_SERVER}" \
  --kubeconfig k0-kubeconfig
kubectl config set clusters.gke.certificate-authority-data \
  ${C1_CERTIFICATE_AUTHORITY_DATA} \
  --kubeconfig k0-kubeconfig
kubectl config set-credentials cloudbuilder --kubeconfig k0-kubeconfig
kubectl config set users.cloudbuilder.client-certificate-data \
  ${C1_CLIENT_CERTIFICATE_DATA} \
  --kubeconfig k0-kubeconfig
kubectl config set users.cloudbuilder.client-key-data \
  ${C1_CLIENT_KEY_DATA} \
  --kubeconfig k0-kubeconfig
kubectl config set-context cloudbuilder \
  --cluster=gke \
  --user=cloudbuilder \
  --kubeconfig k0-kubeconfig
kubectl config use-context cloudbuilder \
  --kubeconfig k0-kubeconfig
