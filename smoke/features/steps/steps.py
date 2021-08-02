# @mark.steps
# ----------------------------------------------------------------------------
# STEPS:
# ----------------------------------------------------------------------------
import os
import time
import urllib3
from behave import given, when, then
from pyshould import should
from kubernetes import config, client
from smoke.features.steps.openshift import Openshift
from smoke.features.steps.project import Project
from smoke.features.steps.plugins import Plugins



# Test results file path
scripts_dir = os.getenv('OUTPUT_DIR')

# Path to pipeline job to test agent images
maven_template ='./smoke/samples/maven_pipeline.yaml'
nodejs_template = './smoke/samples/nodejs_pipeline.yaml'



# variables needed to get the resource status
deploy_pod = "jenkins-1-deploy"
jenkins_master_pod = ''
current_project = ''
config.load_kube_config()
v1 = client.CoreV1Api()
oc = Openshift()
podStatus = {}
buildconfigs = {'sample-pipeline':'1','openshift-jee-sample':'1'}
builds = {}
# Parse the base plugins from the file and store them in a dictonary with key=plugin-name & value=plugin-version

baseplugins = './2/contrib/openshift/base-plugins.txt'
p = Plugins()
plugins = p.getPlugins(baseplugins)


def triggerbuild(buildconfig,namespace):
    print('Triggering build:',buildconfig)
    res = oc.start_build(buildconfig,namespace)
    print(res)

# STEP
@given(u'Project "{project_name}" is used')
def given_project_is_used(context, project_name):
    project = Project(project_name)
    global current_project
    current_project = project_name
    context.current_project = current_project
    context.oc = oc
    if not project.is_present():
        print("Project is not present, creating project: {}...".format(project_name))
        project.create() | should.be_truthy.desc(
            "Project {} is created".format(project_name))
    print("Project {} is created!!!".format(project_name))
    context.project = project


# STEP
@given(u'Project [{project_env}] is used')
def given_namespace_from_env_is_used(context, project_env):
    env = os.getenv(project_env)
    assert env is not None, f"{project_env} environment variable needs to be set"
    print(f"{project_env} = {env}")
    given_project_is_used(context, env)


@given(u'we have a openshift cluster')
def loginCluster(context):
    print("Using [{}]".format(current_project))

@when(u'User enters oc new-app jenkins-ephemeral command')
def ephemeralTemplate(context):
    res = oc.new_app('jenkins-ephemeral', current_project)
    if(res == None):
        print("Error while installing jenkins using ephemeral template")
        raise AssertionError

@then(u'route.route.openshift.io "jenkins" created')
def checkRoute(context):
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
    try:
        res = oc.get_configmap(current_project)
        if not 'jenkins' in res:
            raise AssertionError("configmap creation failed")
        item = oc.search_resource_in_namespace('cm', 'jenkins-trusted-ca-bundle', current_project)
    except AssertionError:
        print('Problem with configmap')


@then(u'deploymentconfig.apps.openshift.io "jenkins" created')
def checkDC(context):
    try:
        res = oc.get_deploymentconfig(current_project)
        if not 'jenkins' in res:
            raise AssertionError("deploymentconfig creation failed")
        item = oc.search_resource_in_namespace('dc', 'jenkins', current_project)
        print(f'deploymentconfig {item} created')
    except AssertionError:
        print('Problem with deploymentconfig')


@then(u'serviceaccount "jenkins" created')
def checkSA(context):
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
    try:
        res = oc.get_service(current_project)
        if not 'jenkins' in res:
            raise AssertionError("service acoount creation failed")
        item = oc.search_resource_in_namespace('svc', 'jenkins', current_project)
        print(f'service {item} created')
    except AssertionError:
        print(f'Problem with service jenkins')

@then(u'We check for deployment pod status to be "Completed"')
def deploymentPodStatus(context):
    time.sleep(120)
    print("Getting deployment pod status")
    deploy_pod_status = oc.get_resource_info_by_jsonpath('pods',deploy_pod,current_project,json_path='{.status.phase}')
    if not 'Succeeded' in deploy_pod_status:
        raise AssertionError


@then(u'We check for jenkins master pod status to be "Ready"')
def jenkinsMasterPodStatus(context):
    global jenkins_master_pod
    jenkins_master_pod = ""
    jenkins_master_pod = getmasterpod(current_project)
    print('---------Getting default jenkins pod name---------')
    print(jenkins_master_pod)
    containerState = oc.get_resource_info_by_jsonpath('pods',jenkins_master_pod,current_project,json_path='{.status.containerStatuses[*].ready}')
    if 'false' in containerState:
        raise AssertionError
    else:
         print(containerState)

@then(u'persistentvolumeclaim "jenkins" created')
def verify_pvc(context):
    if not 'jenkins' in oc.search_resource_in_namespace('pvc','jenkins',current_project):
        raise AssertionError
    else:
        res = oc.search_resource_in_namespace('pvc','jenkins',current_project)
        print(res)


@then(u'we check the pvc status is "Bound"')
def pvc_status(context):
    print('---------Getting pvc status---------')
    pvcState = oc.get_resource_info_by_jsonpath('pvc','jenkins',current_project,json_path='{.status.phase}')
    if 'Bound' in pvcState:
        print(pvcState)
    else:
        raise AssertionError

@given(u'The jenkins pod is up and runnning')
def checkJenkins(context):
    jenkinsMasterPodStatus(context)

@when(u'User enters oc new-app jenkins-persistent command')
def persistentTemplate(context):
    res = oc.new_app('jenkins-persistent', current_project)
    if(res == None):
        print("Error while installing jenkins using persistent template")
        raise AssertionError

@when(u'The user enters new-app command with nodejs_template')
def createPipeline(context):
    res = oc.new_app_from_file(nodejs_template,current_project)
    time.sleep(30)
    if 'sample-pipeline' in oc.search_resource_in_namespace('bc','sample', current_project):
        print('Buildconfig sample-pipeline created')
    elif 'nodejs-postgresql-example' in oc.search_resource_in_namespace('bc','postgersql',current_project):
        print('Buildconfig nodejs-postgersql-example created')
    else:
        raise AssertionError
    print(res)

@then(u'Trigger the build using oc start-build')
def startbuild(context):
    triggerbuild('sample-pipeline',current_project)


@then(u'verify the build status of "nodejs-postgresql-example-1" build is Complete')
def verifynodejsBuildStatus(context):
    time.sleep(300)
    buildState = oc.get_resource_info_by_jsonpath('build','nodejs-postgresql-example-1',current_project,json_path='{.status.phase}')
    if not 'Complete' in buildState:
        raise AssertionError
    else:
        print("Build nodejs-postgresql-example-1 status:{buildState}")

@then(u'verify the build status of "nodejs-postgresql-example-2" build is Complete')
def verifynodejsBuildBStatus(context):
    buildState = oc.get_resource_info_by_jsonpath('build','nodejs-postgresql-example-2',current_project,json_path='{.status.phase}')
    if not 'Complete' in buildState:
        raise AssertionError
    else:
        print("Build nodejs-postgresql-example-2 status:{buildState}")


@then(u'route nodejs-postgresql-example must be created and be accessible')
def connectApp(context):
    print('Getting application route/url')
    app_name = 'nodejs-postgresql-example'
    route = oc.get_route_host(app_name,current_project)
    url = 'http://'+str(route)
    print('--->App url:')
    print(url)
    http = urllib3.PoolManager()
    res = http.request_encode_url('GET',url)
    connection_status = res.status
    if connection_status == 200:
        print('---> Application is accessible via the route')
        print(url)
        http.clear()
    else:
        raise Exception

@when(u'The user create objects from the sample maven template by processing the template and piping the output to oc create')
def createMavenTemplate(context):
    res = oc.oc_process_template(maven_template)
    print(res)

@when(u'verify imagestream.image.openshift.io/openshift-jee-sample & imagestream.image.openshift.io/wildfly exist')
def verifyImageStream(context):
    if not 'openshift-jee-sample' in oc.search_resource_in_namespace('imagestream','openshift-jee-sample', current_project):
        raise AssertionError
    elif not 'wildfly' in oc.search_resource_in_namespace('imagestream','wildfly', current_project):
        raise AssertionError
    else:
        res = oc.get_resource_lst('imagestream',current_project)
        print(res)

@when(u'verify buildconfig.build.openshift.io/openshift-jee-sample & buildconfig.build.openshift.io/openshift-jee-sample-docker exist')
def verifyBuildConfig(context):
    if not 'openshift-jee-sample' in oc.search_resource_in_namespace('buildconfig','openshift-jee-sample', current_project):
        raise AssertionError
    elif not 'openshift-jee-sample-docker' in oc.search_resource_in_namespace('buildconfig','openshift-jee-sample-docker', current_project):
        raise AssertionError
    else:
        res = oc.get_resource_lst('buildconfig',current_project)
        print(res)

@when(u'verify deploymentconfig.apps.openshift.io/openshift-jee-sample is created')
def verifyDeploymentConfig(context):
    if not 'openshift-jee-sample' in oc.search_resource_in_namespace('deploymentconfig','openshift-jee-sample',current_project):
        raise AssertionError
    else:
        res = oc.search_resource_in_namespace('deploymentconfig','openshift-jee-sample',current_project)
        print(res)

@when(u'verify service/openshift-jee-sample is created')
def verifySvc(context):
    if not 'openshift-jee-sample' in oc.search_resource_in_namespace('service','openshift-jee-sample',current_project):
        raise AssertionError
    else:
        res = oc.search_resource_in_namespace('service','openshift-jee-sample',current_project)
        print(res)

@when(u'verify route.route.openshift.io/openshift-jee-sample is created')
def verifyRoute(context):
    if not 'openshift-jee-sample' in oc.search_resource_in_namespace('route','openshift-jee-sample',current_project):
        raise AssertionError
    else:
        res = oc.search_resource_in_namespace('route','openshift-jee-sample',current_project)
        print(res)

@then(u'Trigger the build using oc start-build openshift-jee-sample')
def startBuild(context):
    triggerbuild('openshift-jee-sample',current_project)
    time.sleep(300)


@then(u'verify the build status of openshift-jee-sample-docker build is Complete')
def verifyDockerBuildStatus(context):
    buildState = oc.get_resource_info_by_jsonpath('build','openshift-jee-sample-docker-1',current_project,json_path='{.status.phase}')
    if not 'Complete' in buildState:
        raise AssertionError
    else:
        print("Build openshift-jee-sample-docker-1 status:{buildState}")
    

@then(u'verify the build status of openshift-jee-sample-1 is Complete')
def verifyJenkinsBuildStatus(context):
    buildState = oc.get_resource_info_by_jsonpath('build','openshift-jee-sample-1',current_project,json_path='{.status.phase}')
    if not 'Complete' in buildState:
        raise AssertionError
    else:
        print("Build openshift-jee-sample-1-deploy status:{buildState}")


@then(u'verify the JaveEE application is accessible via route openshift-jee-sample')
def pingApp(context):
    print('Getting application route/url')
    app_name = 'openshift-jee-sample'
    route = oc.get_route_host(app_name,current_project)
    url = 'http://'+str(route)
    print('--->App url:')
    print(url)
    http = urllib3.PoolManager()
    res = http.request_encode_url('GET',url)
    connection_status = res.status
    if connection_status == 200:
        print('---> Application is accessible via the route')
        print(url)
        http.clear()
    else:
        raise Exception

@then(u'we delete deploymentconfig.apps.openshift.io "jenkins"')
def del_dc(context):
    global jenkins_master_pod
    jenkins_master_pod = ''
    res = oc.delete("deploymentconfig","jenkins",current_project)
    if res == None:
        raise AssertionError

@then(u'we delete route.route.openshift.io "jenkins"')
def del_route(context):
    res = oc.delete("route","jenkins",current_project)
    if res == None:
        raise AssertionError


@then(u'delete configmap "jenkins-trusted-ca-bundle"')
def del_cm(context):
    res = oc.delete("configmap","jenkins-trusted-ca-bundle",current_project)
    if res == None:
        raise AssertionError


@then(u'delete serviceaccount "jenkins"')
def del_sa(context):
    res = oc.delete("serviceaccount","jenkins",current_project)
    if res == None:
        raise AssertionError


@then(u'delete rolebinding.authorization.openshift.io "jenkins_edit"')
def del_rb(context):
    res = oc.delete("rolebinding","jenkins_edit",current_project)
    if res == None:
        raise AssertionError

@then(u'delete service "jenkins"')
def del_svc(context):
    res = oc.delete("service","jenkins",current_project)
    if res == None:
        raise AssertionError


@then(u'delete service "jenkins-jnlp"')
def del_svc_jnlp(context):
    res = oc.delete("service","jenkins-jnlp",current_project)
    if res == None:
        raise AssertionError

@then(u'delete all buildconfigs')
def del_bc(context):
    res = oc.delete("bc","--all",current_project)
    if res == None:
        raise AssertionError

@then(u'delete all builds')
def del_builds(context):
    res = oc.delete("builds","--all",current_project)
    if res == None:
        raise AssertionError

@then(u'delete all deploymentconfig')
def del_alldc(context):
    res = oc.delete("deploymentconfig","--all",current_project)
    if res == None:
        raise AssertionError

@then(u'delete all build pods')
def del_pods(context):
    pods = v1.list_namespaced_pod(current_project)
    buildpods = []
    for i in pods.items:
        if 'jenkins-1-deploy' not in i.metadata.name and jenkins_master_pod not in i.metadata.name:
            buildpods.append(i.metadata.name)
    for pod in buildpods:
        res = oc.delete('pod',pod,current_project)
        print("Deleting: ",res)

@then(u'We rsh into the master pod and check the jobs count')
def getjobcount(context):
    for jobnames,_ in buildconfigs.items():
        exec_command = 'cat /var/lib/jenkins/jobs/'+current_project+'/jobs/'+current_project+'-'+jobnames+'/nextBuildNumber'
        count = oc.exec_in_pod(jenkins_master_pod,exec_command)
        buildconfigs[jobnames] = str(count)
    print(buildconfigs)

@when(u'We delete the jenkins master pod')
def deletemaster(context):
    master_pod = getmasterpod(current_project)
    res = oc.delete("pods",master_pod,current_project)
    time.sleep(60)
    if res == None:
        raise AssertionError

@then(u'We rsh into the master pod & Compare if the data persist or is lost upon pod restart')
def comparejobs(context):
    for jobnames,_ in buildconfigs.items():
        master_pod = getmasterpod(current_project)
        exec_command = 'cat /var/lib/jenkins/jobs/'+current_project+'/jobs/'+current_project+'-'+jobnames+'/nextBuildNumber'
        count = oc.exec_in_pod(master_pod,exec_command)
        buildconfigs[jobnames] = str(count)
    
    for jobnames, _ in buildconfigs.items():
        if(buildconfigs[jobnames] == '1'):
            print("Data doesnt persist")
            raise AssertionError
    print(buildconfigs)


def getmasterpod(namespace: str)-> str:
    '''
    returns the jenkins master pod name
    '''
    pods = v1.list_namespaced_pod(current_project)
    for i in pods.items:
        if 'jenkins-1-deploy' not in i.metadata.name and 'jenkins-1' in i.metadata.name:
            master_pod = i.metadata.name
    return str(master_pod)

@when(u'We rsh into the master pod')
def step_impl(context):
    pass

@then(u'We compare the plugins version inside the master pod with the plugins listed in plugins.txt')
def step_impl(context):
    pass

@when(u'We Trigger multiple builds using oc start-build openshift-jee-sample')
def multiplebuilds(context):
    global builds
    count = 1
    ## creating a dictionary of builds that keeps a track of {buildname: build_status}
         # This will be used to check the build reconcilation
    while(count <= 5):
        triggerbuild('openshift-jee-sample',current_project)
        build_name ='openshift-jee-sample-' + str(count)
        builds[build_name] = None
        count+=1
    
@when(u'We slace down the pod count in the replication controller to "0" from "1"')
def podscaling(context):
    rc_name = 'jenkins-1'
    oc.scaleReplicas(current_project,0,rc_name)
    replicas = oc.get_resource_info_by_jsonpath("dc","jenkins",current_project,json_path='{.status.availableReplicas}')
    if not '0' in replicas:
        raise AssertionError
    else:
        print('There are ',replicas,' running pods of jenkins')
    
@then(u'We delete some builds')
def deletebuilds(context):
    global builds
    rm_build = ['openshift-jee-sample-2','openshift-jee-sample-4']
    for build_name in builds.keys():
        builds[build_name] = oc.get_resource_info_by_jsonpath("build",build_name,current_project,json_path='{.status.phase}')
    print("------------Fetching all builds and build status------------")
    print(builds)
    print("------------Deleting a few  builds------------")
    for items in rm_build:
        res = oc.delete("build",items,current_project)
        print(res)
        builds.pop(items)
    print("------------Fetching all builds and build status------------")
    print(builds)
    # This sleep is used to give time for the jenkins master pod to be back
    time.sleep(60)


@then(u'verify sync plugin is able to reconcile the build state and delete the job runs associated with the builds we deleted')
def verifybuildSync(context):
    time.sleep(360)
    for build_name in builds.keys():
        builds[build_name] = oc.get_resource_info_by_jsonpath("build",build_name,current_project,json_path='{.status.phase}')
        if not "Complete" in builds[build_name]:
            print(build_name,':',builds[build_name])
            raise AssertionError
        else:
            print(build_name,':',builds[build_name])