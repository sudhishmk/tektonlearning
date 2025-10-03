# Day 8: Smart Pipelines - Conditional Execution with `WhenExpressions`
---
## 🧠 Concept of the Day

Not every task in a pipeline needs to run every single time. You might want to deploy only from the `main` branch or run an extra security scan only if a specific file has changed. For this, Tekton provides **`WhenExpressions`**.

A `WhenExpression` is a condition you can add to a `Task` within a `Pipeline`. If the condition evaluates to `true`, the `Task` runs. If it evaluates to `false`, the `Task` is gracefully **skipped**, and the pipeline continues (assuming other tasks don't depend on its results).

This allows you to build powerful, flexible pipelines that can change their behavior based on input `Parameters` or the `Results` of previous tasks. The expression checks an `input` value against an `operator` (like `in` or `notin`) and a list of `values`.

---
## 💼 Real-World Use Case

A common CI/CD workflow is to run unit tests on every commit, but only build and push a container image when commits are merged into the `main` branch.

A `Pipeline` would be triggered for every commit, passing the branch name as a `Parameter`.
1.  **`unit-test-task`**: Runs on every execution.
2.  **`build-and-push-task`**: This task has a `WhenExpression` that checks if the branch parameter is `main`.
    `when: [{input: "$(params.git-branch)", operator: in, values: ["main"]}]`

For a commit to a feature branch, the test task runs and the build task is skipped. For a merge to `main`, both tasks run.

---
## 💻 Code/Config Example

This `Pipeline` has a task that only runs if a boolean-like parameter is set to `"true"`.

**`pipeline.yaml`**
```yaml
apiVersion: tekton.dev/v1beta1
kind: Pipeline
metadata:
  name: conditional-pipeline
spec:
  params:
    - name: run-optional-task
      type: string
      default: "false"

  tasks:
    - name: always-run-this
      taskRef:
        name: echo-message
      params:
        - name: message
          value: "This is the first task, it always runs."

    - name: optional-task
      runAfter: ["always-run-this"]
      when:
        - input: "$(params.run-optional-task)"
          operator: in
          values: ["true"]
      taskRef:
        name: echo-message
      params:
        - name: message
          value: "This is the second task, it only runs if the condition is met."
```

---
## 🤔 Daily Self-Assessment

**Question:** In a `Pipeline`, if a `Task` with a `WhenExpression` is skipped, is the overall `PipelineRun` considered "Failed"?

**Answer:** No. The `PipelineRun` is considered "Succeeded". A skipped task is a valid, expected outcome, not an error.

---
## 🛠️ Practical To-do exercise

Today, you'll enhance your Week 1 project. You'll modify the `go-test` task to produce a `result` and add a new `Task` that only runs if the tests passed, based on that result.

1.  **Navigate and Create Directory**:
    ```bash
    cd tekton-learning
    mkdir -p week2/day8
    ```

2.  **Enhance the `go-test` Task**: Copy your `go-test-task.yaml` from `week1/project` into the new `week2/day8` directory. Now, modify it to output a `result`.

    ```yaml
    # week2/day8/go-test-task.yaml
    apiVersion: tekton.dev/v1beta1
    kind: Task
    metadata:
      name: go-test-with-results # Give it a new name
    spec:
      workspaces:
        - name: source
      results: # <-- Add this results section
        - name: status
          description: "The status of the tests (passed/failed)."
      steps:
        - name: test
          image: golang:1.22
          script: |
            #!/bin/sh
            set -e
            cd $(workspaces.source.path)/week1/project
            go test -v ./...
            # If the test command above succeeds, we write "passed"
            echo -n "passed" > $(results.status.path)
    ```

3.  **Create the Conditional `Pipeline`**: Create `week2/day8/pipeline.yaml`. This pipeline will use the new test task.
    ```yaml
    # week2/day8/pipeline.yaml
    apiVersion: tekton.dev/v1beta1
    kind: Pipeline
    metadata:
      name: ci-pipeline-with-conditions
    spec:
      params:
        - name: repo-url
          type: string
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
            name: go-test-with-results # Reference our new task
          workspaces:
            - name: source
              workspace: shared-workspace

        - name: report-success
          runAfter: ["run-go-tests"]
          when: # <-- The conditional part!
            - input: "$(tasks.run-go-tests.results.status)"
              operator: in
              values: ["passed"]
          taskRef:
            name: echo-message # Using the generic echo task from last week
          params:
            - name: message
              value: "✅ All tests passed successfully!"
    ```

4.  **Create the `PipelineRun`**: Create `week2/day8/pipelinerun.yaml`.
    ```yaml
    # week2/day8/pipelinerun.yaml
    apiVersion: tekton.dev/v1beta1
    kind: PipelineRun
    metadata:
      name: conditional-ci-run-1
    spec:
      pipelineRef:
        name: ci-pipeline-with-conditions
      params:
        - name: repo-url
          value: [https://github.com/YOUR_USERNAME/tekton-learning.git](https://github.com/YOUR_USERNAME/tekton-learning.git) # <-- IMPORTANT: CHANGE THIS
      workspaces:
        - name: shared-workspace
          persistentVolumeClaim:
            claimName: tekton-shared-workspace
    ```
    **Remember to replace `YOUR_USERNAME`.**

5.  **Apply and Run**:
    ```bash
    # Apply your new task and pipeline
    kubectl apply -f week2/day8/go-test-task.yaml
    kubectl apply -f week2/day8/pipeline.yaml

    # Run the pipeline
    kubectl apply -f week2/day8/pipelinerun.yaml
    ```

6.  **Verify the Result**:
    * Watch the run: `kubectl get pipelinerun conditional-ci-run-1 -w`.
    * You'll see the `clone` and `test` tasks run, followed by the `report-success` task.
    * Check the logs for the `report-success` pod. It should print `"✅ All tests passed successfully!"`. Because the condition was met, the task ran.

7.  **Commit Your Work**:
    ```bash
    git add .
    git commit -m "Day 8: Add WhenExpressions for conditional execution"
    git push origin main
    ```
