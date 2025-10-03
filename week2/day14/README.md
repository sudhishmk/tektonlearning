# Day 14: Preventing Runaway Pipelines with `Timeouts`
---
## 🧠 Concept of the Day

In a production environment, you cannot allow pipelines to run forever. A stuck network call, an infinite loop in a script, or a cluster resource issue could cause a `PipelineRun` to hang indefinitely, consuming valuable resources.

To prevent this, Tekton allows you to specify **`Timeouts`**. You can set a timeout for the entire pipeline or for individual tasks. If a `PipelineRun` or `TaskRun` exceeds its allotted time, Tekton will gracefully terminate it and mark it as `Failed` with the reason `PipelineRunTimeout` or `TaskRunTimeout`.

Timeouts can be set at multiple levels, with more specific settings overriding more general ones:
1.  **On the `PipelineRun`**: `spec.timeouts.pipeline` sets the deadline for the entire run. This is the highest-level override.
2.  **On the `Pipeline`**: `spec.timeouts.pipeline` sets a default for all runs of this pipeline. `spec.timeouts.tasks` sets a default for each task within the pipeline.

---
## 💼 Real-World Use Case

A pipeline runs a complex set of end-to-end tests that usually take about 15 minutes. The team wants to ensure that if the test environment is unresponsive, the pipeline doesn't hang for the default one-hour period.

They add `timeouts: { pipeline: "20m" }` to their `Pipeline` spec. This provides a 5-minute buffer over the expected time. If the tests ever take longer than 20 minutes, the `PipelineRun` automatically fails, freeing up the resources and immediately alerting the team that something is wrong with the test environment.

---
## 💻 Code/Config Example

This `PipelineRun` executes a `Pipeline` that has a step to `sleep 30`. However, the `PipelineRun` itself has a timeout of 15 seconds, which will cause it to fail first.

**`pipelinerun-with-timeout.yaml`**
```yaml
apiVersion: tekton.dev/v1beta1
kind: PipelineRun
metadata:
  name: timeout-demo-run
spec:
  pipelineSpec: # Defining the pipeline inline for simplicity
    tasks:
      - name: long-running-task
        taskSpec:
          steps:
            - name: sleep
              image: ubuntu
              script: |
                echo "Starting a task that will take 30 seconds..."
                sleep 30
                echo "This message will never be seen."
  timeouts:
    pipeline: "15s" # <-- This timeout is shorter than the task's sleep duration
```

---
## 🤔 Daily Self-Assessment

**Question:** If a `PipelineRun` specifies a timeout of `5m` and the `Pipeline` it references specifies a default timeout of `10m`, which timeout value is used for that specific run?

**Answer:** The `PipelineRun`'s timeout of `5m` takes precedence. Specifications on a `PipelineRun` always override the defaults set in the `Pipeline`.

---
## 🛠️ Practical To-do exercise

Today you will apply a timeout to a pipeline and watch it fail as expected.

1.  **Navigate and Create Directory**:
    ```bash
    cd tekton-learning
    mkdir -p week2/day14
    ```

2.  **Create a `Pipeline` with a Long-Running Task**: Create `week2/day14/pipeline.yaml`. For simplicity, we'll define the task `inline` using `taskSpec`.
    ```yaml
    # week2/day14/pipeline.yaml
    apiVersion: tekton.dev/v1beta1
    kind: Pipeline
    metadata:
      name: pipeline-with-long-task
    spec:
      tasks:
        - name: sleep-task
          taskSpec:
            steps:
              - name: sleeper
                image: ubuntu
                script: |
                  echo "Starting to sleep for 20 seconds."
                  sleep 20
                  echo "Finished sleeping."
    ```

3.  **Create the `PipelineRun` with a Timeout**: Create `week2/day14/pipelinerun.yaml`. This is where you'll set a timeout that is shorter than the 20-second sleep.
    ```yaml
    # week2/day14/pipelinerun.yaml
    apiVersion: tekton.dev/v1beta1
    kind: PipelineRun
    metadata:
      name: timeout-run-example
    spec:
      pipelineRef:
        name: pipeline-with-long-task
      timeouts:
        pipeline: "10s" # 10 seconds is less than the 20 second sleep
    ```

4.  **Apply and Run**:
    ```bash
    kubectl apply -f week2/day14/pipeline.yaml
    kubectl apply -f week2/day14/pipelinerun.yaml
    ```

5.  **Verify the Result**:
    * Watch the `PipelineRun`'s status: `kubectl get pipelinerun timeout-run-example -w`.
    * You will see its status change to `Running`. After about 10 seconds, it will abruptly change to `Failed`.
    * Now, inspect the `PipelineRun` in detail to see the reason for the failure:
        ```bash
        kubectl describe pipelinerun timeout-run-example
        ```
    * Look for the `Status` section at the end of the output. You should see a condition that says `Reason: PipelineRunTimeout`. This confirms your timeout worked correctly.

6.  **Commit Your Work**:
    ```bash
    git add .
    git commit -m "Day 14: Add and test PipelineRun timeouts"
    git push origin main
    ```
