from behave import given, then, when

@when(u'true')
def step_impl(context):
    print( "fake when")

@then(u'fail')
def step_impl(context):
    raise AssertionError