Feature: Upon jenkins master pod deletion/destruction the data(jobs) persist when using persistent template

    As a user of Jenkins persistent template
    I want to test upon jenkins master pod deletion/destruction the data(jobs) persist

    Background:
    Given Project [TEST_NAMESPACE] is used

    @automated @customer-scenario
    Scenario: Test persistence of jenkins while using persistent template : JKNS-09-TC01
      Given The jenkins pod is up and runnning
      Then We rsh into the master pod and check the jobs count
      When We delete the jenkins master pod
      Then We ensure that jenkins deployment config is ready
      And We rsh into the master pod & Compare if the data persist or is lost upon pod restart
