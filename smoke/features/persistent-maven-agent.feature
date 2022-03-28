Feature: Testing jenkins agent maven image

  As a user of Jenkins on openshift
    I want to deploy JavaEE application on OpenShift
        

  Background:
    Given Project [TEST_NAMESPACE] is used

  @automated @customer-scenario
  Scenario: Deploy JavaEE application on OpenShift : JKNS-07-TC01
      Given The jenkins pod is up and runnning
      When The user create objects from the sample maven template by processing the template and piping the output to oc create
      Then we check that the resources are created
        | resource         | resource_name               |
        | imagestream      | openshift-jee-sample        |
        | imagestream      | wildfly                     |
        | buildconfig      | openshift-jee-sample        |
        | buildconfig      | openshift-jee-sample-docker |
        | deploymentconfig | openshift-jee-sample        |
        | service          | openshift-jee-sample        |
        | route            | openshift-jee-sample        |
      Then Trigger the build using oc start-build openshift-jee-sample
      Then verify the build status of openshift-jee-sample-1 is Complete
      And verify the build status of openshift-jee-sample-docker build is Complete
      And verify the JaveEE application is accessible via route openshift-jee-sample