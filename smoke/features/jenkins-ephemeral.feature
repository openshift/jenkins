Feature: Deploy Jenkins on openshift using template based install

    As a user of openshift
    I want to deploy Jenkins and configure my CI/CD on openshift cluster

    Background:
    Given Project [TEST_NAMESPACE] is used
    Then we delete deploymentconfig.apps.openshift.io "jenkins"
        And  we delete route.route.openshift.io "jenkins"
        And  delete configmap "jenkins-trusted-ca-bundle"
        And  delete serviceaccount "jenkins"
        And  delete rolebinding.authorization.openshift.io "jenkins_edit"
        And  delete service "jenkins-jnlp"
        And  delete service "jenkins"
        And delete all deploymentconfig
        And delete all remaining test resources

    Scenario: Create jenkins  using ephemeral template
        Given we have a openshift cluster
        When User enters oc new-app jenkins-ephemeral command
        Then route.route.openshift.io "jenkins" created
        And  configmap "jenkins-trusted-ca-bundle" created
        And  deploymentconfig.apps.openshift.io "jenkins" created
        And  serviceaccount "jenkins" created
        And rolebinding.authorization.openshift.io "jenkins_edit" created
        And service "jenkins-jnlp" created
        And service "jenkins" created
        Then We check for deployment pod status to be "Completed"
        And We check for jenkins master pod status to be "Ready"