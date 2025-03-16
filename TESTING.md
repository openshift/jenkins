Test
---------------------------------

This repository also provides a test framework which checks basic functionality
of the Jenkins image.

The TARGET variable is used to determine whether ubi, rhel8, or rhel9 is selected
when building the test image.

*  **RHEL based image**

    To test a RHEL based Jenkins image, you need to run the test on a properly
    subscribed RHEL machine.

    ```
    $ cd jenkins
    $ make test TARGET=rhel9 VERSION=2
     ```

*  **UBI image**

    ```
    $ cd jenkins
    $ make test TARGET=ubi VERSION=2
    ```

**Notice: By omitting the `VERSION` parameter, the build/test action will be performed
on all provided versions of Jenkins. Since we are currently providing only version `2`,
you can omit this parameter.**

## PR testing for this repository

As with the plugins focused on OpenShift integration, see [the contribution guide](CONTRIBUTING_TO_OPENSHIFT_JENKINS_IMAGE_AND_PLUGINS.md).
