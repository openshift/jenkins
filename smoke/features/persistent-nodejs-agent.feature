Feature: Testing jenkins agent nodejs image

  As a user of Jenkins on openshift
    I want to deploy Nodejs application on OpenShift
        

  Background:
    Given Project [TEST_NAMESPACE] is used

  @automated @customer-scenario
  Scenario: Deploy sample application on openshift : JKNS-08-TC01
    Given The jenkins pod is up and runnning
    When The user create objects from the "smoke/samples/nodejs_pipeline.yaml" template by processing the template and piping the output to oc create
    Then we check that the resources are created
      | resource         | resource_name             |
      | buildconfig      | sample-pipeline           |
      | secret           | nodejs-postgresql-example |
      | service          | nodejs-postgresql-example |
      | route            | nodejs-postgresql-example |
      | imagestream      | nodejs-postgresql-example |
      | buildconfig      | nodejs-postgresql-example |
      | deploymentconfig | nodejs-postgresql-example |
      | service          | postgresql                |
      | deploymentconfig | postgresql                |
    Then Trigger the build using "oc start-build sample-pipeline"
    And verify the build status of "sample-pipeline-1" build is Complete
    Then verify the build status of "nodejs-postgresql-example-1" build is Complete
    And route nodejs-postgresql-example must be created and be accessible