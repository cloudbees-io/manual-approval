# manual-approval
Request manual approval from users and teams

Use this job to requests workflow execution approval. 

Refer to the link:https://docs.cloudbees.com/docs/cloudbees-platform/latest/workflows/runs#_workflow_run_evidence[run evidence documentation] for more information.

== Inputs

[cols="2a,1a,1a,3a",options="header"]
.Input details
|===

| Input name
| Data type
| Required?
| Description

| `approvers`
| String
|No
| List of user IDs that are to be notified for approval. If left blank, all eligible users will be notified.

| `delegates`
|String
| Yes
| Path to this repository: cloudbees-io/manual-approval/custom-job.yml@v1


| `disallowLaunchByUser`
|String
| Yes
| If true, the the user that started the workflow is not allowed to perform the approval.

| `instruction`
|String
| Yes
| Text that will be shown in the runtime approval dialog, log and evidence views.

|===

== Usage example

In your YAML file, add:

[source,yaml]
----
 <approval-name>:
    delegates: cloudbees-io/manual-approval/custom-job.yml@v1
    with:
      approvers: <approver-names>
      disallowLaunchByUser: false
      instruction: <Approval instructions>  
----

NOTE: For more information 

== License

This code is made available under the 
link:https://opensource.org/license/mit/[MIT license].

== References

* Learn more about link:https://docs.cloudbees.com/docs/cloudbees-platform/latest/actions[using actions in CloudBees workflows].
* Learn about link:https://docs.cloudbees.com/docs/cloudbees-platform/latest/[the CloudBees platform].
