# Day 4: From Tasks to Workflows with `Pipelines`
---
## 🧠 Concept of the Day

While a `Task` is a sequence of `Steps`, a **`Pipeline`** is a sequence of `Tasks`. It's the resource you use to define your entire CI/CD workflow graph. You can specify the order of execution, run tasks in parallel, and create complex dependencies.

To execute a `Pipeline`, you create a **`PipelineRun`**. Much like a `TaskRun` instantiates a `Task`, a `PipelineRun` instantiates a `Pipeline`. It binds the `Pipeline` to specific runtime parameters and workspaces.

When a `PipelineRun` starts, it creates the necessary `TaskRun`s for each `Task` in the `Pipeline` according to the defined order. You control this order using the `runAfter` keyword, which tells a `Task` to wait for another one to complete successfully before starting.

---
## 💼 Real-World Use Case

A standard Continuous Integration (CI) pipeline is a perfect example. A `Pipeline` resource would define the following flow:
1.  **`clone-repo` Task**: The starting point with no dependencies.
2.  **`run-linter` Task**: Runs after `clone-repo`.
3.  **`run-unit-tests` Task**: Also runs after `clone-repo`. Notice that the linter and unit tests can run in parallel to save time.
4.  **`build-image` Task**: This task would have `runAfter: [run-linter, run-unit-tests]`, ensuring it only runs if both the linting and testing stages pass successfully.

A `PipelineRun` for this `Pipeline` would be triggered automatically on every new pull request.

---
## 💻 Code/Config Example

First, a `Pipeline` that defines two `Tasks` to be run sequentially.

**`pipeline.yaml`**
```yaml
apiVersion: tekton.dev/v1beta1
kind: Pipeline
metadata:
  name: demo-pipeline
spec:
  workspaces:
    - name: shared-data
  params:
    - name: repo-url
      type: string

  tasks:
    - name: fetch-repo
      taskRef:
        name: git-clone # Using the Task we created on Day 3
      workspaces:
        - name: source # The workspace name from the Task
          workspace: shared-data # The workspace name from this Pipeline
      params:
        - name: repo-url
          value: $(params.repo-url) # Pass the Pipeline's param to the Task's param

    - name: list-files
      runAfter: ["fetch-repo"] # <-- This enforces sequential order
      taskRef:
        name: list-files-task # A simple task we will create
      workspaces:
        - name: directory
          workspace: shared-data
```

---
## 🤔 Daily Self-Assessment

**Question:** If a `Pipeline` defines two `Tasks`, 'A' and 'B', and Task 'B' has `runAfter: [A]`, what happens to the overall `PipelineRun` if Task 'A' fails?

**Answer:** Task 'B' will not be scheduled to run. The `PipelineRun` will stop execution and be marked with a "Failed" status.

---
## 🛠️ Practical To-do exercise

Today you'll build and run your first `Pipeline`, which will use the `git-clone` task from yesterday and a new task to inspect the cloned files.

1.  **Navigate and Create Directory**:
    ```bash
    cd tekton-learning
    mkdir -p week1/day4
    ```

2.  **Create a Helper `Task`**: First, create a new, simple task that lists files recursively. Create `week1/day4/list-files-task.yaml`:
    ```yaml
    # week1/day4/list-files-task.yaml
    apiVersion: tekton.dev/v1beta1
    kind: Task
    metadata:
      name: list-files-task
    spec:
      description: "This task lists all files in a workspace recursively."
      workspaces:
        - name: directory
          description: The directory to inspect.
      steps:
        - name: list
          image: ubuntu
          script: |
            #!/bin/bash
            echo "Listing all files in the workspace..."
            ls -R $(workspaces.directory.path)
    ```

3.  **Create the `Pipeline`**: Now define the workflow in `week1/day4/pipeline.yaml`. This pipeline orchestrates the `git-clone` and `list-files-task` `Tasks`.
    ```yaml
    # week1/day4/pipeline.yaml
    apiVersion: tekton.dev/v1beta1
    kind: Pipeline
    metadata:
      name: git-clone-and-list-pipeline
    spec:
      description: "This pipeline clones a repo and then lists its contents."
      workspaces:
        - name: shared-data
          description: The workspace that will be shared between tasks.
      params:
        - name: repo-url
          type: string
          description: The git repo URL to clone.

      tasks:
        - name: fetch-from-git
          taskRef:
            name: git-clone # Assumes the Task from Day 3 is on the cluster
          workspaces:
            - name: source
              workspace: shared-data
          params:
            - name: repo-url
              value: $(params.repo-url)

        - name: list-cloned-files
          runAfter: ["fetch-from-git"] # <-- Ensures this runs second
          taskRef:
            name: list-files-task
          workspaces:
            - name: directory
              workspace: shared-data
    ```

4.  **Create the `PipelineRun`**: Create `week1/day4/pipelinerun.yaml` to trigger the pipeline.
    ```yaml
    # week1/day4/pipelinerun.yaml
    apiVersion: tekton.dev/v1beta1
    kind: PipelineRun
    metadata:
      name: clone-list-pipeline-run-2
    spec:
      pipelineRef:
        name: git-clone-and-list-pipeline
      workspaces:
        - name: shared-data
          persistentVolumeClaim:
            claimName: tekton-shared-workspace
      params:
        - name: repo-url
          value: https://github.com/YOUR_USERNAME/tektonlearning.git # <-- IMPORTANT: CHANGE THIS
    ```
    **Remember to replace `YOUR_USERNAME` with your actual GitHub username.**

5.  **Apply and Run**: Apply the files in order.
    ```bash
    # Make sure tasks from previous days are on the cluster
    # kubectl apply -f week1/day3/task.yaml

    # 1. Apply the new helper task
    kubectl apply -f week1/day4/list-files-task.yaml

    # 2. Apply the pipeline definition
    kubectl apply -f week1/day4/pipeline.yaml

    # 3. Trigger the pipeline run
    kubectl apply -f week1/day4/pipelinerun.yaml
    ```

6.  **Check the Logs**: A `PipelineRun` creates `TaskRun`s.
    * Watch the `PipelineRun`'s progress: `kubectl get pipelinerun clone-list-pipeline-run-1 -w`.
    * Once it completes, find the Pods it created: `kubectl get pods | grep clone-list-pipeline-run-2`. You will see two Pods, one for each `Task`.
    * Check the logs of the second pod (the one for `list-cloned-files`). You should see the file structure of your Git repository!

7.  **Commit Your Work**:
    ```bash
    git add .
    git commit -m "Day 4: Creating a multi-task Pipeline"
    git push origin main
    ```

[Go to Day 5](../Day5/README.md)
