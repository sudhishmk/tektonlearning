# Day 2: Sharing is Caring - `Workspaces`
---
## 🧠 Concept of the Day

A `Task`'s steps run as separate containers within the same Pod, but they don't share a filesystem by default beyond an empty `/workspace` directory. To share data—like source code cloned in one step and tested in another—you use a **`Workspace`**.

A `Workspace` is a declaration in a `Task` that says, "I need a volume mounted at this path." It's an abstract request for storage. The `TaskRun` is then responsible for fulfilling this request by binding the declared `Workspace` to a concrete storage source, most commonly a **PersistentVolumeClaim (PVC)**. This decouples your task logic from the specific storage implementation, making tasks highly reusable.

---
## 💼 Real-World Use Case

Consider a typical build process. A `Task` has two steps:
1.  **`git-clone`**: This step checks out the source code from a repository into a shared directory.
2.  **`maven-build`**: This step runs `mvn package` on the source code.

For the second step to see the code cloned by the first, both must mount the same `Workspace`. The `TaskRun` would provide a PVC that gets mounted into both steps' containers, allowing the build step to access the output of the clone step.

---
## 💻 Code/Config Example

First, you need a `PersistentVolumeClaim` to provide the actual storage.

**`pvc.yaml`**
```yaml
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: tekton-shared-workspace
spec:
  accessModes:
    - ReadWriteOnce
  resources:
    requests:
      storage: 1Gi
```

---
## 🤔 Daily Self-Assessment

**Question:** What happens if a `Task` defines two `Steps`, and both try to write a file with the same name (e.g., `results.txt`) to the same `Workspace`?

**Answer:** The steps execute sequentially. The second step would simply overwrite the file created by the first step. This is a standard filesystem behavior and a common pattern for passing artifacts.

---
## 🛠️ Practical To-do exercise

Today, you'll create a task with two steps that communicate by writing and reading from a shared PVC-backed workspace.

1.  **Navigate to your repo**: `cd tekton-learning`.

2.  **Create Day 2 Directory**: `mkdir -p week1/day2`.

3.  **Create the PVC**: Create a file named `week1/day2/pvc.yaml` with this content. This requests 1Gi of storage from your cluster.
    ```yaml
    # week1/day2/pvc.yaml
    apiVersion: v1
    kind: PersistentVolumeClaim
    metadata:
      name: tekton-shared-workspace
    spec:
      accessModes:
        - ReadWriteOnce # This PVC can be mounted by one node at a time
      resources:
        requests:
          storage: 1Gi
    ```

4.  **Create the Task**: Create `week1/day2/task.yaml`. Note the `workspaces` declaration and the `$(workspaces.output.path)` variable used to get the mount path.
    ```yaml
    # week1/day2/task.yaml
    apiVersion: tekton.dev/v1beta1
    kind: Task
    metadata:
      name: create-and-read-file
    spec:
      workspaces:
        - name: output
          description: The workspace where a file will be written and then read.
      steps:
        - name: create-file
          image: ubuntu
          script: |
            #!/bin/bash
            echo "Hello from Step 1!" > $(workspaces.output.path)/my-file.txt
            echo "File created. Contents of workspace:"
            ls -l $(workspaces.output.path)

        - name: read-file
          image: ubuntu
          script: |
            #!/bin/bash
            echo "---"
            echo "Reading file in Step 2..."
            cat $(workspaces.output.path)/my-file.txt
    ```

5.  **Create the TaskRun**: Create `week1/day2/taskrun.yaml`. This file connects the abstract `output` workspace in the `Task` to the concrete `tekton-shared-workspace` PVC.
    ```yaml
    # week1/day2/taskrun.yaml
    apiVersion: tekton.dev/v1beta1
    kind: TaskRun
    metadata:
      name: create-and-read-file-run
    spec:
      taskRef:
        name: create-and-read-file
      workspaces:
        - name: output
          persistentVolumeClaim:
            claimName: tekton-shared-workspace
    ```

6.  **Apply and Run**: Apply the files in order.
    ```bash
    # 1. Create the storage
    kubectl apply -f week1/day2/pvc.yaml

    # 2. Create the Task template
    kubectl apply -f week1/day2/task.yaml

    # 3. Run the Task
    kubectl apply -f week1/day2/taskrun.yaml
    ```

7.  **Check the logs**:
    * Get the name of the `TaskRun` pod: `kubectl get pods | grep create-and-read-file-run`
    * View the logs: `kubectl logs <pod-name-from-previous-command> --all-containers`
    * You should see the output from both steps, with the second step successfully printing "Hello from Step 1!".

8.  **Commit your work**:
    ```bash
    git add .
    git commit -m "Day 2: Using Workspaces to share data between steps"
    git push origin main
    ```
