# Day 3: Making Tasks Reusable with `Parameters`
---
## 🧠 Concept of the Day

So far, our tasks have been static. To make them truly reusable, we need a way to pass in different inputs each time we run them. This is the job of **`Parameters` (or `params`)**.

You declare `params` in a `Task`'s specification, giving each a `name`, `type` (usually `string` or `array`), and an optional `default` value. Inside your `Task`'s steps, you access these values using the variable substitution syntax: `$(params.your-param-name)`.

When you create a `TaskRun`, you provide the concrete values for the `params` declared in the `Task`. If you provide a value in the `TaskRun`, it overrides the `default` value from the `Task`. This allows you to create generic tasks (e.g., `git-clone`, `build-image`) that can be customized for specific needs at runtime.

---
## 💼 Real-World Use Case

Imagine a single, company-wide `Task` for building container images with Kaniko. Instead of creating a new task for every microservice, you create one generic `build-image` `Task`. This `Task` would have `params` like:
* `IMAGE_URL`: The full name of the image to build and push (e.g., `gcr.io/my-project/my-app:v1.2.3`).
* `DOCKERFILE_PATH`: The path to the Dockerfile within the source code workspace.
* `BUILD_CONTEXT`: The directory to run the build from.

Each team can then create `TaskRun`s for this `Task`, providing the specific values for their service, promoting consistency and reducing boilerplate.

---
## 💻 Code/Config Example

This `Task` defines two parameters: `greeting` and `recipient`. Notice that `recipient` has a default value.

**`task.yaml`**
```yaml
apiVersion: tekton.dev/v1beta1
kind: Task
metadata:
  name: personalized-greeting
spec:
  params:
    - name: greeting
      type: string
      description: The salutation to use.
    - name: recipient
      type: string
      description: The person or thing to greet.
      default: "World"
  steps:
    - name: greet
      image: ubuntu
      script: |
        #!/bin/bash
        echo "$(params.greeting), $(params.recipient)!"
```
---
## 🤔 Daily Self-Assessment

**Question:** What happens if a `Task` declares a `param` *without* a default value, and the `TaskRun` does *not* provide a value for it?

**Answer:** The `TaskRun` will fail validation and will not be created. Tekton's admission webhook will reject it with an error stating that a required parameter is missing.

---
## 🛠️ Practical To-do exercise

Today, we'll create a reusable `git-clone` task and use it to clone the `tekton-learning` repository you created. This task will combine `params` and a `workspace`.

1.  **Navigate and Create Directory**:
    ```bash
    cd tekton-learning
    mkdir -p week1/day3
    ```

2.   **Create the PVC**: Create a file named `week1/day3/pvc.yaml` with this content. This requests 1Gi of storage from your cluster.
    ```yaml
    # week1/day2/pvc.yaml
    apiVersion: v1
    kind: PersistentVolumeClaim
    metadata:
      name: tekton-shared-workspace-2
    spec:
      accessModes:
        - ReadWriteOnce # This PVC can be mounted by one node at a time
      resources:
        requests:
          storage: 1Gi
    ```

3. **Create the `git-clone` Task**: Create a file at `week1/day3/task.yaml`. This `Task` will be a generic. It requires a `workspace` to clone into and takes a `repo-url` parameter.
    ```yaml
    # week1/day3/task.yaml
    apiVersion: tekton.dev/v1beta1
    kind: Task
    metadata:
      name: git-clone
    spec:
      description: "This task clones a git repository into a workspace."
      workspaces:
        - name: source
          description: The workspace where the source code will be cloned.
      params:
        - name: repo-url
          type: string
          description: The URL of the git repository to clone.
      steps:
        - name: clone
          image: alpine/git # A small image with git installed
          script: |
            #!/bin/sh
            set -ex

            echo "Cloning repository $(params.repo-url)..."
            git clone $(params.repo-url) $(workspaces.source.path)

            echo "Contents of the workspace:"
            ls -la $(workspaces.source.path)
    ```

3.  **Create the TaskRun**: Now, create `week1/day3/taskrun.yaml`. This `TaskRun` will use the generic `git-clone` `Task` but provide *specific* values for your repository. It binds the `source` workspace to the PVC you created on Day 2.
    ```yaml
    # week1/day3/taskrun.yaml
    apiVersion: tekton.dev/v1beta1
    kind: TaskRun
    metadata:
      name: clone-our-learning-repo
    spec:
      taskRef:
        name: git-clone
      params:
        - name: repo-url
          value: https://github.com/YOUR_USERNAME/tekton-learning.git# <-- IMPORTANT: CHANGE THIS
      workspaces:
        - name: source
          persistentVolumeClaim:
            claimName: tekton-shared-workspace-2 # Re-using the PVC from Day 2
    ```
    **Remember to replace `YOUR_USERNAME` with your actual GitHub username.**

4.  **Apply and Run**:
    ```bash
    # Create a new pvc
    # kubectl apply -f week1/day3/pvc.yaml

    # 1. Apply the new generic Task
    kubectl apply -f week1/day3/task.yaml

    # 2. Run it with specific parameters
    kubectl apply -f week1/day3/taskrun.yaml
    ```

5.  **Check the Logs**:
    * Find the Pod for the new `TaskRun`: `kubectl get pods | grep clone-our-learning-repo`
    * Check its logs: `kubectl logs <pod-name-from-previous-command>`
    * You should see the output of the `git clone` command and the `ls -la` command, showing the files from your repository!

6.  **Commit Your Work**:
    ```bash
    git add .
    git commit -m "Day 3: Creating a reusable git-clone task with params"
    git push origin main
    ```
    
    
