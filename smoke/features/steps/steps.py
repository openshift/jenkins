# @mark.steps
# ----------------------------------------------------------------------------
# STEPS:
# ----------------------------------------------------------------------------
import json
import os
import time

import jenkins

import urllib3
from behave import given, then, when
from kubernetes import client, config
from pyshould import should
from smoke.features.steps.openshift import Openshift
from smoke.features.steps.plugins import Plugins
from smoke.features.steps.project import Project

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

@then(u'route.route.openshift.io "jenkins" created')
def checkRoute(context):
    current_project = context.current_project
    try:
        res = oc.get_route('jenkins', current_project)
        if not 'jenkins' in res:
            raise AssertionError("Route creation failed")
        item = oc.search_resource_in_namespace('route', 'jenkins', current_project)
        print(f'route {item} created')
    except AssertionError:
        print('Problem with route')


'''
Pre 4.6 configmap not available'
'''

@then(u'configmap "jenkins-trusted-ca-bundle" created')
def checkConfigmap(context):
    current_project = context.current_project
    try:
        res = oc.get_configmap(current_project)
        if not 'jenkins' in res:
            raise AssertionError("configmap creation failed")
        item = oc.search_resource_in_namespace('cm', 'jenkins-trusted-ca-bundle', current_project)
    except AssertionError:
        print('Problem with configmap')

@then(u'serviceaccount "jenkins" created')
def checkSA(context):
    current_project = context.current_project
    try:
        res = oc.get_service_account(current_project)
        if not 'jenkins' in res:
            raise AssertionError("service acoount creation failed")
        item = oc.search_resource_in_namespace('sa', 'jenkins', current_project)
        print(f'serviceaccount {item} created')
    except AssertionError:
        print('Problem with serviceaccount')


@then(u'rolebinding.authorization.openshift.io "jenkins_edit" created')
def checkRolebinding(context):
    current_project = context.current_project
    try:
        res = oc.get_role_binding(current_project)
        if not 'jenkins' in res:
            raise AssertionError("rolebinding failed")
        item = oc.search_resource_in_namespace('rolebinding', 'jenkins_edit', current_project)
        print(f'rolebinding {item} created')
    except AssertionError:
        print('Problem with rolebinding')


@then(u'service "jenkins-jnlp" created')
def checkSVCJNLP(context):
    current_project = context.current_project
    try:
        res = oc.get_service(current_project)
        if not 'jenkins-jnlp' in res:
            raise AssertionError("service acoount creation failed")
        item = oc.search_resource_in_namespace('svc', 'jenkins-jnlp', current_project)
        print(f'service {item} created')
    except AssertionError:
        print(f'Problem with serviceJNLP')


@then(u'service "jenkins" created')
def checkSVC(context):
    current_project = context.current_project
    try:
        res = oc.get_service(current_project)
        if not 'jenkins' in res:
            raise AssertionError("service account creation failed")
        item = oc.search_resource_in_namespace('svc', 'jenkins', current_project)
        print(f'service {item} created')
    except AssertionError:
        print(f'Problem with service jenkins')

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
            raise AssertionError
@then(u'we check configmap "jenkins-agent-java-builder" and "jenkins-agent-nodejs" should be created')
def verify_configmap(context):
    current_project = context.current_project
    if not 'jenkins-agent-java-builder' in oc.search_resource_in_namespace('configmap','jenkins-agent-java-builder',current_project):
        raise AssertionError
    elif not 'jenkins-agent-nodejs' in oc.search_resource_in_namespace('configmap','jenkins-agent-nodejs',current_project):
        raise AssertionError
@when(u'the user creates a new build refering to "https://github.com/akram/pipes.git\#container-nodes"')
def trigger_new_build(context):
    res = oc.new_build(podtemplate_build_ref)
    time.sleep(30)
    if res == None:
        raise AssertionError
@then(u'buildconfig.build.openshift.io "pipes" should be created')
def search_buildconfig(context):
    current_project = context.current_project
    item = oc.search_resource_in_namespace('buildconfig', 'pipes', current_project)
    if item is None:
        raise AssertionError
    else:
        print(f'buildconfig {item} found')
    
@then(u'we check for "java-builder" agent node and "nodejs-builder" agent node are created')
def check_pods(context):
    current_project = context.current_project
    time.sleep(75)
    '''This sleep is needed as the jenkinspipeline takes time to build and it takes time for the java-builder agent node to comeup'''
    java_builder_pod = oc.search_resource_in_namespace('pods','java-builder-template',current_project)
    if java_builder_pod is None:
        raise AssertionError
    else:
        time.sleep(45)
        '''This sleep is needed as the jenkinspipeline takes time to build and it takes time for the nodejs agent node to comeup'''
        nodejs_pod = oc.search_resource_in_namespace('pods','nodejs-builder-template',current_project)
        if nodejs_pod is None:
            raise AssertionError
        else:
            print(nodejs_pod)

@then(u'The build pipes-1 should be in "Complete" state')
def check_build_status(context):
    current_project = context.current_project
    time.sleep(30)
    current_phase = oc.get_resource_info_by_jsonpath('build','pipes-1',current_project,'.status.phase',wait=True)
    if 'Failed' in current_phase:
        raise AssertionError
    else:
        print('Pipes-1 current phase is ', current_phase)