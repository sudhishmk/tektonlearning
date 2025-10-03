# Day 11: Building Images with Kaniko & Handling Secrets
---
## 🧠 Concept of the Day

To build container images in a Kubernetes-native way, you can't rely on a Docker daemon. The community standard for this is **Kaniko**. Kaniko is a tool that builds a container image from a `Dockerfile` and pushes it to a registry, all from within a standard, unprivileged container—perfect for Tekton.

Pushing to a registry requires authentication. The proper way to handle sensitive data like registry credentials in Kubernetes is with **`Secrets`**. You create a `Secret` of type `kubernetes.io/dockerconfigjson` containing your credentials.

To make this `Secret` available to your `Task's` Pod, you don't mount it directly. Instead, you link it to a **`ServiceAccount`**. Then, you execute your `PipelineRun` using that `ServiceAccount`. Tekton automatically finds the linked image pull secret and makes it available to tools like Kaniko, which are designed to look for it in the standard location.

The flow is: **`Secret` -> linked to `ServiceAccount` -> used by `PipelineRun` -> consumed by `Task`**.

---
## 💼 Real-World Use Case

A CI pipeline successfully tests a Java application. The next stage is to build the application image. The `PipelineRun` is configured to use a `ServiceAccount` named `image-pusher`. This `ServiceAccount` is linked to a `Secret` containing credentials for the company's private Harbor registry. A `Task` using the official Kaniko bundle from Artifact Hub executes. Kaniko reads the `Dockerfile`, builds the image, and automatically discovers the credentials mounted from the `ServiceAccount`'s `Secret` to authenticate and push the new image to Harbor.

---
## 💻 Code/Config Example

First, a `Secret` to hold Docker Hub credentials.
```bash
# You don't usually write YAML for this. Use the kubectl command:
kubectl create secret docker-registry docker-credentials \
  --docker-server=[https://index.docker.io/v1/](https://index.docker.io/v1/) \
  --docker-username=YOUR_DOCKERHUB_USERNAME \
  --docker-password=YOUR_DOCKERHUB_PASSWORD_OR_TOKEN
```

---
## 🤔 Daily Self-Assessment

**Question:** Why is it a security best practice to use a `ServiceAccount` to link a `Secret` to a `PipelineRun` instead of passing credentials as `params`?

**Answer:** It's a fundamental security principle of least privilege and separation of concerns. `Secrets` are a dedicated Kubernetes object for sensitive data, which can be controlled with RBAC and are not stored in plaintext in the `PipelineRun`'s definition. Passing credentials as `params` would expose them in logs and the `PipelineRun` object, which is a major security risk.

---
## 🛠️ Practical To-do exercise

Today, you will add a real image build-and-push step to your CI pipeline.

**Prerequisite**: You need a free account on [Docker Hub](https://hub.docker.com/).

1.  **Navigate and Create Directory**:
    ```bash
    cd tekton-learning
    mkdir -p week2/day11
    ```

2.  **Add a `Dockerfile` to your Project**: Create a file named `Dockerfile` inside your `week1/project/` directory. This is a simple, multi-stage build.
    ```Dockerfile
    # week1/project/Dockerfile

    # Build stage
    FROM golang:1.22-alpine AS builder
    WORKDIR /app
    COPY . .
    RUN go build -o /app/hello-ci .

    # Final stage
    FROM alpine:latest
    WORKDIR /app
    COPY --from=builder /app/hello-ci .
    CMD ["/app/hello-ci"]
    ```
    Now, add, commit, and push this new file to your GitHub repository.
    ```bash
    git add week1/project/Dockerfile
    git commit -m "feat: Add Dockerfile for Go app"
    git push origin main
    ```

3.  **Create Kubernetes Resources**:
    * **Create the `Secret`**: Run this command, substituting your actual Docker Hub username and password/access token.
        ```bash
        kubectl create secret docker-registry docker-credentials \
          --docker-username=YOUR_DOCKERHUB_USERNAME \
          --docker-password=YOUR_DOCKERHUB_PASSWORD
        ```
    * **Create the `ServiceAccount`**: Create `week2/day11/service-account.yaml`:
        ```yaml
        # week2/day11/service-account.yaml
        apiVersion: v1
        kind: ServiceAccount
        metadata:
          name: image-builder-sa
        secrets:
          - name: docker-credentials
        ```

4.  **Create the Build & Push `Pipeline`**: Create `week2/day11/pipeline.yaml`. This pipeline adds a Kaniko task from Artifact Hub.
    ```yaml
    # week2/day11/pipeline.yaml
    apiVersion: tekton.dev/v1beta1
    kind: Pipeline
    metadata:
      name: test-and-build-pipeline
    spec:
      params:
        - name: repo-url
        - name: image-reference # e.g., docker.io/your-user/tekton-learning
      workspaces:
        - name: shared-workspace

      tasks:
        - name: clone-repo
          taskRef:
            name: git-clone
            bundle: gcr.io/tekton-releases/catalog/upstream/git-clone:0.9
          params:
            - name: url
              value: $(params.repo-url)
          workspaces:
            - name: output
              workspace: shared-workspace

        - name: run-go-tests
          runAfter: ["clone-repo"]
          taskRef:
            name: go-test-with-results
          workspaces:
            - name: source
              workspace: shared-workspace

        - name: build-and-push
          runAfter: ["run-go-tests"]
          taskRef:
            name: kaniko
            bundle: gcr.io/tekton-releases/catalog/upstream/kaniko:0.6
          params:
            - name: IMAGE
              value: $(params.image-reference)
          workspaces:
            - name: source
              workspace: shared-workspace
            - name: dockerconfig
              workspace: not-used # Kaniko auto-detects the SA secret
    ```
    *Note: We must declare the `dockerconfig` workspace for the Kaniko task, but we don't provide it. This signals to Kaniko to use the `ServiceAccount`'s secret.*

5.  **Create the `PipelineRun`**: Create `week2/day11/pipelinerun.yaml`. **This MUST include the `serviceAccountName`**.
    ```yaml
    # week2/day11/pipelinerun.yaml
    apiVersion: tekton.dev/v1beta1
    kind: PipelineRun
    metadata:
      name: test-build-push-run-1
    spec:
      pipelineRef:
        name: test-and-build-pipeline
      serviceAccountName: image-builder-sa # <-- CRITICAL
      params:
        - name: repo-url
          value: [https://github.com/YOUR_USERNAME/tekton-learning.git](https://github.com/YOUR_USERNAME/tekton-learning.git)
        - name: image-reference
          value: docker.io/YOUR_DOCKERHUB_USERNAME/tekton-learning:latest # <-- CRITICAL
      workspaces:
        - name: shared-workspace
          persistentVolumeClaim:
            claimName: tekton-shared-workspace
    ```
    **Remember to replace `YOUR_USERNAME` and `YOUR_DOCKERHUB_USERNAME`**.

6.  **Apply and Run**:
    ```bash
    kubectl apply -f week2/day11/service-account.yaml
    kubectl apply -f week2/day11/pipeline.yaml
    kubectl apply -f week2/day11/pipelinerun.yaml
    ```

7.  **Verify the Result**:
    * Watch the run: `kubectl get pipelinerun test-build-push-run-1 -w`. The Kaniko step will take a minute or two.
    * If it succeeds, log in to your Docker Hub account. You should see a new `tekton-learning` repository with a `latest` tag!

8.  **Commit Your Work**:
    ```bash
    git add .
    git commit -m "Day 11: Add Kaniko task to build and push images"
    git push origin main
    ```
