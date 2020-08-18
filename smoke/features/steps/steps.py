# @mark.steps
# ----------------------------------------------------------------------------
# STEPS:
# ----------------------------------------------------------------------------
import time
import logging
import urllib3
from datetime import date
from behave import given, then, when
from pyshould import should
from kubernetes import config, client
# import jenkins
# from jenkinsapi.jenkins import Jenkins
from smoke.features.steps.openshift import Openshift
from smoke.features.steps.project import Project


config.load_kube_config()
v1 = client.CoreV1Api()


def get_filename_datetime():
    # Use current date to get a text file name.
    return "" + str(date.today()) + ".txt"


# Get full path for writing.
file_name = get_filename_datetime()
path = "./smoke/logs-" + file_name

logging.basicConfig(filename=path, format='%(asctime)s: %(levelname)s: %(message)s', datefmt='%m/%d/%Y %I:%M:%S %p')
logger = logging.getLogger()
logger.setLevel(logging.INFO)

project_name = 'jenkins-test'
oc = Openshift()
deploy_pod = "jenkins-1-deploy"
samplebclst = ['sample-pipeline','nodejs-mongodb-example']
samplepipeline = "https://raw.githubusercontent.com/openshift/origin/master/examples/jenkins/pipeline/samplepipeline.yaml"
mavenpipeline = "https://raw.githubusercontent.com/openshift/origin/master/examples/jenkins/pipeline/maven-pipeline.yaml"
nodejspipeline = "https://github.com/akram/scrum-planner.git"


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
    podStatus = {}
    context.v1 = v1
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


@given(u'The jenkins pod is up and runnning')
def checkJenkins(context):
    time.sleep(30)
    podStatus = {}
    status = ""
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


@when(u'The user enters new-app command with sample-pipeline')
def createPipeline(context):
    # bclst = ['sample-pipeline','nodejs-mongodb-example']
    res = oc.new_app_from_file(samplepipeline,project_name)
    for item, value in enumerate(samplebclst):
        if 'sample-pipeline' in oc.search_resource_in_namespace('bc',value, project_name):
            logger.info('Buildconfig sample-pipeline created')
        elif 'nodejs-mongodb-example' in oc.search_resource_in_namespace('bc',value,project_name):
            logger.info('Buildconfig nodejs-mongodb-example created')
        else:
            logger.error("----> Something went wrong with createPipeline")
            raise AssertionError
    logger.info(res)


@then(u'Trigger the build using oc start-build')
def startbuild(context):
    for item,value in enumerate(samplebclst):
        res = oc.start_build(value,project_name)
        if not value in res:
            raise AssertionError
        else:
            logger.info(res)


@then(u'nodejs-mongodb-example pod must come up')
def check_app_pod(context):
    time.sleep(120)
    podStatus = {}
    podSet = set()
    bcdcSet = set()
    pods = v1.list_namespaced_pod(project_name)
    for i in pods.items:
        podStatus[i.metadata.name] = i.status.phase
        podSet.add(i.metadata.name)
    
    for items in podSet:
        if 'build' in items:
           bcdcSet.add(items)
        elif 'deploy' in items:
            bcdcSet.add(items)

    app_pods = podSet.difference(bcdcSet)
    for items in app_pods:
        logger.info('Getting pods')
        logger.info(items)
    
    for items in app_pods:
        for pod in podStatus.keys():
            status = podStatus[items]
            if not 'Running' in status:
                raise AssertionError
    logger.info('---> App pods are ready')

@then(u'route nodejs-mongodb-example must be created and be accessible')
def connectApp(context):
    logger.info('Getting application route/url')
    app_name = 'nodejs-mongodb-example'
    time.sleep(30)
    route = oc.get_route_host(app_name,project_name)
    url = 'http://'+str(route)
    logger.info('--->App url:')
    logger.info(url)
    http = urllib3.PoolManager()
    res = http.request('GET', url)
    connection_status = res.status
    if connection_status == 200:
        logger.info('---> Application is accessible via the route')
        logger.info(url)
    else:
        raise Exception
    
