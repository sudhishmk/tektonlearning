# Troubleshooting guide for Tekton 

***

## Collecting Logs

Logs are the most critical source of information. They can be categorized into logs from your pipeline runs and logs from the Tekton controllers themselves.

### **Controller Logs (Platform-Level)**

If the entire system is behaving unexpectedly (e.g., no `PipelineRuns` are starting), check the controller logs. These live in the `$tekton-pipelines$` namespace by default.

* **Tekton Pipelines Controller:** Manages `PipelineRuns`, `TaskRuns`, etc.
    ```bash
    kubectl logs -l app=tekton-pipelines-controller -n tekton-pipelines
    ```

* **Tekton Triggers Controller:** Manages `EventListeners`, `TriggerTemplates`, etc.
    ```bash
    kubectl logs -l app=tekton-triggers-controller -n tekton-triggers
    ```

* **Tekton Dashboard (if installed):**
    ```bash
    kubectl logs -l app=tekton-dashboard -n tekton-pipelines
    ```

### **PipelineRun & TaskRun Logs**

Use the Tekton CLI (`tkn`) for quick access to logs from specific runs.

* **Get logs for the last PipelineRun:**
    ```bash
    tkn pipelinerun logs --last -n <namespace>
    ```

* **Get logs for a specific TaskRun:**
    ```bash
    tkn taskrun logs <taskrun-name> -n <namespace>
    ```

### **Pod Logs**

If `tkn` is unavailable, get logs directly from the Pods created by a `TaskRun`.

1.  **Find the Pod for your TaskRun:**
    ```bash
    kubectl get pods -l tekton.dev/taskRun=<taskrun-name> -n <namespace>
    ```
2.  **Get logs from a specific step (container) in the Pod:**
    ```bash
    kubectl logs <pod-name> -c step-<step-name> -n <namespace>
    ```

***

## Inspecting Tekton Resources

Describing resources shows their status, events, and configuration, which is essential for debugging.

* **Describe a PipelineRun:** This is the most useful command for most failures.
    ```bash
    kubectl describe pipelinerun <pipelinerun-name> -n <namespace>
    ```
    
* **Describe a TaskRun:** Drill down to the specific failing `TaskRun`.
    ```bash
    kubectl describe taskrun <taskrun-name> -n <namespace>
    ```
* **Describe a Pod:** Reveals container-level issues like image pull errors.
    ```bash
    kubectl describe pod <pod-name-from-taskrun> -n <namespace>
    ```

***

## Troubleshooting Tekton Triggers

When a Git hook or event doesn't create a `PipelineRun`, follow these steps.

1.  **Check the EventListener Pod:** Triggers create a dedicated `Service` and `Deployment` for the `EventListener`. All event processing happens here.
    * **Find the EventListener pod:**
        ```bash
        kubectl get pods -l eventlistener=<eventlistener-name> -n <namespace>
        ```
    * **Check its logs:** This is the most important step. The logs will show incoming requests, any processing errors, and whether it successfully created the `PipelineRun`.
        ```bash
        kubectl logs <eventlistener-pod-name> -n <namespace>
        ```

2.  **Describe the Trigger Resources:** Check for configuration errors or status issues.
    ```bash
    # Check the main listener
    kubectl describe eventlistener <el-name> -n <namespace>

    # Check the template for creating resources
    kubectl describe triggertemplate <tt-name> -n <namespace>

    # Check the binding that extracts parameters from the event payload
    kubectl describe triggerbinding <tb-name> -n <namespace>
    ```

3.  **Common Trigger Issues:**
    * **Event Not Received:** Check the `EventListener` pod logs. If there are no log entries when you send an event, the problem is likely network-related (e.g., incorrect webhook URL, firewall, or issue with the `Service`/`Ingress`).
    * **Template Instantiation Failed:** The `EventListener` log will show an error. This usually means a parameter from the `TriggerBinding` doesn't match what the `TriggerTemplate` expects, or there's a syntax error in the template.
    * **RBAC/Permission Errors:** The `EventListener` log may show that the `ServiceAccount` associated with it doesn't have permission to create a `PipelineRun`. Check the `ServiceAccount` and its `Roles`/`RoleBindings`.

***

## Checking Configuration in ConfigMaps

Tekton's behavior can be customized globally via `ConfigMaps` in the `$tekton-pipelines$` namespace. Incorrect settings here can cause widespread issues.

* **List important ConfigMaps:**
    ```bash
    kubectl get configmaps -n tekton-pipelines
    ```
* **Check the defaults ConfigMap:** This `ConfigMap` sets system-wide defaults, like the default `ServiceAccount`, `timeout`, and `podTemplate`.
    ```bash
    kubectl describe configmap config-defaults -n tekton-pipelines
    ```
* **Check the feature-flags ConfigMap:** This enables or disables alpha/beta features. An incorrect setting here could cause unexpected behavior.
    ```bash
    kubectl describe configmap feature-flags -n tekton-pipelines
    ```

***

## Common Errors and Warnings

| Error / Status | Common Cause | Diagnostic Steps |
| :--- | :--- | :--- |
| **`ImagePullBackOff`** | Node cannot pull a container image due to an incorrect name, tag, or private registry auth failure. | 1. `$kubectl describe pod <pod-name> -n <namespace>$`.<br>2. Check the "Events" section.<br>3. Verify the image name/tag in your `Task`.<br>4. Ensure `ImagePullSecrets` are correct. |
| **`CreateContainerConfigError`** | Pod creation failed due to a missing `ConfigMap` or `Secret` it's trying to mount. | 1. `$kubectl describe pod <pod-name> -n <namespace>$`.<br>2. Look for "secret... not found" messages in Events.<br>3. Verify the referenced resources exist. |
| **`PipelineRun "failed" immediately`** | Validation error due to missing parameters, incorrect resource references, or syntax errors in the YAML. | 1. `$kubectl describe pipelinerun <pr-name> -n <namespace>$`.<br>2. The `Message` in the `Status` condition will explain the reason. |
| **`TaskRun` is stuck in `Pending`** | Pod cannot be scheduled due to insufficient cluster resources (CPU/memory) or node affinity/taint issues. | 1. `$kubectl describe pod <pod-name> -n <namespace>$`.<br>2. The Events section will show why scheduling failed (e.g., "Insufficient cpu"). |
