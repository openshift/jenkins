Feature: Use sidecar pattern for Jenkins pod templates

    As a developer using Jenkins on openshift to build my application
    I want to use the base Jenkins agent image as a sidecar in my PodTemplate and
    Provide new Kubernetes Plugin Pod Templates which uses the sidecar pattern for NodeJS and Maven.
    So that I can use any s2i builder image in my Jenkins pipelines
    
    

    Background:
    Given Project [TEST_NAMESPACE] is used
    And delete all buildconfigs
    And delete all builds
    And delete all deploymentconfig
    And delete all remaining test resources

    Scenario: Trigger a build that verifies the new pod templates can successfully execute a JenkinsPipeline build.
    Given The jenkins pod is up and runnning
    When The user creates a new build using oc new-build command
    Then buildconfig.build.openshift.io "pipes" should be created
    And build pipes-1 should be in "Running" state
    Then we wait for "java-builder" pod to have state as Ready[2/2]
    And wait for "nodejs-builder" pod to have state as Ready[2/2]
    And The build pipes-1 should be in "Complete" state

    