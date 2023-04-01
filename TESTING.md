Test
---------------------------------

This repository also provides a test framework which checks basic functionality
of the Jenkins image.

With v3, users can choose between testing Jenkins based on a RHEL (where you are running on a platform with subscriptions) or CentOS image.
With v4, there is not CentOS vs. RHEL distinction, but we still use TARGET to control whether subscriptions are required when building the test image,
and we reuse the v3 values (i.e. `rhel7`) for that purpose.

*  **RHEL based image**

    To test a RHEL7 based Jenkins image, you need to run the test on a properly
    subscribed RHEL machine.

    ```
    $ cd jenkins
    $ make test TARGET=rhel7 VERSION=2
    ```

*  **CentOS based image**

    ```
    $ cd jenkins
    $ make test VERSION=2
    ```

**Notice: By omitting the `VERSION` parameter, the build/test action will be performed
on all provided versions of Jenkins. Since we are currently providing only version `2`,
you can omit this parameter.**

## PR testing for this repository

As with the plugins focused on OpenShift integration, see [the contribution guide](CONTRIBUTING_TO_OPENSHIFT_JENKINS_IMAGE_AND_PLUGINS.md).
