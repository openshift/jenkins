Feature: Delete all resources created using jenkins ephemeral template

    We want to delete the resources created using jenkins ephemeral template
    We want this to continue in the same namespace & start testing the jenkins persistent template based install

    Background:
    Given Project [TEST_NAMESPACE] is used

    @automated @customer-scenario
    Scenario: Delete all resources : JKNS-05-TC01
        Given we have a openshift cluster
        Then we delete deploymentconfig.apps.openshift.io "jenkins"
        And  we delete route.route.openshift.io "jenkins"
        And  delete configmap "jenkins-trusted-ca-bundle"
        And  delete serviceaccount "jenkins"
        And  delete rolebinding.authorization.openshift.io "jenkins_edit"
        And  delete service "jenkins-jnlp"
        And  delete service "jenkins"
        And  delete all buildconfigs
        And  delete all builds
        And delete all deploymentconfig
        And delete all services
        And delete all imagestream
        And delete all remaining test resources