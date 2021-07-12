Feature: Upon jenkins master pod deletion/destruction the data(jobs) persist when using persistent template

    As a user of Jenkins persistent template
    I want to test upon jenkins master pod deletion/destruction the data(jobs) persist

    Background:
    Given Project [TEST_NAMESPACE] is used

    Scenario: Test persistence of jenkins while using persistent template
      Given The jenkins pod is up and runnning
      Then We rsh into the master pod and check the jobs count
      When We delete the jenkins master pod
      Then We check for deployment pod status to be "Completed"
      Then We check for jenkins master pod status to be "Ready"
      And We rsh into the master pod & Compare if the data persist or is lost upon pod restart
