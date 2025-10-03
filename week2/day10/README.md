# Day 10: Graceful Exits - `finally` Tasks
---
## 🧠 Concept of the Day

In a `Pipeline`, if any `Task` fails, the entire `PipelineRun` stops, and subsequent tasks are not executed. This can leave your environment in a messy state (e.g., a temporary database is left running). To solve this, Tekton has a **`finally`** block.

The `finally` block in a `Pipeline` defines a set of tasks that are **guaranteed to execute** after all `tasks` in the main body have completed or failed. These final tasks are perfect for cleanup, sending notifications, or updating a deployment status.

Tasks within the `finally` block have access to a special variable, **`$(tasks.status)`**, which holds the overall status of the regular tasks (`Succeeded`, `Failed`, etc.). This allows your cleanup task to know the outcome of the pipeline and act accordingly.

---
## 💼 Real-World Use Case

A pipeline provisions a temporary Kubernetes namespace for integration testing. The main tasks then deploy an application and run tests within that namespace.
1.  **`setup-task`**: Creates the test namespace.
2.  **`test-task`**: Runs the integration tests. This task might fail.
3.  **`cleanup-task` (in `finally`)**: This task runs `kubectl delete namespace <test-namespace>`.

Whether the `test-task` succeeds or fails, the `cleanup-task` in the `finally` block is guaranteed to run, ensuring the temporary namespace is always deleted and preventing resource leaks.

---
## 💻 Code/Config Example

This `Pipeline` includes a `finally` task that reports the overall status.

**`pipeline.yaml`**
```yaml
apiVersion: tekton.dev/v1beta1
kind: Pipeline
metadata:
  name: finally-demo-pipeline
spec:
  tasks:
    - name: main-task
      taskRef:
        name: echo-message
      params:
        - name: message
          value: "Running the main part of the pipeline..."

  finally: # <-- This block executes regardless of the tasks' status
    - name: cleanup-and-notify
      taskRef:
        name: echo-message
      params:
        - name: message # This param receives the overall status
          value: "Pipeline complete. Final status was: $(tasks.status)"
```

---
## 🤔 Daily Self-Assessment

**Question:** Can a `Task` in the `finally` block have a `runAfter` dependency on a regular `Task` in the main `tasks` block?

**Answer:** No. `finally` tasks exist outside the main execution graph. They cannot have `runAfter` dependencies on regular tasks, and regular tasks cannot depend on them. They are designed to run only after the main directed acyclic graph (DAG) of tasks is complete.

---
## 🛠️ Practical To-do exercise

Today, you'll build a pipeline that intentionally fails but still performs a cleanup action using a `finally` task.

1.  **Navigate and Create Directory**:
    ```bash
    cd tekton-learning
    mkdir -p week2/day10
    ```

2.  **Create a `Task` that Fails**: For this exercise, we need a task that is guaranteed to fail. Create `week2/day10/failing-task.yaml`:
    ```yaml
    # week2/day10/failing-task.yaml
    apiVersion: tekton.dev/v1beta1
    kind: Task
    metadata:
      name: intentionally-fail
    spec:
      steps:
        - name: fail
          image: ubuntu
          script: |
            #!/bin/bash
            echo "This task is about to fail..."
            exit 1
    ```

3.  **Create the `Pipeline` with a `finally` block**: Create `week2/day10/pipeline.yaml`. This pipeline will run the failing task, but the `finally` block will still execute.

    ```yaml
    # week2/day10/pipeline.yaml
    apiVersion: tekton.dev/v1beta1
    kind: Pipeline
    metadata:
      name: robust-pipeline-with-finally
    spec:
      tasks:
        - name: attempt-main-work
          taskRef:
            name: intentionally-fail # Our new failing task

      finally:
        - name: report-final-status
          taskRef:
            name: echo-message # Our reusable echo task
          params:
            - name: message
              value: "🔔 Final Report: The pipeline finished with status: $(tasks.status)"
    ```

4.  **Create the `PipelineRun`**: Create `week2/day10/pipelinerun.yaml`.
    ```yaml
    # week2/day10/pipelinerun.yaml
    apiVersion: tekton.dev/v1beta1
    kind: PipelineRun
    metadata:
      name: finally-run-1
    spec:
      pipelineRef:
        name: robust-pipeline-with-finally
    ```

5.  **Apply and Run**:
    ```bash
    # Apply the new failing task and the pipeline
    kubectl apply -f week2/day10/failing-task.yaml
    kubectl apply -f week2/day10/pipeline.yaml

    # Run it!
    kubectl apply -f week2/day10/pipelinerun.yaml
    ```

6.  **Verify the Result**:
    * Watch the `PipelineRun`: `kubectl get pipelinerun finally-run-1 -w`. You will see it eventually enter the `Failed` state. This is expected.
    * Now, find the `TaskRuns` created by this `PipelineRun`:
        ```bash
        kubectl get taskruns | grep finally-run-1
        ```
        You will see two `TaskRuns`: one for `attempt-main-work` (which failed) and one for `report-final-status` (which succeeded).
    * Check the logs of the `report-final-status` pod:
        ```bash
        # Find the pod name for the report-final-status task
        kubectl logs <pod-name-for-report-final-status>
        ```
        The output will be: `"🔔 Final Report: The pipeline finished with status: Failed"`. This proves your `finally` task ran successfully even when the main pipeline failed.

7.  **Commit Your Work**:
    ```bash
    git add .
    git commit -m "Day 10: Add finally tasks for guaranteed execution"
    git push origin main
    ```


