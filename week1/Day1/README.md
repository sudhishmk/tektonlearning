# Day 1: The Core Duo - `Task` and `TaskRun`
---
## 🧠 Concept of the Day

In Tekton, the fundamental unit of work is a **`Task`**. A `Task` is a reusable, parameterized template that defines a sequence of **`Steps`**. Each `Step` is essentially a container execution (e.g., running a script, building an image, running tests).

However, a `Task` itself does nothing. It's just a blueprint. To execute it, you create an instance of it called a **`TaskRun`**. The `TaskRun` binds the `Task` to specific inputs (like a Git commit SHA) and triggers its execution on the cluster. Think of it like a class (`Task`) and an object (`TaskRun`).

---
## 💼 Real-World Use Case

A common CI/CD scenario is linting code before it can be merged. You would define a generic `Task` named `golang-lint` that checks out code and runs `golangci-lint`. Every time a developer opens a pull request, an automation tool (like a GitHub Action or a Tekton `Trigger`) would create a new `TaskRun` for this `golang-lint` `Task`, pointing it to the specific code from the PR. This ensures the same linting logic is applied consistently to every change.

---
## 💻 Code/Config Example

Here is a basic `Task` that simply echoes a message. It defines one `Step` that uses a standard `ubuntu` image.

**`task.yaml`**
```yaml
apiVersion: tekton.dev/v1beta1
kind: Task
metadata:
  name: hello-world-task
spec:
  description: "A simple task that prints a hello message."
  steps:
    - name: echo-hello
      image: ubuntu
      script: |
        #!/bin/bash
        echo "Hello, Tekton! This is the first step."
```
And here is the TaskRun that executes the Task defined above.

**`taskrun.yaml`**
```yaml
apiVersion: tekton.dev/v1beta1
kind: TaskRun
metadata:
  name: hello-world-task-run-1
spec:
  taskRef:
    name: hello-world-task
```

---
## 🤔 Daily Self-Assessment

**Question:** If you delete a `TaskRun` object from your Kubernetes cluster, what happens to the `Task` it was created from?

**Answer:** Nothing. The `Task` is a template and remains on the cluster, ready to be instantiated by new `TaskRun`s. Deleting the `TaskRun` only removes the record of that specific execution and cleans up its associated Pod.

---
## 🛠️ Practical To-do exercise

Today, you'll set up your environment and run your first `Task`. This will form the basis for all future exercises.

1.  **Set up a local cluster**: If you don't have one, install [Minikube](https://minikube.sigs.k8s.io/docs/start/) or [Kind](https://kind.sigs.k8s.io/docs/user/quick-start/). Start your cluster.
    * For Minikube: `minikube start --memory 4g --cpus 2`
    * For Kind: `kind create cluster`

2.  **Install Tekton Pipelines**: Apply the latest release from the official docs.
    ```bash
    kubectl apply --filename https://storage.googleapis.com/tekton-releases/pipeline/latest/release.yaml
    ```
    Verify the installation by checking for pods in the `tekton-pipelines` namespace: `kubectl get pods --namespace tekton-pipelines`. Wait for them to be `Running`.

3.  **Create your Git repository**:
    * Create a new public repository on GitHub called `tekton-learning`.
    * Clone it to your local machine: `git clone https://github.com/YOUR_USERNAME/tekton-learning.git`.
    * `cd tekton-learning`

4.  **Create Day 1 files**:
    * Create a directory structure: `mkdir -p week1/day1`.
    * Inside `week1/day1`, create the `task.yaml` and `taskrun.yaml` files with the content from the **Code/Config Example** section above.

5.  **Run your Task**:
    * Apply the `Task` definition to your cluster: `kubectl apply -f week1/day1/task.yaml`.
    * Trigger the execution by applying the `TaskRun`: `kubectl apply -f week1/day1/taskrun.yaml`.

6.  **Check the results**:
    * See the `TaskRun`'s status: `kubectl get taskrun hello-world-task-run-1`. Wait for `SUCCEEDED` to be `True`.
    * View the logs from the `TaskRun`. The easiest way is with the `tkn` CLI (highly recommended!), but you can also find the pod created by the `TaskRun`.
        * Find the pod: `kubectl get pods | grep hello-world-task-run-1`
        * View its logs: `kubectl logs <pod-name-from-previous-command>`
        * You should see the output: `"Hello, Tekton! This is the first step."`

7.  **Commit your work**:
    ```bash
    git add .
    git commit -m "Day 1: First Tekton Task and TaskRun"
    git push origin main
    ```

You have now successfully defined and executed a piece of work using Tekton's core components!

[Go to Day 2](Day2/README.md)
