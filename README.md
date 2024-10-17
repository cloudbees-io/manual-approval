# manual-approval
Request manual approval from users and teams

Use this job  

Refer to the link:https://docs.cloudbees.com/docs/cloudbees-platform/latest/workflows/runs#_workflow_run_evidence[run evidence documentation] for more information.

== Inputs

[cols="2a,1a,1a,3a",options="header"]
.Input details
|===

| Input name
| Data type
| Required?
| Description

| `delegates`
|String
| Yes
| 

| `approvers`
| String
|No
| 

| `disallowLaunchByUser`
|String
| Yes
| 

| `instruction`
|String
| Yes
| 

|===

== Usage example

In your YAML file, add:

[source,yaml]
----
      - name: Publish workflow evidence item
        uses: cloudbees-io/publish-evidence-item@v1
        with:
          content: |
            ## Test markup and property rendering
            - Run ID: ${{ cloudbees.run_id }}
            - [backend.tar](https://ourcompany.com/repo/backend.tar)       
----

NOTE: For more information 

== License

This code is made available under the 
link:https://opensource.org/license/mit/[MIT license].

== References

* Learn more about link:https://docs.cloudbees.com/docs/cloudbees-platform/latest/actions[using actions in CloudBees workflows].
* Learn about link:https://docs.cloudbees.com/docs/cloudbees-platform/latest/[the CloudBees platform].
