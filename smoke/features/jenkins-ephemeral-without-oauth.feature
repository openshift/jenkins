Feature: Deploy Jenkins on openshift using template based install

    As a user of Jenkins on OpenShift
    I want ensure that I can login to Jenkins after setting a custom 
    password and disabling oauth

    Background:
    Given Project [TEST_NAMESPACE] is used

    Scenario Outline: Create jenkins using ephemeral template by passing 
              the environement variables JENKINS_PASSWORD=password2 and 
              OPENSHIFT_OAUTH_ENABLE=false and check that we can login to jenkins with the provided password.
        Given we have a openshift cluster
        And cleared from all test resources
        And environment variables <env_vars> are set
        Examples: Input Variables
            |env_vars                |
            |{ "JENKINS_PASSWORD": "password2", "OPENSHIFT_ENABLE_OAUTH": "false" }|
        When User enters oc new-app jenkins-ephemeral command using env vars
        Then route.route.openshift.io "jenkins" created
        And  deploymentconfig.apps.openshift.io "jenkins" created
        Then We ensure that jenkins deployment config status mets criteria "condition=Available"
        Then We check that JENKINS_PASSWORD environement variable is set to password2
        Then We ensure that we can login to jenkins using admin and password2
        Then We set env var JENKINS_PASSWORD to value password3 in deploymentconfig jenkins
        Then We ensure that jenkins deployment config is ready
        Then We check that JENKINS_PASSWORD environement variable is set to password3
        Then We ensure that we can login to jenkins using admin and password3
