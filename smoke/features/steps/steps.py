# @mark.steps
# ----------------------------------------------------------------------------
# STEPS:
# ----------------------------------------------------------------------------
import time
import logging
from datetime import date
from behave import given, then, when
from pyshould import should
from kubernetes import config, client
from smoke.features.steps.openshift import Openshift
from smoke.features.steps.project import Project

config.load_kube_config()
v1 = client.CoreV1Api()


def get_filename_datetime():
    # Use current date to get a text file name.
    return "" + str(date.today()) + ".txt"


# Get full path for writing.
file_name = get_filename_datetime()
path = "./smoke/logs" + file_name

logging.basicConfig(filename=path, format='%(asctime)s: %(levelname)s: %(message)s', datefmt='%m/%d/%Y %I:%M:%S %p')
logger = logging.getLogger()
logger.setLevel(logging.INFO)

project_name = 'jenkins-test'
oc = Openshift()
deploy_pod = "jenkins-1-deploy"
pod_name_pattern = "{name}.*$(?<!-build)"


@given(u'Project jenkins-test is used')
def setProject(context):
    project = Project(project_name)
    context.oc = oc
    if not project.is_present():
        logger.info("Project is not present, creating project: {}...".format(project_name))
        project.create() | should.be_truthy.desc("Project {} is created".format(project_name))

    logger.info(f'Project {project_name} is created!!!')
    context.project = project


@when(u'User enters oc new-app jenkins-ephemeral command')
def createOperator(context):
    res = oc.new_app('jenkins-ephemeral', project_name)
    logger.info(f'{res}')
    logger.info("Checking resources")


@then(u'route.route.openshift.io "jenkins" created')
def checkRoute(context):
    try:
        res = oc.get_route('jenkins', project_name)
        if not 'jenkins' in res:
            raise AssertionError("Route creation failed")
        item = oc.search_resource_in_namespace('route', 'jenkins', project_name)
        logger.info(f'route {item} created')
    except AssertionError:
        logger.debug(f'Problem with route')


'''
Pre 4.6 configmap not available'
'''


@then(u'configmap "jenkins-trusted-ca-bundle" created')
def checkConfigmap(context):
    try:
        res = oc.get_configmap(project_name)
        logger.info('Pre 4.6 not available')
        if not 'jenkins' in res:
            raise AssertionError("configmap creation failed")
        item = oc.search_resource_in_namespace('cm', 'jenkins-trusted-ca-bundle', project_name)
        logger.info(f'configmap {item} created')
    except AssertionError:
        logger.debug(f'Problem with configmap')


@then(u'deploymentconfig.apps.openshift.io "jenkins" created')
def checkDC(context):
    try:
        res = oc.get_deploymentconfig(project_name)
        if not 'jenkins' in res:
            raise AssertionError("deploymentconfig creation failed")
        item = oc.search_resource_in_namespace('dc', 'jenkins', project_name)
        logger.info(f'deploymentconfig {item} created')
    except AssertionError:
        logger.debug(f'Problem with deploymentconfig')


@then(u'serviceaccount "jenkins" created')
def checkSA(context):
    try:
        res = oc.get_service_account(project_name)
        if not 'jenkins' in res:
            raise AssertionError("service acoount creation failed")
        item = oc.search_resource_in_namespace('sa', 'jenkins', project_name)
        logger.info(f'serviceaccount {item} created')
    except AssertionError:
        logger.debug(f'Problem with serviceaccount')


@then(u'rolebinding.authorization.openshift.io "jenkins_edit" created')
def checkrolebinding(context):
    try:
        res = oc.get_role_binding(project_name)
        if not 'jenkins' in res:
            raise AssertionError("rolebinding failed")
        item = oc.search_resource_in_namespace('rolebinding', 'jenkins_edit', project_name)
        logger.info(f'rolebinding {item} created')
    except AssertionError:
        logger.debug(f'Problem with rolebinding')


@then(u'service "jenkins-jnlp" created')
def checkSVCJNLP(context):
    try:
        res = oc.get_service(project_name)
        if not 'jenkins-jnlp' in res:
            raise AssertionError("service acoount creation failed")
        item = oc.search_resource_in_namespace('svc', 'jenkins-jnlp', project_name)
        logger.info(f'service {item} created')
    except AssertionError:
        logger.debug(f'Problem with serviceJNLP')


@then(u'service "jenkins" created')
def checkSVC(context):
    try:
        res = oc.get_service(project_name)
        if not 'jenkins' in res:
            raise AssertionError("service acoount creation failed")
        item = oc.search_resource_in_namespace('svc', 'jenkins', project_name)
        logger.info(f'service {item} created')
    except AssertionError:
        logger.debug(f'Problem with service jenkins')


@then(u'The operator pod and deployment pod must be runnning')
def verifyPodStatus(context):
    context.v1 = v1
    podStatus = {}
    time.sleep(300)
    pods = v1.list_namespaced_pod(project_name)
    for i in pods.items:
        logger.info("Getting pod list")
        logger.info(i.status.pod_ip)
        logger.info(i.metadata.name)
        logger.info(i.status.phase)
        podStatus[i.metadata.name] = i.status.phase
    for pod in podStatus.keys():
        status = podStatus[pod]
        if 'Running' in status:
            logger.info("still checking pod status")
            logger.info(pod)
            logger.info(podStatus[pod])
        elif 'Succeeded' in status:
            logger.info("checking pod status")
            logger.info(pod)
            logger.info(podStatus[pod])
        else:
            logger.critical("Pod is not ready->")
            raise AssertionError
    # logger.info("Getting Pod List and status")
    # logger.info(podStatus)
