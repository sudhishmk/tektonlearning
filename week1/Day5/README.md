# Day 5: Connecting Tasks with `Results`
---
## 🧠 Concept of the Day

While `Workspaces` are great for sharing files and directories, they are clumsy for passing small, specific pieces of data. For that, Tekton has **`Results`**.

A `Task` can declare that it will produce one or more `results`. Each result is a simple string value. Inside a `Task` step, you write the value to a special file that Tekton provides: `$(results.your-result-name.path)`. Tekton automatically reads the contents of this file when the `Task` completes.

Then, in a `Pipeline`, a subsequent `Task` can consume this output. You can use the result to populate a `param` in a later task using the variable substitution syntax: `$(tasks.PRODUCER_TASK_NAME.results.YOUR_RESULT_NAME)`. This creates a powerful data flow between your `Tasks`.

---
## 💼 Real-World Use Case

A very common pattern is to build a container image and then deploy it.
1.  **`build-and-push` Task**: This `Task` builds an image using Kaniko or Buildah. Upon pushing the image, the registry returns a unique, immutable image digest (e.g., `sha256:f1b3f...`). The `Task` writes this digest to a `result` named `image-digest`.
2.  **`deploy-to-k8s` Task**: This `Task` runs after the build. It takes the `image-digest` `result` from the first `Task` as a `param`. It then uses `kubectl set image` or a similar tool to update a Kubernetes Deployment to use the exact image digest, ensuring the deployed image is precisely the one that was just built and tested.

---
## 💻 Code/Config Example

A `Task` that "generates" a commit SHA and outputs it as a result.

**`producer-task.yaml`**
```yaml
apiVersion: tekton.dev/v1beta1
kind: Task
metadata:
  name: get-mock-commit-sha
spec:
  results:
    - name: commit-sha
      description: The git commit SHA.
  steps:
    - name: generate-sha
      image: bash
      script: |
        #!/bin/bash
        # Note: Using tr to remove the newline character
        echo -n "a1b2c3d4e5f6" > $(results.commit-sha.path)
```
---
## 🤔 Daily Self-Assessment

**Question:** What is the key difference between using a `Workspace` and a `Result` to pass data between `Tasks`?

**Answer:** A `Workspace` is for sharing a filesystem and is suited for large artifacts like source code or build binaries. A `Result` is for passing small, string-based metadata like a commit SHA, image digest, or a generated URL.

---
## 🛠️ Practical To-do exercise

Today, you'll modify yesterday's pipeline. You'll add a `Task` to get the real commit SHA from your cloned repository and another `Task` to print it, passing the SHA between them as a `Result`.

1.  **Navigate and Create Directory**:
    ```bash
    cd tekton-learning
    mkdir -p week1/day5
    ```

2.  **Create the SHA Producer `Task`**: This `Task` runs a `git` command in the workspace to find the commit SHA and writes it to a result. Create `week1/day5/get-commit-sha-task.yaml`:
    ```yaml
    # week1/day5/get-commit-sha-task.yaml
    apiVersion: tekton.dev/v1beta1
    kind: Task
    metadata:
      name: get-commit-sha
    spec:
      description: This task gets the commit SHA from a git repository in a workspace.
      workspaces:
        - name: source
          description: The workspace containing the git repository.
      results:
        - name: commit-sha
          description: The exact commit SHA that was checked out.
      steps:
        - name: get-sha
          image: alpine/git
          script: |
            #!/bin/sh
            set -e
            cd $(workspaces.source.path)
            # Use 'echo -n' to avoid the trailing newline that Tekton results dislike
            echo -n $(git rev-parse HEAD) > $(results.commit-sha.path)
            echo "Found SHA: $(cat $(results.commit-sha.path))"
    ```

3.  **Create the Echo Consumer `Task`**: This is a generic `Task` that just prints whatever message it receives. Create `week1/day5/echo-task.yaml`:
    ```yaml
    # week1/day5/echo-task.yaml
    apiVersion: tekton.dev/v1beta1
    kind: Task
    metadata:
      name: echo-message
    spec:
      params:
        - name: message
          type: string
          default: "No message provided."
      steps:
        - name: echo
          image: ubuntu
          script: 'echo "Message received: $(params.message)"'
    ```

4.  **Update the `Pipeline`**: Create a new pipeline definition in `week1/day5/pipeline.yaml`. This version will add our two new tasks to yesterday's pipeline.
    ```yaml
    # week1/day5/pipeline.yaml
    apiVersion: tekton.dev/v1beta1
    kind: Pipeline
    metadata:
      name: git-clone-and-report-sha
    spec:
      workspaces:
        - name: shared-data
      params:
        - name: repo-url
          type: string
      tasks:
        - name: fetch-repo
          taskRef:
            name: git-clone
          workspaces:
            - name: source
              workspace: shared-data
          params:
            - name: repo-url
              value: $(params.repo-url)

        - name: find-commit-id
          runAfter: ["fetch-repo"]
          taskRef:
            name: get-commit-sha
          workspaces:
            - name: source
              workspace: shared-data

        - name: report-commit-id
          runAfter: ["find-commit-id"]
          taskRef:
            name: echo-message
          params:
            - name: message
              value: "Cloned repo and found commit: $(tasks.find-commit-id.results.commit-sha)"
    ```

5.  **Create the `PipelineRun`**: Create `week1/day5/pipelinerun.yaml`.
    ```yaml
    # week1/day5/pipelinerun.yaml
    apiVersion: tekton.dev/v1beta1
    kind: PipelineRun
    metadata:
      name: report-sha-pipeline-run-1
    spec:
      pipelineRef:
        name: git-clone-and-report-sha
      workspaces:
        - name: shared-data
          persistentVolumeClaim:
            claimName: tekton-shared-workspace
      params:
        - name: repo-url
          value: [https://github.com/YOUR_USERNAME/tekton-learning.git](https://github.com/YOUR_USERNAME/tekton-learning.git) # <-- IMPORTANT: CHANGE THIS
    ```
    **Remember to replace `YOUR_USERNAME` with your GitHub username.**

6.  **Apply and Run**:
    ```bash
    # Apply the two new tasks
    kubectl apply -f week1/day5/get-commit-sha-task.yaml
    kubectl apply -f week1/day5/echo-task.yaml

    # Apply the new pipeline
    kubectl apply -f week1/day5/pipeline.yaml

    # Run it!
    kubectl apply -f week1/day5/pipelinerun.yaml
    ```

7.  **Check the Logs**:
    * Find the pods for your pipeline run: `kubectl get pods | grep report-sha-pipeline-run-1`
    * You'll see three pods. Find the pod for the `report-commit-id` task.
    * Check its logs: `kubectl logs <pod-for-report-commit-id>`.
    * You should see the final message, including the commit SHA from your `tekton-learning` repository! For example: `Message received: Cloned repo and found commit: 2d17c76...`

8.  **Commit Your Work**:
    ```bash
    git add .
    git commit -m "Day 5: Passing data between Tasks with Results"
    git push origin main
    ```

