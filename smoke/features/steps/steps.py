# @mark.steps
# ----------------------------------------------------------------------------
# STEPS:
# ----------------------------------------------------------------------------
import json
import os
import time
import urllib3
from behave import given, then, when
from kubernetes import config
from pyshould import should
from smoke.features.steps.openshift import Openshift
from smoke.features.steps.project import Project
from build_steps import verify_status
# Test results file path
scripts_dir = os.getenv('OUTPUT_DIR')
java_builder = './smoke/samples/java-builder-cm.yaml'
nodejs_builder = './smoke/samples/nodejs-builder-cm.yaml'

# variables needed to get the resource status
deploy_pod = "jenkins-1-deploy"
global jenkins_master_pod
global current_project
current_project = ''
config.load_kube_config()
oc = Openshift()
podStatus = {}
podtemplate_build_ref = 'https://github.com/akram/pipes.git\#container-nodes'

# STEP
@given(u'Project "{project_name}" is used')
def given_project_is_used(context, project_name):
    project = Project(project_name)
    current_project = project_name
    context.current_project = current_project
    context.oc = oc
    if not project.is_present():
        print("Project is not present, creating project: {}...".format(project_name))
        project.create() | should.be_truthy.desc(
            "Project {} is created".format(project_name))
    print("Project {} is created!!!".format(project_name))
    context.project = project
def before_feature(context, feature):
    if scenario.name != None and "TEST_NAMESPACE" in scenario.name:
        print("Scenario using env namespace subtitution found: {0}, env: {}".format(scenario.name, os.getenv("TEST_NAMESPACE")))
        scenario.name = txt.replace("TEST_NAMESPACE", os.getenv("TEST_NAMESPACE"))

# STEP
@given(u'Project [{project_env}] is used')
def given_namespace_from_env_is_used(context, project_env):
    env = os.getenv(project_env)
    assert env is not None, f"{project_env} environment variable needs to be set"
    print(f"{project_env} = {env}")
    given_project_is_used(context, env)

@then(u'We check for jenkins master pod status to be "Ready"')
def jenkinsMasterPodStatus(context):
    current_project = context.current_project
    jenkins_master_pod = oc.getmasterpod(current_project)
    print('---------Getting default jenkins pod name---------')
    print(jenkins_master_pod)
    container_status = oc.get_resource_info_by_jsonpath('pods',jenkins_master_pod,current_project,json_path='{.status.containerStatuses[*].ready}')
    print(container_status)
    if 'false' in container_status:
        raise AssertionError

@given(u'The jenkins pod is up and runnning')
def checkJenkins(context):
    jenkinsMasterPodStatus(context)
@then(u'we configure custom agents as Kubernetes pod template by creating configmap using "smoke/samples/java-builder-cm.yaml" and "smoke/samples/nodejs-builder-cm.yaml"')
def configure_pod_templates(context):
    print("Initiazing java pod template")
    res = oc.oc_create_from_yaml(java_builder)
    if(res == None):
        print("Error while creating java-builder pod template")
        raise AssertionError
    else:
        print("Initiazing nodejs pod template")
        nodejs_res = oc.oc_create_from_yaml(nodejs_builder)
        if(nodejs_res == None):
            print("Error while creating nodejs-builder pod template")
@when(u'the user creates a new build refering to "https://github.com/akram/pipes.git\#container-nodes"')
def trigger_new_build(context):
    res = oc.new_build(podtemplate_build_ref)
    time.sleep(30)
    if res == None:
        raise AssertionError
@then(u'The build pipes-1 should be in "Complete" state')
def check_build_status(context):
    verify_status(context.current_project, 'build', 'pipes-1', 5, 600, 'Complete')