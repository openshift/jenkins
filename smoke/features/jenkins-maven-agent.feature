Feature: Testing jenkins agent maven image

  As a user of Jenkins on openshift
    I want to deploy JavaEE application on OpenShift
        

  Background:
    Given Project [TEST_NAMESPACE] is used

  Scenario: Deploy JavaEE application on OpenShift
      Given The jenkins pod is up and runnning
      When The user create objects from the sample maven template by processing the template and piping the output to oc create
      And verify imagestream.image.openshift.io/openshift-jee-sample & imagestream.image.openshift.io/wildfly exist
      And verify buildconfig.build.openshift.io/openshift-jee-sample & buildconfig.build.openshift.io/openshift-jee-sample-docker exist
      And verify deploymentconfig.apps.openshift.io/openshift-jee-sample is created
      And verify service/openshift-jee-sample is created
      And verify route.route.openshift.io/openshift-jee-sample is created
      Then Trigger the build using oc start-build openshift-jee-sample
      Then verify the build status of openshift-jee-sample-docker build is Complete
      And verify the build status of openshift-jee-sample-1 is Complete
      And verify the JaveEE application is accessible via route openshift-jee-sample