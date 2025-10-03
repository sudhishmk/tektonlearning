# Day 15: Final Project - The Complete CI/CD Pipeline
---
## 🧠 Project Goal

Your mission is to create a fully automated pipeline that does the following when a developer pushes a commit to the `main` branch of your repository:
1.  **Authenticates** the webhook to ensure it's from a trusted source.
2.  **Clones** the repository.
3.  **Runs** the Go unit tests.
4.  **Builds** a container image using the `Dockerfile`.
5.  **Pushes** the image to Docker Hub with a unique tag.
6.  **Deploys** the application to your Kubernetes cluster by applying a `Deployment` manifest with the newly built image tag.
7.  **Reports** the final status of the pipeline, whether it succeeded or failed.

This project will be the centerpiece of your `tekton-learning` repository.

---
## 🛠️ Project Steps

We'll build this piece by piece. All work will be done in a new `final-project` directory.

### 1. The Application Manifest

Your pipeline needs a Kubernetes manifest to deploy.

* **Create the directory**:
    ```bash
    cd tekton-learning
    mkdir -p final-project/kubernetes
    ```
* **Create the `Deployment` manifest**: Create `final-project/kubernetes/deployment.yaml`. Use a placeholder for the image that we will replace during the pipeline run.

    ```yaml
    # final-project/kubernetes/deployment.yaml
    apiVersion: apps/v1
    kind: Deployment
    metadata:
      name: hello-ci-app
    spec:
      replicas: 1
      selector:
        matchLabels:
          app: hello-ci
      template:
        metadata:
          labels:
            app: hello-ci
        spec:
          containers:
          - name: app
            image: IMAGE_PLACEHOLDER # <-- Our pipeline will replace this
            ports:
            - containerPort: 8080 # Let's assume our Go app will listen on 8080
    ```
    *Note: We haven't built a web server yet, but we'll use this manifest as if we had.*

### 2. The Deployment `Task` and RBAC

The pipeline needs permission to interact with the Kubernetes API to create a `Deployment`.

* **Create the RBAC permissions**: Create `final-project/01-rbac.yaml`.
    ```yaml
    # final-project/01-rbac.yaml
    apiVersion: v1
    kind: ServiceAccount
    metadata:
      name: deployer-sa
    ---
    apiVersion: rbac.authorization.k8s.io/v1
    kind: Role
    metadata:
      name: deployer-role
    rules:
    - apiGroups: ["apps"]
      resources: ["deployments"]
      verbs: ["get", "list", "watch", "create", "delete", "patch", "update"]
    ---
    apiVersion: rbac.authorization.k8s.io/v1
    kind: RoleBinding
    metadata:
      name: deployer-role-binding
    subjects:
    - kind: ServiceAccount
      name: deployer-sa
    roleRef:
      kind: Role
      name: deployer-role
      apiGroup: rbac.authorization.k8s.io
    ```
* **Create the `kubectl-apply` Task**: Create `final-project/02-deploy-task.yaml`. This custom task will replace the image placeholder and apply the manifest.
    ```yaml
    # final-project/02-deploy-task.yaml
    apiVersion: tekton.dev/v1beta1
    kind: Task
    metadata:
      name: kubectl-apply
    spec:
      params:
        - name: image-reference
        - name: manifest-path
          default: "final-project/kubernetes/deployment.yaml"
      workspaces:
        - name: source
      steps:
        - name: update-and-apply
          image: alpine/k8s:1.29.0 # An image with kubectl and sed
          script: |
            #!/bin/sh
            set -ex

            cd $(workspaces.source.path)
            # Replace the placeholder with the real image digest from the build step
            sed -i "s|IMAGE_PLACEHOLDER|$(params.image-reference)|g" $(params.manifest-path)

            echo "--- Applying manifest ---"
            cat $(params.manifest-path)
            echo "-----------------------"

            kubectl apply -f $(params.manifest-path)
    ```

### 3. The Complete CI/CD `Pipeline`

Now, let's assemble the full pipeline. Create `final-project/03-pipeline.yaml`.

```yaml
# final-project/03-pipeline.yaml
apiVersion: tekton.dev/v1beta1
kind: Pipeline
metadata:
  name: complete-cicd-pipeline
spec:
  params:
    - name: repo-url
    - name: image-reference
  workspaces:
    - name: source
  tasks:
    - name: clone
      taskRef: { name: git-clone, bundle: "gcr.io/tekton-releases/catalog/upstream/git-clone:0.9" }
      params: [{ name: url, value: $(params.repo-url) }]
      workspaces: [{ name: output, workspace: source }]
    - name: test
      runAfter: ["clone"]
      taskRef: { name: go-test-with-results }
      workspaces: [{ name: source, workspace: source }]
    - name: build
      runAfter: ["test"]
      taskRef: { name: kaniko, bundle: "gcr.io/tekton-releases/catalog/upstream/kaniko:0.6" }
      params:
        - name: IMAGE
          value: $(params.image-reference)
        - name: DOCKERFILE
          value: week1/project/Dockerfile
      workspaces:
        - name: source
          workspace: source
    - name: deploy
      runAfter: ["build"]
      taskRef: { name: kubectl-apply }
      params:
        - name: image-reference
          value: $(tasks.build.results.IMAGE_DIGEST) # Use the exact digest from Kaniko
      workspaces:
        - name: source
          workspace: source
  finally:
    - name: report-status
      taskRef: { name: echo-message }
      params:
        - name: message
          value: "CI/CD Pipeline finished with status: $(tasks.status)"
```

### 4. The Smart `Trigger`

Finally, create the trigger that will run this pipeline only for pushes to `main`. Create `final-project/04-trigger.yaml`.

```yaml
# final-project/04-trigger.yaml
apiVersion: triggers.tekton.dev/v1beta1
kind: TriggerBinding
metadata:
  name: cicd-binding
spec:
  params:
    - name: repo-url
      value: $(body.repository.clone_url)
    - name: commit-sha
      value: $(body.head_commit.id)
---
apiVersion: triggers.tekton.dev/v1beta1
kind: TriggerTemplate
metadata:
  name: cicd-template
spec:
  params:
    - name: repo-url
    - name: commit-sha
  resourcetemplates:
    - apiVersion: tekton.dev/v1beta1
      kind: PipelineRun
      metadata:
        generateName: cicd-run-
      spec:
        pipelineRef: { name: complete-cicd-pipeline }
        serviceAccountName: image-builder-sa # For the Kaniko task
        params:
          - name: repo-url
            value: $(tt.params.repo-url)
          - name: image-reference
            value: docker.io/YOUR_DOCKERHUB_USERNAME/tekton-final-project:$(tt.params.commit-sha)
        workspaces:
        - name: source
          persistentVolumeClaim: { claimName: tekton-shared-workspace }
        taskRunSpecs: # Give the deploy task its own service account
          - pipelineTaskName: deploy
            taskServiceAccountName: deployer-sa
---
apiVersion: triggers.tekton.dev/v1beta1
kind: EventListener
metadata:
  name: cicd-listener
spec:
  serviceAccountName: tekton-pipelines-controller
  triggers:
    - name: main-branch-push
      interceptors:
        - { ref: { name: "github", kind: ClusterInterceptor }, params: [{ name: "secretRef", value: { secretName: github-webhook-secret, secretKey: secretToken } }] }
        - { ref: { name: "cel", kind: ClusterInterceptor }, params: [{ name: "filter", value: "body.ref == 'refs/heads/main'" }] }
      bindings:
        - ref: cicd-binding
      template:
        ref: cicd-template
```

Remember to replace YOUR_DOCKERHUB_USERNAME in the TriggerTemplate.

### 5. Execution and Verification

1.  **Apply all resources**:
    ```bash
    # Apply RBAC, Task, Pipeline, and Trigger
    kubectl apply -f final-project/
    ```
2.  **Expose the Listener**: In a new terminal, `kubectl port-forward service/el-cicd-listener 8080:8080`.
3.  **Simulate the Webhook**: Use the `payload-main-branch.json` from Day 13 and `curl` to send the event.
4.  **Watch**: `kubectl get pipelinerun -w`. Observe the entire flow execute.
5.  **Verify**:
    * Check Docker Hub for your new image, tagged with the commit SHA.
    * `kubectl get deployment hello-ci-app`.
    * `kubectl describe deployment hello-ci-app` and check that the image is set to the specific digest from the build.

---
## 🎉 Congratulations and Next Steps

You've done it! In 15 days, you have gone from the basics of Tekton to building a secure, automated, production-style CI/CD pipeline. You now have a solid foundation and a portfolio project to demonstrate your skills.

### Where to Go from Here?

* **Tekton Dashboard**: Install and explore the [Tekton Dashboard](https://github.com/tektoncd/dashboard) for a graphical UI to visualize your `PipelineRuns`.
* **Tekton Chains**: Dive into supply chain security. [Tekton Chains](https://github.com/tektoncd/chains) can be installed alongside Pipelines to automatically sign the container images you build, creating a verifiable record of their origin.
* **Custom Interceptors**: For very complex logic, learn how to write your own HTTP service that can act as a custom interceptor.
* **Contribute**: The Tekton community is active and welcoming. Explore the projects on GitHub, join the discussions, and consider contributing.

Thank you for following this learning plan. Happy building!

