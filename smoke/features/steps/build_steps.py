import time
import urllib3
from behave import given, then, when
from hamcrest import assert_that, equal_to
from smoke.features.steps.openshift import Openshift

oc = Openshift()
# Path to pipeline job to test agent images
maven_template ='./smoke/samples/maven_pipeline.yaml'
nodejs_template = './smoke/samples/nodejs_pipeline.yaml'
buildconfigs = {'sample-pipeline':'1','openshift-jee-sample':'1'}
builds = {}
def triggerbuild(buildconfig,namespace):
    print('Triggering build: ',buildconfig)
    res = oc.start_build(buildconfig,namespace)
    print(res)

@when(u'The user create objects from the "smoke/samples/nodejs_pipeline.yaml" template by processing the template and piping the output to oc create')
def createPipeline(context):
    res = oc.oc_process_template(nodejs_template)
    print(res)
@then('we check that the resources are created')
def check_resources(context):
    model = getattr(context, "model", None)
    if not model:
        context.model = Openshift()
    for row in context.table:
        res = context.model.search_resource_in_namespace(row["resource"], row["resource_name"],context.current_project)
        assert_that(row["resource_name"],equal_to(res))
@then(u'Trigger the build using "oc start-build sample-pipeline"')
def startbuild(context):
    triggerbuild('sample-pipeline',context.current_project)

@then(u'verify the build status of "sample-pipeline-1" build is Complete')
def verifysamplepipeline(context):
    verify_status(context.current_project, 'build', 'sample-pipeline-1', 5, 600, 'Complete')
@then(u'verify the build status of "nodejs-postgresql-example-1" build is Complete')
def verifynodejsBuildStatus(context):
    verify_status(context.current_project, 'build', 'nodejs-postgresql-example-1', 5, 600, 'Complete')

@then(u'route nodejs-postgresql-example must be created and be accessible')
def connectApp(context):
    print('Getting application route/url')
    app_name = 'nodejs-postgresql-example'
    verify_status(context.current_project, 'route', 'nodejs-postgresql-example', 2, 10, 'True', json_path='{.status.ingress[*].conditions[*].status}')
    route = oc.get_route_host(app_name,context.current_project)
    url = 'http://'+str(route)
    print('--->App url:')
    print(url)
    http = urllib3.PoolManager()
    res = http.request_encode_url('GET',url)
    connection_status = res.status
    count = 1
    while(count <= 30):
        res = http.request_encode_url('GET',url)
        connection_status = res.status
        if connection_status == 200:
            print('---> Application is accessible via the route')
            print(url)
            http.clear()
            break
        else:
            time.sleep(2)
            count+=1
            print("Url: {0}, return code: {1}, res: {2}".format(url,connection_status, res))
    if connection_status != 200:
        raise AssertionError

@when(u'The user create objects from the sample maven template by processing the template and piping the output to oc create')
def createMavenTemplate(context):
    res = oc.oc_process_template(maven_template)
    print(res)
@then(u'Trigger the build using oc start-build openshift-jee-sample')
def startBuild(context):
    triggerbuild('openshift-jee-sample',context.current_project)

@then(u'verify the build status of openshift-jee-sample-docker build is Complete')
def verifyDockerBuildStatus(context):
    verify_status(context.current_project, 'build', 'openshift-jee-sample-docker-1', 10, 900, 'Complete', '{.status.phase}')

@then(u'verify the build status of openshift-jee-sample-1 is Complete')
def verifyJenkinsBuildStatus(context):
    verify_status(context.current_project, 'build', 'openshift-jee-sample-1', 10, 900, 'Complete', '{.status.phase}')

@then(u'We check for deployment pod status to be "Completed"')
def deploymentPodStatus(context):
    verify_status(context.current_project, 'pod', 'jenkins-1-deploy', poll_interval_seconds=2, max_retries=60, expected_status='Succeeded')

@then(u'we check the pvc status is "Bound"')
def pvc_status(context):
    print('---------Getting pvc status---------')
    verify_status(context.current_project, 'pvc', 'jenkins', poll_interval_seconds=1, max_retries=10, expected_status='Bound')

def verify_status(namespace, object_type, object_name, poll_interval_seconds, max_retries, expected_status,json_path='{.status.phase}'):
    count = 1
    print('Getting {object_type} status for {object_name}')
    while(count <= max_retries):
        status = oc.get_resource_info_by_jsonpath(object_type,object_name,namespace,json_path)
        if expected_status in status:
            break
        time.sleep(poll_interval_seconds)
        count+=1
    print("{object_type} {object_name} status:{status}")
    if not expected_status in status:
        raise AssertionError

@then(u'verify the JaveEE application is accessible via route openshift-jee-sample')
def pingApp(context):
    print('Getting application route/url')
    app_name = 'openshift-jee-sample'
    route = oc.get_route_host(app_name,context.current_project)
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

@then(u'We rsh into the master pod and check the jobs count')
def getjobcount(context):
    for jobnames,_ in buildconfigs.items():
        exec_command = 'cat /var/lib/jenkins/jobs/'+context.current_project+'/jobs/'+context.current_project+'-'+jobnames+'/nextBuildNumber'
        jenkins_master_pod = oc.getmasterpod(context.current_project)
        count = oc.exec_in_pod(jenkins_master_pod,exec_command)
        buildconfigs[jobnames] = str(count)
    print(buildconfigs)

@when(u'We delete the jenkins master pod')
def deletemaster(context):
    master_pod = oc.getmasterpod(context.current_project)
    res = oc.delete("pods",master_pod,context.current_project)
    time.sleep(2)
    if res == None:
        raise AssertionError

@then(u'We rsh into the master pod & Compare if the data persist or is lost upon pod restart')
def comparejobs(context):
    for jobnames,_ in buildconfigs.items():
        master_pod = oc.getmasterpod(context.current_project)
        exec_command = 'cat /var/lib/jenkins/jobs/'+context.current_project+'/jobs/'+context.current_project+'-'+jobnames+'/nextBuildNumber'
        count = oc.exec_in_pod(master_pod,exec_command)
        buildconfigs[jobnames] = str(count)
    
    for jobnames, _ in buildconfigs.items():
        if(buildconfigs[jobnames] == '1'):
            print("Data doesnt persist")
            raise AssertionError
    print(buildconfigs)


@when(u'We Trigger multiple builds using oc start-build openshift-jee-sample')
def trigger_builds(context, max_builds=5):
    global builds
    count = 1
    ## creating a dictionary of builds that keeps a track of {buildname: build_status}
         # This will be used to check the build reconcilation
    while(count <= max_builds):
        triggerbuild('openshift-jee-sample',context.current_project)
        build_name ='openshift-jee-sample-' + str(count)
        builds[build_name] = None
        count+=1
    
@when(u'We scale down the pod count in the replication controller to "0" from "1"')
def scale_pod(context):
    rc_name = 'jenkins-1'
    oc.scaleReplicas(context.current_project,0,rc_name)
    replicas = oc.get_resource_info_by_jsonpath("dc","jenkins",context.current_project,json_path='{.status.availableReplicas}')
    if not '0' in replicas:
        raise AssertionError
    else:
        print('There are ',replicas,' running pods of jenkins')
    
@then(u'We delete some builds')
def delete_builds(context):
    global builds
    rm_build = ['openshift-jee-sample-2','openshift-jee-sample-4']
    for build_name in builds.keys():
        builds[build_name] = oc.get_resource_info_by_jsonpath("build",build_name,context.current_project,json_path='{.status.phase}')
    print("------------Fetching all builds and build status------------")
    print(builds)
    print("------------Deleting a few  builds------------")
    for items in rm_build:
        res = oc.delete("build",items,context.current_project)
        print(res)
        builds.pop(items)
    print("------------Fetching all builds and build status------------")
    print(builds)


@then(u'verify sync plugin is able to reconcile the build state and delete the job runs associated with the builds we deleted')
def ensure_all_builds_get_completed(context, poll_interval=10, max_retries=60):
    retries = 0
    while( retries < max_retries):
        time.sleep(poll_interval)
        for build_name in builds.keys():
            builds[build_name] = oc.get_resource_info_by_jsonpath("build",build_name,context.current_project,json_path='{.status.phase}')
            if not ( "Complete" or "New" in builds[build_name]):
                print('Build ', build_name,' was found with status:', builds[build_name])
                raise AssertionError
            else:
                print(build_name,':',builds[build_name])
        build_statuses = ['Complete']
        build_statuses.extend(builds.values())
        if all(status == build_statuses[0] for status in build_statuses):
            print( 'All builds are in Complete state, returning')
            break
        retries+=1
    if retries >= max_retries:
        print( 'At least one build was not in Complete state after:', builds)
        raise AssertionError
