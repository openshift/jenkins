Feature: Use sidecar pattern for Jenkins pod templates

    As a developer using Jenkins on openshift to build my application
    I want to use the base Jenkins agent image as a sidecar in my PodTemplate and
    Provide new Kubernetes Plugin Pod Templates which uses the sidecar pattern for NodeJS and Maven.
    So that I can use any s2i builder image in my Jenkins pipelines
    
    

    Background:
    Given Project [TEST_NAMESPACE] is used

    @automated @customer-scenario
    Scenario: Trigger a build that verifies the new pod templates can successfully execute a JenkinsPipeline build. : JKNS-10-TC01
    Given The jenkins pod is up and runnning
    Then we configure custom agents as Kubernetes pod template by creating configmap using "smoke/samples/java-builder-cm.yaml" and "smoke/samples/nodejs-builder-cm.yaml"
    When the user creates a new build refering to "https://github.com/akram/pipes.git\#container-nodes"
    Then we check that the resources are created
    | resource         | resource_name              |
    | configmap        | jenkins-agent-java-builder |
    | configmap        | jenkins-agent-nodejs       |
    | buildconfig      | pipes                      |
    And The build pipes-1 should be in "Complete" state

    