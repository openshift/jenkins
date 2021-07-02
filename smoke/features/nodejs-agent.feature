Feature: Testing jenkins agent nodejs image

  As a user of Jenkins Operator
    I want to deploy Nodejs application on OpenShift
        

  Background:
    Given Project [TEST_NAMESPACE] is used

  Scenario: Deploy sample application on openshift
    Given The jenkins pod is up and runnning
    When The user enters new-app command with nodejs_template
    Then Trigger the build using oc start-build
    Then nodejs-postgresql-example pod must come up
    And route nodejs-postgresql-example must be created and be accessible