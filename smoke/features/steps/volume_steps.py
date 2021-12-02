from smoke.features.steps.openshift import Openshift

oc = Openshift()

@then(u'persistentvolumeclaim "jenkins" created')
def verify_pvc(context):
    if not 'jenkins' in oc.search_resource_in_namespace('pvc','jenkins',context.current_project):
        raise AssertionError
    else:
        res = oc.search_resource_in_namespace('pvc','jenkins',context.current_project)
        print(res)


