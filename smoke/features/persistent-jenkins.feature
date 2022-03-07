Feature: Deploy Jenkins with persistent volume on openshift using template based install

    As a user of openshift
    I want to deploy Jenkins with persistent volume and configure my CI/CD on openshift cluster

    Background:
    Given Project [TEST_NAMESPACE] is used

    @automated @customer-scenario
    Scenario: Create jenkins  using persistent template : JKNS-06-TC01
        Given we have a openshift cluster
        When User enters oc new-app jenkins-persistent command
        Then we check that the resources are created
        | resource              | resource_name             |
        | route                 | jenkins                   |
        | configmap             | jenkins-trusted-ca-bundle |
        | persistentvolumeclaim | jenkins                   |
        | deploymentconfig      | jenkins                   |
        | serviceaccount        | jenkins                   |
        | rolebinding           | jenkins_edit              |
        | service               | jenkins-jnlp              |
        | service               | jenkins                   |
        Then we check the pvc status is "Bound"
        Then We check for deployment pod status to be "Completed"
        And We check for jenkins master pod status to be "Ready"