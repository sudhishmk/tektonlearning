# 🗓️ Weekly Review: Week 2
---

This week, you moved from fundamentals to advanced, production-level concepts. You've built a powerful toolkit for creating robust, secure, and flexible pipelines.

**Summary of Topics:**
* **Day 8:** Controlling task execution with conditional `WhenExpressions`.
* **Day 9:** Fanning out tasks for parallel execution using `Matrix`.
* **Day 10:** Guaranteeing cleanup and notifications with `finally` tasks.
* **Day 11:** Building container images with Kaniko and securely managing credentials with `Secrets` and `ServiceAccounts`.
* **Day 12:** Securing `EventListeners` by validating webhooks with `github` interceptors.
* **Day 13:** Creating fine-grained trigger rules with `cel` interceptor filters.
* **Day 14:** Preventing stuck builds and controlling execution duration with `Timeouts`.

### 🎉 Weekly Mini-Project: The "Production-Ready" CI/CD Pipeline

Your goal is to combine the key concepts from this week into a single, advanced pipeline.

1.  **Start with the `test-and-build-pipeline`** from Day 11.
2.  **Add a "Deploy" Stage**:
    * Add a final `Task` to the pipeline named `deploy-to-prod-and-staging`.
    * This `Task` should use the `echo-message` task and a `matrix` to simulate parallel deployments to `staging` and `production`.
3.  **Add a Conditional Deployment**:
    * Add a `param` to your `Pipeline` called `is-production-deploy` (type `string`, default `"false"`).
    * Modify the `deploy-to-prod-and-staging` task. It should no longer be a single matrix. Instead, create two separate tasks:
        * A `deploy-to-staging` task that always runs after the build.
        * A `deploy-to-production` task that also runs after the build, but **only when** the `is-production-deploy` param is `"true"`. Use a `WhenExpression` for this.
4.  **Add a `finally` block**:
    * Add a `finally` task named `report-status`.
    * It should use the `echo-message` task to print the final status: `Pipeline finished with status: $(tasks.status)`.
5.  **Add a Timeout**:
    * In your `PipelineRun`, set a reasonable overall timeout for the pipeline, for example, `10m`.
6.  **Run it twice**:
    * First, run the `PipelineRun` with `is-production-deploy` left as the default (`"false"`). Observe that the staging deployment runs and the production deployment is skipped.
    * Second, create a new `PipelineRun` where you explicitly set the `is-production-deploy` param to `"true"`. Observe that both deployment tasks now run.

Completing this project demonstrates mastery over the advanced control flow and security features of Tekton.


[Go to Troubleshooting Guide](../../TroubleshootingGuide.md)
