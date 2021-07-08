Feature: Deploy Jenkins with persistent volume on openshift using template based install

    As a user of openshift
    I want to deploy Jenkins with persistent volume and configure my CI/CD on openshift cluster

    Background:
    Given Project [TEST_NAMESPACE] is used

    Scenario: Create jenkins  using persistent template
        Given we have a openshift cluster
        When User enters oc new-app jenkins-persistent command
        Then route.route.openshift.io "jenkins" created
        And  configmap "jenkins-trusted-ca-bundle" created
        And  persistentvolumeclaim "jenkins" created
        Then we check the pvc status is "Bound"
        And  deploymentconfig.apps.openshift.io "jenkins" created
        And  serviceaccount "jenkins" created
        And rolebinding.authorization.openshift.io "jenkins_edit" created
        And service "jenkins-jnlp" created
        And service "jenkins" created
        Then We check for deployment pod status to be "Completed"
        And We check for jenkins master pod status to be "Ready"