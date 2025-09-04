---
name: Bug report
about: Create a report to help us improve GoldenPipe
title: '[BUG] '
labels: bug
assignees: ''

---

**Describe the bug**
A clear and concise description of what the bug is.

**To Reproduce**
Steps to reproduce the behavior:
1. Go to '...'
2. Click on '....'
3. Scroll down to '....'
4. See error

**Expected behavior**
A clear and concise description of what you expected to happen.

**Screenshots**
If applicable, add screenshots to help explain your problem.

**Environment (please complete the following information):**
 - OS: [e.g. Ubuntu 22.04, Windows Server 2022]
 - Kubernetes Version: [e.g. 1.28.0]
 - KubeVirt Version: [e.g. 1.1.0]
 - CDI Version: [e.g. 1.55.0]
 - GoldenPipe Version: [e.g. v1.0.0]

**Golden Image Details:**
 - OS Type: [e.g. linux, windows]
 - Base ISO URL: [e.g. https://releases.ubuntu.com/22.04/ubuntu-22.04.3-server-amd64.iso]
 - Customizations: [e.g. packages, scripts, files]

**Logs**
Please provide relevant logs:
```bash
kubectl logs deployment/goldenpipe -n goldenpipe-system
kubectl logs vm/<vm-name> -n goldenpipe-system
```

**Additional context**
Add any other context about the problem here.

**Compliance Information (if applicable):**
 - FedRAMP Level: [e.g. Moderate, High]
 - Data Classification: [e.g. Public, Sensitive, Confidential]
 - Environment: [e.g. Development, Staging, Production]
