#!/bin/bash
set -e

echo "Configuring Kubernetes tools..."

# Install kubectl
curl -LO "https://dl.k8s.io/release/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/linux/amd64/kubectl"
chmod +x kubectl
mv kubectl /usr/local/bin/

# Install helm
curl https://raw.githubusercontent.com/helm/helm/main/scripts/get-helm-3 | bash

# Install k9s (optional but useful)
curl -sS https://webinstall.dev/k9s | bash

# Create kubectl completion
kubectl completion bash > /etc/bash_completion.d/kubectl

# Create helm completion
helm completion bash > /etc/bash_completion.d/helm

# Create useful aliases
cat >> /etc/bash.bashrc << 'EOF'

# Kubernetes aliases
alias k=kubectl
alias kgp='kubectl get pods'
alias kgs='kubectl get services'
alias kgd='kubectl get deployments'
alias kgn='kubectl get nodes'
alias kdp='kubectl describe pod'
alias kds='kubectl describe service'
alias kdd='kubectl describe deployment'
alias kdn='kubectl describe node'
alias kaf='kubectl apply -f'
alias kdf='kubectl delete -f'
alias kex='kubectl exec -it'
alias kl='kubectl logs'
alias kpf='kubectl port-forward'

# Enable kubectl completion for aliases
complete -F __start_kubectl k

EOF

# Create kubeconfig directory
mkdir -p /root/.kube
mkdir -p /home/*/.kube 2>/dev/null || true

# Set up kubectl configuration template
cat > /etc/kubectl-config-template << 'EOF'
apiVersion: v1
clusters:
- cluster:
    server: https://kubernetes.default.svc.cluster.local
  name: default-cluster
contexts:
- context:
    cluster: default-cluster
    user: default-user
  name: default-context
current-context: default-context
kind: Config
preferences: {}
users:
- name: default-user
  user:
    token: ""
EOF

echo "Kubernetes tools configuration completed successfully!"
echo "kubectl version: $(kubectl version --client --short 2>/dev/null || echo 'Not configured')"
echo "helm version: $(helm version --short 2>/dev/null || echo 'Not configured')"
