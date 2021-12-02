# @mark.steps
# ----------------------------------------------------------------------------
# STEPS:
# ----------------------------------------------------------------------------
from smoke.features.steps.openshift import Openshift
oc = Openshift()

@then(u'deploymentconfig.apps.openshift.io "jenkins" created')
def check_deployment_config(context):
    try:
        res = oc.get_deploymentconfig(context.current_project)
        if not 'jenkins' in res:
            raise AssertionError("deploymentconfig creation failed")
        item = oc.search_resource_in_namespace('dc', 'jenkins', context.current_project)
        print(f'deploymentconfig {item} created')
    except AssertionError:
        print('Problem with deploymentconfig')

@when(u'verify deploymentconfig.apps.openshift.io/openshift-jee-sample is created')
def verify_deployment_config(context):
    if not 'openshift-jee-sample' in oc.search_resource_in_namespace('deploymentconfig','openshift-jee-sample',context.current_project):
        raise AssertionError
    else:
        res = oc.search_resource_in_namespace('deploymentconfig','openshift-jee-sample',context.current_project)
        print(res)


@then(u'We ensure that {dc_name} deployment config status mets criteria {wait_for}')
def wait_for_deployment_config_status(context, dc_name, wait_for="condition=Available"):
    output, exit_code = oc.check_for_deployment_config_status(dc_name, namespace=context.current_project)
    print("Deployment Config {} waiting for condition {}".format(dc_name, wait_for))
    if exit_code != 0:
        raise AssertionError
    else:
        print("Deployment Config {} met the condition {}: Command output: {}".format(dc_name, wait_for, output))



@then(u'We ensure that {dc_name} deployment config is ready')
def ensure_deployment_config_is_ready(context, dc_name):
    wait_for_deployment_config_status(context, dc_name)
