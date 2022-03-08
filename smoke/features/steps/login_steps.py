# @mark.steps
# ----------------------------------------------------------------------------
# STEPS:
# ----------------------------------------------------------------------------
import os

import jenkins
from smoke.features.steps.openshift import Openshift

oc = Openshift()

@given(u'we have a openshift cluster')
def loginCluster(context):
    print("Using [{}]".format(context.current_project))
    jenkins_master_pod = ""
    #jenkins_master_pod = getmasterpod(current_project)

# @then(u'We ensure that we can login to jenkins using {username} and {password}')
# def login_to_jenkins(context, username, password):
#     pass
#     # current_project = context.current_project
#     # host = "https://" + oc.get_route_host('jenkins', current_project)
#     # os.environ.setdefault("PYTHONHTTPSVERIFY", "0")
#     # server = jenkins.Jenkins(host, username, password)
#     # print( "Jenkins: ", server)
#     # if server == None:
#     #     raise AssertionError
#     # user = server.get_whoami()
#     # version = server.get_version()
#     # print('Hello %s from Jenkins %s' % (user['fullName'], version))


