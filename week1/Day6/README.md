# Day 6: Don't Reinvent the Wheel - Artifact Hub & Bundles
---
## 🧠 Concept of the Day

Writing every `Task` from scratch is inefficient. **Artifact Hub** is the central, vendor-neutral registry for finding, installing, and publishing cloud-native packages, including official and community-verified Tekton `Tasks`.

Instead of applying task YAML files manually, the modern approach is to consume them as **OCI bundles**. A bundle is a standard container image that packages the Tekton resource definitions. You reference a task from a bundle directly in your `Pipeline` using the `taskRef.bundle` field. This field points to the full OCI image path of the task, which you can find on Artifact Hub.

This method keeps your `Pipeline` definitions self-contained and ensures you are pulling versioned, trusted components without needing to store them in your own source code.

---
## 💼 Real-World Use Case

A platform team is building a standard deployment `Pipeline`. They need to send a Slack notification on success or failure. Instead of writing a custom `Task` to interact with the Slack API, they search Artifact Hub and find the official `slack-post-message` `Task`. They reference its OCI bundle in their `Pipeline` and provide the necessary `params` (like a secret webhook URL and the message). This saves development time and ensures they're using a well-maintained component.

---
## 💻 Code/Config Example

Here's how you would reference the official `git-clone` task from its OCI bundle on Artifact Hub. Notice the `bundle` field pointing to the container image path and tag.

```yaml
# In a Pipeline definition...
tasks:
  - name: fetch-source-from-hub
    taskRef:
      name: git-clone # The name of the Task *inside* the bundle
      bundle: gcr.io/tekton-releases/catalog/upstream/git-clone:0.9 # The OCI image path from Artifact Hub
    # You still provide workspaces and params, but their names
    # must match what the Task in the bundle expects.
    workspaces:
      - name: output # The git-clone Hub task calls its workspace 'output'
        workspace: shared-data
    params:
      - name: url # The git-clone Hub task calls its repo param 'url'
        value: $(params.repo-url)
```


