# Day 12: Securing Triggers with Interceptors
---
## 🧠 Concept of the Day

On Day 7, you created an `EventListener` that exposed an HTTP endpoint. In a real-world scenario, this endpoint would be on the public internet, meaning anyone could send a request to it and trigger your pipeline. To prevent this, Tekton Triggers has **`Interceptors`**.

An `Interceptor` is a middleware that processes a webhook payload *before* it reaches your `TriggerBinding`. You can use them to **validate**, **filter**, and **transform** incoming events. The most critical use case is validation.

Tekton provides built-in `ClusterInterceptors` for common tasks. The `github` interceptor, for example, can validate that an incoming webhook request is genuinely from GitHub by checking its signature against a shared secret token. If the signature is invalid, the interceptor rejects the request, and the pipeline is never triggered.

---
## 💼 Real-World Use Case

A team exposes an `EventListener` to automate their deployment pipeline from a GitHub `push` event.
1.  They create a webhook in their GitHub repository and configure it with a secret token (e.g., `"super-secret-123"`).
2.  They create a corresponding `Secret` in their Kubernetes cluster containing the same token.
3.  In their `EventListener`, they add a `github` `Interceptor` that references this Kubernetes `Secret`.

Now, when GitHub sends a webhook, it includes a special `X-Hub-Signature-256` header. The Tekton `Interceptor` uses the shared secret to calculate its own signature from the payload and rejects any request where the signatures do not match. This secures the endpoint from unauthorized triggers.

---
## 💻 Code/Config Example

Here's how you add an `interceptor` to an `EventListener` to validate GitHub webhooks.

First, you need a secret holding the token.
```bash
# The 'webhook-secret-token' should be a random, secure string
kubectl create secret generic github-webhook-secret --from-literal=secretToken="webhook-secret-token"
```

---
## 🤔 Daily Self-Assessment

**Question:** What is the primary purpose of using a GitHub `Interceptor` in an `EventListener`?

**Answer:** To validate that the incoming webhook is genuinely from GitHub and not from a malicious or unauthorized source. It accomplishes this by verifying the webhook's digital signature using a shared secret.

---
## 🛠️ Practical To-do exercise

Today you will secure the `EventListener` you created on Day 7. You will add an interceptor and see how it rejects requests that don't have a valid signature.

1.  **Navigate and Create Directory**:
    ```bash
    cd tekton-learning
    mkdir -p week2/day12
    ```

2.  **Create the Webhook `Secret`**: Choose a simple secret string for this exercise (in the real world, this should be very random and complex).
    ```bash
    kubectl create secret generic github-webhook-secret --from-literal=secretToken="12345"
    ```

3.  **Update the `EventListener`**: Copy your `EventListener` YAML from `week1/day7/03-event-listener.yaml` to `week2/day12/event-listener.yaml` and add the `interceptors` block.
    ```yaml
    # week2/day12/event-listener.yaml
    apiVersion: triggers.tekton.dev/v1beta1
    kind: EventListener
    metadata:
      name: github-listener-secured # New name to avoid conflicts
    spec:
      serviceAccountName: tekton-pipelines-controller
      triggers:
        - name: github-trigger
          interceptors: # <-- ADD THIS BLOCK
            - name: "validate-signature"
              ref:
                name: "github"
                kind: ClusterInterceptor
              params:
                - name: "secretRef"
                  value:
                    secretName: github-webhook-secret
                    secretKey: secretToken
          bindings:
            - ref: github-push-binding
          template:
            ref: clone-pipeline-template
    ```

4.  **Apply the New `EventListener`**:
    ```bash
    # You can delete the old one first if you like
    # kubectl delete eventlistener github-listener
    kubectl apply -f week2/day12/event-listener.yaml
    ```

5.  **Expose the `EventListener`**: In a **separate terminal**, run `kubectl port-forward` on the new, secured service.
    ```bash
    # This command will continue running. Leave this terminal open.
    kubectl port-forward service/el-github-listener-secured 8080:8080
    ```

6.  **Simulate a FAILED Webhook**: Go back to your original terminal. Use the same `curl` command and `payload.json` from Day 7.
    ```bash
    # Make sure you have the payload.json from Day 7
    curl -v \
    -H 'X-GitHub-Event: push' \
    -H 'Content-Type: application/json' \
    -d @payload.json \
    http://localhost:8080
    ```
    **Observe the result!** Instead of `202 Accepted`, you should now receive an `HTTP/1.1 400 Bad Request`. The response body will say something like `Error: missing X-Hub-Signature-256 header`. Your interceptor has successfully rejected the invalid request! This is the key lesson.

7.  **Understanding a SUCCESSFUL Webhook**: To make this request succeed, you would need to calculate a SHA256 HMAC signature of the payload and include it in the `X-Hub-Signature-256` header. A real webhook from GitHub does this automatically. For today, the important takeaway is that your endpoint is no longer open to anonymous requests.

8.  **Cleanup and Commit**:
    ```bash
    # Stop the port-forward command (Ctrl+C)
    # You can leave the secret and EventListener for future lessons

    git add .
    git commit -m "Day 12: Secure EventListener with a GitHub Interceptor"
    git push origin main
    ```

