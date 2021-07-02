Feature: Verify plugins with the correct version are installed inside jenkins master pod

    As security check
    we want to make sure we have the latest plugins installed inside the master pod

    Background:
    Given Project [TEST_NAMESPACE] is used

    Scenario: Verify plugins are installed with desired version
        Given The jenkins pod is up and runnning
        When We rsh into the master pod
        Then We compare the plugins version inside the master pod with the plugins listed in plugins.txt