## Testing the template based install of jenkins on openshift

### The smoke test written using behave framework for python has the below tree structure
<pre>
.
├── features
│   ├── environment.py
│   ├── jenkins-ephemeral.feature
│   ├── jenkins-ephemeral-without-oauth.feature
│   ├── jenkins-maven-agent.feature
│   ├── jenkins-nodejs-agent.feature
│   ├── jenkins-teardown-ephemeral.feature
│   ├── persistent-jenkins.feature
│   ├── persistent-maven-agent.feature
│   ├── persistent-nodejs-agent.feature
│   ├── persistent-test.feature
│   ├── README.md
│   ├── steps
│   │   ├── build_steps.py
│   │   ├── command.py
│   │   ├── debug.py
│   │   ├── delete_steps.py
│   │   ├── deployment_steps.py
│   │   ├── env_var_steps.py
│   │   ├── login_steps.py
│   │   ├── openshift.py
│   │   ├── plugins.py
│   │   ├── project.py
│   │   ├── steps.py
│   │   ├── template_steps.py
│   │   └── volume_steps.py
│   └── stress-test.feature
├── requirements.txt
└── samples
    ├── maven_pipeline.yaml
    └── nodejs_pipeline.yaml
</pre>
### Run smoke test on your local machine
<pre>
- oc login [cluster-server] -u username -p password --kubeconfig=kubeconfig
- export KUBECONFIG=kubeconfig
- make smoke
</pre>

### Test Results 
<pre>- The test results are JUnit files generated for each feature & are collected in out dir post test run is complete
</pre>
