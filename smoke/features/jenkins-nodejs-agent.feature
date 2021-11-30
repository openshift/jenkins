Feature: Testing jenkins agent nodejs image

  As a user of Jenkins on openshift
    I want to deploy Nodejs application on OpenShift
        

  Background:
    Given Project [TEST_NAMESPACE] is used

  Scenario: Deploy sample application on openshift
    Given The jenkins pod is up and runnning
    When The user enters new-app command with nodejs_template
    Then Trigger the build using oc start-build
    Then verify the build status of "nodejs-postgresql-example-1" build is Complete
    Then verify the build status of "nodejs-postgresql-example-2" build is Complete
    Then We check for deployment pod status to be "Completed"
    And route nodejs-postgresql-example must be created and be accessible