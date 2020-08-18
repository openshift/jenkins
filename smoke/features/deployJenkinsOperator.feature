Feature: Deploy Jenkins Operator

    As a user of Jenkins Operator
    I deploy Jenkins and configure my CI/CD on openshift cluster

    Scenario: Create jenkins operator using ephemeral template
        Given Project jenkins-test is used
        When User enters oc new-app jenkins-ephemeral command
        Then route.route.openshift.io "jenkins" created
        And  configmap "jenkins-trusted-ca-bundle" created
        And  deploymentconfig.apps.openshift.io "jenkins" created
        And  serviceaccount "jenkins" created
        And rolebinding.authorization.openshift.io "jenkins_edit" created
        And service "jenkins-jnlp" created
        And service "jenkins" created
        Then The operator pod and deployment pod must be runnning

    Scenario: Deploy sample application on openshift
        Given The jenkins pod is up and runnning
        When The user enters new-app command with sample-pipeline
        Then Trigger the build using oc start-build
        Then nodejs-mongodb-example pod must come up
        And route nodejs-mongodb-example must be created and be accessible