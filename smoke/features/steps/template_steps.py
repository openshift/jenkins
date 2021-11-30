from smoke.features.steps.openshift import Openshift

oc = Openshift()

@when(u'User enters oc new-app jenkins-persistent command')
def persistentTemplate(context):
    res = oc.new_app('jenkins-persistent', context.current_project)
    if(res == None):
        print("Error while installing jenkins using persistent template")
        raise AssertionError

@when(u'User enters oc new-app jenkins-ephemeral command')
def ephemeralTemplate(context):
    res = oc.new_app('jenkins-ephemeral', context.current_project)
    if(res == None):
        print("Error while installing jenkins using ephemeral template")
        raise AssertionError

@when(u'User enters oc new-app jenkins-ephemeral command using env vars')
def ephemeral_template_with_env_vars(context):
    given_env_vars = context.given_env_vars
    command  = "jenkins-ephemeral "
    print( given_env_vars )
    for i in given_env_vars.items():
            print( " -e {}={} ".format(*i))
            command += " -e {}={} ".format(*i)
    print( "oc new-app command: {0}".format(command))
    res = oc.new_app(command, context.current_project)
    if(res == None):
        print("Error while installing jenkins using ephemeral template")
        raise AssertionError

