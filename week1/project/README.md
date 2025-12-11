# 🗓️ Weekly Review: Week 1
---

Amazing work! You've gone from zero to a fully automated pipeline this week, covering the most critical concepts in the Tekton ecosystem.

**Summary of Topics:**
* **Day 1:** Core units of work, `Task` and `TaskRun`.
* **Day 2:** Sharing files between `Steps` with `Workspaces`.
* **Day 3:** Making `Tasks` reusable with `Parameters`.
* **Day 4:** Orchestrating `Tasks` into workflows with `Pipelines`.
* **Day 5:** Passing small outputs between `Tasks` with `Results`.
* **Day 6:** Using reusable `Tasks` from Artifact Hub and enabling features with `feature-flags`.
* **Day 7:** Automating pipeline execution from webhooks with **Tekton Triggers**.

### 🎉 Weekly Mini-Project: A Simple Go CI Pipeline

It's time to integrate everything. Your project is to build a basic CI pipeline for a "Hello World" Go application.

**1. Create the Application Files**

* Create a new directory: `mkdir -p week1/project`.
* Create a simple Go application file `week1/project/main.go`:
    ```go
    package main
    import "fmt"
    func main() { fmt.Println(Message()) }
    func Message() string { return "Hello, CI!" }
    ```
* Create a test file `week1/project/main_test.go`:
    ```go
    package main
    import "testing"
    func TestMessage(t *testing.T) {
        if Message() != "Hello, CI!" {
            t.Errorf("Message() = %s; want Hello, CI!", Message())
        }
    }
    ```
* Commit and push these new files to your `tekton-learning` repository.

**2. Create the CI `Pipeline` and `Task`**

* **Create a `go-test` Task**: Write a `Task` definition in `week1/project/go-test-task.yaml`.
    * It should accept a `workspace` named `source`.
    * It should have one `step` using a `golang:1.22` image.
    * The `script` should `cd` into the source workspace and run `go test -v ./...`.
* **Create the `Pipeline`**: Write a `Pipeline` in `week1/project/pipeline.yaml`.
    * It should have two `Tasks`. The first should use the `git-clone` bundle from Artifact Hub. The second should be your `go-test` task, running `after` the clone and using the same workspace.
* **Create the `PipelineRun`**: Write a `PipelineRun` in `week1/project/pipelinerun.yaml` to execute your new pipeline manually and verify it works.

Apply and run everything. When it succeeds, you'll have an end-to-end CI pipeline ready for automation!

[Go to Day 8](../../week2/day8/README.md)
