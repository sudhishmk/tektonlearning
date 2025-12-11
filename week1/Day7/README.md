# Day 7: Automating Runs with Tekton Triggers
---
## 🧠 Concept of the Day

**Tekton Triggers** is a companion project to Tekton Pipelines that allows you to instantiate `PipelineRuns` (and other resources) based on events. It acts as the bridge between external events (like a `git push`) and your CI/CD workflows.

The core components are:
* **`EventListener`**: A Kubernetes `Service` that exposes an HTTP endpoint to receive events from sources like GitHub, GitLab, or any system that can send a webhook.
* **`TriggerBinding`**: Responsible for extracting specific fields from the event's JSON payload (e.g., the repository URL or the commit SHA). It extracts this data using JSONPath notation.
* **`TriggerTemplate`**: A blueprint for a resource that will be created when the event is received. This is most often a `PipelineRun`. It uses the values extracted by the `TriggerBinding` to parameterize the resource.
* **`Trigger`**: The glue that connects an `EventListener`, a `TriggerBinding`, and a `TriggerTemplate`. It specifies which binding and template to use for a given event.

The flow is: **Webhook -> `EventListener` -> `Trigger` -> `TriggerBinding` extracts data -> `TriggerTemplate` creates a `PipelineRun`**.

---
## 💼 Real-World Use Case

The most common use case is GitOps-driven CI. A developer pushes a commit to a pull request on GitHub. GitHub is configured to send a `push` event webhook to a Tekton `EventListener`'s public URL. The `TriggerBinding` extracts the repository URL and the specific commit SHA from the webhook payload. The `TriggerTemplate` then uses these values to create a `PipelineRun` for the main CI pipeline, which clones that *exact* commit, runs tests, and builds an image.

---
## 💻 Code/Config Example

Here is a simplified set of resources to create a `PipelineRun` from a webhook.

**`trigger-binding.yaml`**
```yaml
apiVersion: triggers.tekton.dev/v1beta1
kind: TriggerBinding
metadata:
  name: github-binding
spec:
  params:
    - name: git-repo-url
      value: $(body.repository.clone_url) # Extracts the repo URL from the JSON body
```

---
## 🤔 Daily Self-Assessment

**Question:** In the Tekton Triggers model, which component is responsible for parsing the incoming JSON payload from a webhook and extracting specific fields like the committer's username?

**Answer:** The `TriggerBinding`. It uses JSONPath expressions to extract values from the event body.

---
## 🛠️ Practical To-do exercise

Today's setup is more involved as we're installing a new component and simulating a webhook.

1.  **Install Tekton Triggers**: Apply the latest release YAML to your cluster.
    ```bash
    kubectl apply --filename [https://storage.googleapis.com/tekton-releases/triggers/latest/release.yaml](https://storage.googleapis.com/tekton-releases/triggers/latest/release.yaml)
    kubectl apply --filename [https://storage.googleapis.com/tekton-releases/triggers/latest/interceptors.yaml](https://storage.googleapis.com/tekton-releases/triggers/latest/interceptors.yaml)
    ```
    Wait for the pods to be `Running` in the `tekton-pipelines` namespace.

2.  **Navigate and Create Directory**:
    ```bash
    cd tekton-learning
    mkdir -p week1/day7
    ```

3.  **Create Trigger Resources**: Create the following three files.
    * **`week1/day7/01-trigger-binding.yaml`**:
        ```yaml
        apiVersion: triggers.tekton.dev/v1beta1
        kind: TriggerBinding
        metadata:
          name: github-push-binding
        spec:
          params:
            - name: git-repo-url
              value: $(body.repository.clone_url)
        ```
    * **`week1/day7/02-trigger-template.yaml`**: This templates the `PipelineRun` from Day 6.
        ```yaml
        apiVersion: triggers.tekton.dev/v1beta1
        kind: TriggerTemplate
        metadata:
          name: clone-pipeline-template
        spec:
          params:
            - name: git-repo-url
          resourcetemplates:
            - apiVersion: tekton.dev/v1beta1
              kind: PipelineRun
              metadata:
                generateName: clone-bundle-run-
              spec:
                pipelineRef:
                  name: git-clone-bundle-pipeline # Pipeline from Day 6
                params:
                  - name: repo-url-from-pipeline
                    value: $(tt.params.git-repo-url)
                workspaces:
                - name: shared-data
                  persistentVolumeClaim:
                    claimName: tekton-shared-workspace
        ```
    * **`week1/day7/03-event-listener.yaml`**: This ties everything together.
        ```yaml
        apiVersion: triggers.tekton.dev/v1beta1
        kind: EventListener
        metadata:
          name: github-listener
        spec:
          serviceAccountName: tekton-pipelines-controller # Default SA
          triggers:
            - name: github-trigger
              bindings:
                - ref: github-push-binding
              template:
                ref: clone-pipeline-template
        ```

4.  **Apply the Resources**:
    ```bash
    kubectl apply -f week1/day7/
    ```

5.  **Expose the EventListener**: In a **separate terminal**, run `kubectl port-forward`. This will make the listener accessible on your local machine.
    ```bash
    # This command will continue running. Leave this terminal open.
    kubectl port-forward service/el-github-listener 8080:8080 -n default
    ```

6.  **Simulate the Webhook**: Go back to your original terminal. We'll use `curl` to send a fake GitHub push event to the listener.
    * Create a file `payload.json` with this content, **changing `YOUR_USERNAME`**:
        ```json
        {
          "repository": {
            "clone_url": "[https://github.com/YOUR_USERNAME/tekton-learning.git](https://github.com/YOUR_USERNAME/tekton-learning.git)"
          }
        }
        ```
    * Now, `POST` this payload to your listener:
        ```bash
        curl -v \
        -H 'X-GitHub-Event: push' \
        -H 'Content-Type: application/json' \
        -d @payload.json \
        http://localhost:8080
        ```
    * You should get an `HTTP/1.1 202 Accepted` response.

7.  **Verify the Result**: Check for a new `PipelineRun` that was created automatically!
    ```bash
    kubectl get pipelineruns | grep clone-bundle-run
    ```
    You have successfully triggered a pipeline from an external event! You can now stop the `kubectl port-forward` command.

8.  **Commit Your Work**:
    ```bash
    git add .
    git commit -m "Day 7: Add Tekton Triggers to automate PipelineRuns"
    git push origin main
    ```

[Go to Week 1 Project](../project/README.md)


