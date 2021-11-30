import json
from smoke.features.steps.openshift import Openshift
oc = Openshift()

@given(u'environment variables {env_vars} are set')
def given_env_vars(context, env_vars):
    context.given_env_vars = json.loads(env_vars)

@then(u'We set env var {key} to value {value} in deploymentconfig {dc_name}')
def set_env(context, dc_name, key, value):
        print( "Passed env var key:'{}', value:'{}'".format(key, value))
        exit_code, output = oc.set_env_for_deployment_config(dc_name, context.current_project, key, value)

@then(u'We check that JENKINS_PASSWORD environement variable is set to {jenkins_password_env_value}')
def verify_jenkins_password_env_value(context, jenkins_password_env_value):
        print( "Passed jenkins_password_env_value '{}'".format(jenkins_password_env_value))
        # grep -w does exact match search
        exec_command = "env | grep -w JENKINS_PASSWORD | cut -f2 -d="
        jenkins_master_pod = oc.getmasterpod(context.current_project)
        result = oc.exec_in_pod(jenkins_master_pod,exec_command).strip()
        print( "Execution of command '{}' on pod '{}' returned '{}'".format( exec_command, jenkins_master_pod, result))
        if result != jenkins_password_env_value:
            raise AssertionError
        else:
             print('JENKINS_PASSWORD successfully set and found with value: ', result)

