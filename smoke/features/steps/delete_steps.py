from smoke.features.steps.openshift import Openshift
from kubernetes import client, config

oc = Openshift()
v1 = client.CoreV1Api()

@then(u'we delete deploymentconfig.apps.openshift.io "jenkins"')
def del_dc(context):
    res = oc.delete("deploymentconfig","jenkins",context.current_project)
    if res == None:
        raise AssertionError

@then(u'we delete route.route.openshift.io "jenkins"')
def del_route(context):
    res = oc.delete("route","jenkins",context.current_project)
    if res == None:
        raise AssertionError

@then(u'delete configmap "jenkins-trusted-ca-bundle"')
def del_cm(context):
    res = oc.delete("configmap","jenkins-trusted-ca-bundle",context.current_project)
    if res == None:
        raise AssertionError

@then(u'delete serviceaccount "jenkins"')
def del_sa(context):
    res = oc.delete("serviceaccount","jenkins",context.current_project)
    if res == None:
        raise AssertionError

@then(u'delete rolebinding.authorization.openshift.io "jenkins_edit"')
def del_rb(context):
    res = oc.delete("rolebinding","jenkins_edit",context.current_project)
    if res == None:
        raise AssertionError

@then(u'delete service "jenkins"')
def del_svc(context):
    res = oc.delete("service","jenkins",context.current_project)
    if res == None:
        raise AssertionError

@then(u'delete service "jenkins-jnlp"')
def del_svc_jnlp(context):
    res = oc.delete("service","jenkins-jnlp",context.current_project)
    if res == None:
        raise AssertionError

@then(u'delete all buildconfigs')
def del_bc(context):
    res = oc.delete("bc","--all",context.current_project)
    if res == None:
        raise AssertionError

@then(u'delete all builds')
def del_builds(context):
    res = oc.delete("builds","--all",context.current_project)
    if res == None:
        raise AssertionError

@then(u'delete all deploymentconfig')
def del_alldc(context):
    res = oc.delete("deploymentconfig","--all",context.current_project)
    if res == None:
        raise AssertionError
@then(u'delete all services')
def del_allsvc(context):
    res = oc.delete("service","--all",context.current_project)
    if res == None:
        raise AssertionError

@then(u'delete all imagestream')
def del_all_is(context):
    res = oc.delete("is","--all",context.current_project)
    if res == None:
        raise AssertionError


@then(u'delete all remaining test resources')
@given(u'cleared from all test resources')
def del_all_remaining_test_resources(context):
    delete_command = "all,rolebindings.authorization.openshift.io,bc,cm,is,pvc,sa,secret"
    oc.delete(delete_command,"-l app=jenkins-ephemeral",context.current_project)
    oc.delete(delete_command,"-l app=jenkins-persistent",context.current_project)
    oc.delete(delete_command,"-l app=openshift-jee-sample",context.current_project)
    oc.delete(delete_command,"-l app=jenkins-pipeline-example",context.current_project)