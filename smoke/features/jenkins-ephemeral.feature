Feature: Deploy Jenkins on openshift using template based install

    As a user of openshift
    I want to deploy Jenkins and configure my CI/CD on openshift cluster

    Background:
    Given Project [TEST_NAMESPACE] is used
    # Then we delete deploymentconfig.apps.openshift.io "jenkins"
    #     And  we delete route.route.openshift.io "jenkins"
    #     And  delete configmap "jenkins-trusted-ca-bundle"
    #     And  delete serviceaccount "jenkins"
    #     And  delete rolebinding.authorization.openshift.io "jenkins_edit"
    #     And  delete service "jenkins-jnlp"
    #     And  delete service "jenkins"
    #     And delete all deploymentconfig
    #     And delete all remaining test resources

    @automated @customer-scenario
    Scenario: Create jenkins  using ephemeral template : JKNS-02-TC01
        Given we have a openshift cluster
        When User enters oc new-app jenkins-ephemeral command
        Then we check that the resources are created
        | resource         | resource_name             |
        | route            | jenkins                   |
        | configmap        | jenkins-trusted-ca-bundle |
        | deploymentconfig | jenkins                   |
        | serviceaccount   | jenkins                   |
        | rolebinding      | jenkins_edit              |
        | service          | jenkins-jnlp              |
        | service          | jenkins                   |
        Then We check for deployment pod status to be "Completed"
        And We check for jenkins master pod status to be "Ready"