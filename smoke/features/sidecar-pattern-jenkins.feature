Feature: Use sidecar pattern for Jenkins pod templates

    As a developer using Jenkins on openshift to build my application
    I want to use the base Jenkins agent image as a sidecar in my PodTemplate and
    Provide new Kubernetes Plugin Pod Templates which uses the sidecar pattern for NodeJS and Maven.
    So that I can use any s2i builder image in my Jenkins pipelines
    
    

    Background:
    Given Project [TEST_NAMESPACE] is used

    Scenario: Trigger a build that verifies the new pod templates can successfully execute a JenkinsPipeline build.
    Given The jenkins pod is up and runnning
    Then we configure custom agents as Kubernetes pod template by creating configmap using "smoke/samples/java-builder-cm.yaml" and "smoke/samples/nodejs-builder-cm.yaml"
    And we check configmap "jenkins-agent-java-builder" and "jenkins-agent-nodejs" should be created
    When the user creates a new build refering to "https://github.com/akram/pipes.git#pod-templates"
    Then buildconfig.build.openshift.io "pipes" should be created
    And build pipes-1 should be in "Running" state
    Then we wait for "java-builder-template" pod to have state as Ready[2/2]
    And wait for "nodejs-builder-template" pod to have state as Ready[2/2]
    And The build pipes-1 should be in "Complete" state

    