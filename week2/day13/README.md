# Day 13: Advanced Triggering with CEL Interceptors
---
## 🧠 Concept of the Day

The `github` interceptor is great for verifying a webhook's authenticity, but often you need more granular control. For example, you might only want to trigger a pipeline for pushes to the `main` branch, or if a commit message contains `[deploy-prod]`.

For this, Tekton provides the **`cel` Interceptor**. CEL, or **Common Expression Language**, is a simple, fast, and secure expression language that you can embed directly in your `EventListener` YAML.

The `cel` interceptor allows you to write a `filter` expression that evaluates the incoming JSON payload.
* If the expression evaluates to `true`, the event is passed to the next interceptor or to the `TriggerBinding`.
* If it evaluates to `false`, the event is silently and safely discarded, and no `PipelineRun` is created.

You can chain interceptors, using `github` for authentication first, followed by `cel` for fine-grained filtering.

---
## 💼 Real-World Use Case

A company practices "ChatOps." A developer can trigger a deployment to the staging environment by leaving a comment on a GitHub pull request that says `/deploy-staging`.
1. An `EventListener` receives all pull request comment webhooks.
2. A `github` interceptor first validates the webhook's authenticity.
3. A `cel` interceptor runs next. Its `filter` is: `body.action == 'created' && body.comment.body.startsWith('/deploy-staging')`.
4. If a comment is `/deploy-staging`, the filter passes, and a `TriggerBinding` extracts the PR number to create a `PipelineRun`. Any other comment is ignored.

---
## 💻 Code/Config Example

Here's how to chain the `github` interceptor with a `cel` interceptor that only allows events for the `main` branch.

**`event-listener.yaml`**
```yaml
apiVersion: triggers.tekton.dev/v1beta1
kind: EventListener
# ... metadata ...
spec:
  triggers:
    - name: github-main-branch-trigger
      interceptors:
        # First, authenticate the webhook
        - name: "validate-github"
          ref:
            name: "github"
            kind: ClusterInterceptor
          params:
            - name: "secretRef"
              value:
                secretName: github-webhook-secret
                secretKey: secretToken

        # Second, filter by branch name
        - name: "filter-on-main-branch"
          ref:
            name: "cel"
            kind: ClusterInterceptor
          params:
            - name: "filter"
              value: "body.ref == 'refs/heads/main'" # <-- The CEL expression
      bindings:
        - ref: github-push-binding
      template:
        ref: clone-pipeline-template
```

---
## 🤔 Daily Self-Assessment

**Question:** In an `EventListener`, if you have multiple interceptors (e.g., `github`, `cel`), in what order do they execute?

**Answer:** They execute sequentially in the order they are defined in the `interceptors` array. If any interceptor in the chain rejects or filters out the event, the processing stops, and subsequent interceptors and the `TriggerBinding` are not executed.

---
## 🛠️ Practical To-do exercise

Today you'll enhance the secure `EventListener` from Day 12, adding a `cel` interceptor to ensure it only triggers `PipelineRuns` for pushes to your `main` branch.

1.  **Navigate and Create Directory**:
    ```bash
    cd tekton-learning
    mkdir -p week2/day13
    ```

2.  **Update the `EventListener`**: Copy your `EventListener` from `week2/day12/event-listener.yaml` to `week2/day13/event-listener.yaml` and add the `cel` interceptor.

    ```yaml
    # week2/day13/event-listener.yaml
    apiVersion: triggers.tekton.dev/v1beta1
    kind: EventListener
    metadata:
      name: github-listener-filtered # New name
    spec:
      serviceAccountName: tekton-pipelines-controller
      triggers:
        - name: github-trigger
          interceptors:
            # 1. First interceptor authenticates
            - name: "validate-signature"
              ref:
                name: "github"
                kind: ClusterInterceptor
              params:
                - name: "secretRef"
                  value:
                    secretName: github-webhook-secret
                    secretKey: secretToken

            # 2. Second interceptor filters the payload
            - name: "filter-main-branch"
              ref:
                name: "cel"
                kind: ClusterInterceptor
              params:
                - name: "filter"
                  value: "body.ref == 'refs/heads/main'"
          bindings:
            - ref: github-push-binding
          template:
            ref: clone-pipeline-template
    ```

3.  **Apply the New `EventListener`**:
    ```bash
    # You can delete the old one first
    # kubectl delete eventlistener github-listener-secured
    kubectl apply -f week2/day13/event-listener.yaml
    ```

4.  **Expose the `EventListener`**: In a **separate terminal**, run `port-forward` for the new listener.
    ```bash
    kubectl port-forward service/el-github-listener-filtered 8080:8080
    ```

5.  **Simulate a FILTERED Webhook**: Let's send a payload for a feature branch.
    * Create a file `payload-feature-branch.json` with this content, **changing `YOUR_USERNAME`**:
        ```json
        {
          "ref": "refs/heads/a-new-feature",
          "repository": {
            "clone_url": "[https://github.com/YOUR_USERNAME/tekton-learning.git](https://github.com/YOUR_USERNAME/tekton-learning.git)"
          }
        }
        ```
    * Now, `POST` this payload.
        ```bash
        curl -v \
        -H 'X-GitHub-Event: push' \
        -H 'Content-Type: application/json' \
        -d @payload-feature-branch.json \
        http://localhost:8080
        ```
    * **Observe the result!** You should get an `HTTP/1.1 204 No Content` response. This means the request was received, but the `cel` interceptor evaluated the filter to `false` and stopped processing. No `PipelineRun` was created.

6.  **Simulate a VALID Webhook (for the filter)**: Now, let's send a payload for the `main` branch.
    * Create a file `payload-main-branch.json` with this content, **changing `YOUR_USERNAME`**:
        ```json
        {
          "ref": "refs/heads/main",
          "repository": {
            "clone_url": "[https://github.com/YOUR_USERNAME/tekton-learning.git](https://github.com/YOUR_USERNAME/tekton-learning.git)"
          }
        }
        ```
    * `POST` this payload.
        ```bash
        curl -v \
        -H 'X-GitHub-Event: push' \
        -H 'Content-Type: application/json' \
        -d @payload-main-branch.json \
        http://localhost:8080
        ```
    * You will still get a `400 Bad Request` because our `github` interceptor is still correctly blocking the request due to the missing signature. But this test proves that if the request *was* valid, the `cel` filter would have evaluated to `true` and allowed it to proceed.

7.  **Commit Your Work**:
    ```bash
    # Stop the port-forward command (Ctrl+C)
    git add .
    git commit -m "Day 13: Add CEL Interceptor to filter triggers"
    git push origin main
    ```

[Go to Day 14](../day14/README.md)
