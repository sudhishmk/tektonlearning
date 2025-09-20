# Prerequisites for the 15-Day Tekton Expertise Plan

Welcome to the 15-day journey to becoming a Tekton expert! This guide will ensure you have all the necessary knowledge, tools, and configurations in place before you begin. A little preparation now will allow you to focus entirely on Tekton's concepts and practices each day.

---

### 🎯 Learning Goal

The objective of this plan is to take you from an intermediate understanding of CI/CD and Kubernetes to an expert level in using Tekton to build powerful, scalable, and cloud-native pipelines.

### 🧑‍💻 Who is this plan for?

This plan is designed for individuals with an **intermediate** level of DevOps experience. You should be comfortable with the following core concepts:

* **Kubernetes Fundamentals:** You understand what Pods, Deployments, Services, and PersistentVolumeClaims (PVCs) are and have used `kubectl` to interact with a cluster.
* **Containerization:** You know how to build a Dockerfile and understand the concept of container images and registries.
* **CI/CD Principles:** You are familiar with the basic stages of a CI/CD workflow (e.g., build, test, deploy).
* **YAML:** You are comfortable reading and writing YAML configuration files.
* **Basic Shell Scripting:** You can read and understand simple `sh` or `bash` scripts.

---

### ✅ Required Local Tooling

Please install and configure the following tools on your local machine before starting Day 1.

| Tool | Purpose | Installation Guide & Verification |
| :--- | :--- | :--- |
| **`kubectl`** | The standard Kubernetes CLI. | **Install:** [Official Docs](https://kubernetes.io/docs/tasks/tools/install-kubectl/) <br/> **Verify:** `kubectl version --client` |
| **Local K8s Cluster** | A sandbox to run Tekton. | We recommend **Minikube**. <br/> **Install:** [Official Docs](https://minikube.sigs.k8s.io/docs/start/) <br/> **Verify:** `minikube status` |
| **`tkn` (Tekton CLI)** | The official CLI for Tekton. | **Install:** [Official Docs](https://tekton.dev/docs/cli/install/) <br/> **Verify:** `tkn version` |
| **Git** | For version control and cloning exercise repos. | **Install:** [Official Docs](https://git-scm.com/book/en/v2/Getting-Started-Installing-Git) <br/> **Verify:** `git --version` |

---

### 🔧 Environment Setup (One-Time Task)

Complete these steps once before you begin Day 1. This will prepare your local Kubernetes cluster with Tekton.

**1. Start Your Minikube Cluster**
We need a cluster with sufficient resources. Start Minikube with the following command:
```bash
minikube start --memory=4g --cpus=2 --driver=docker
