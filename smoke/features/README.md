## Testing the template based install of jenkins on openshift

### The smoke test written using behave framework for python has the below tree structure
<pre><font color="#0000FF"><b>smoke</b></font>
├── <font color="#0000FF"><b>features</b></font>
│   ├── environment.py
│   ├── jenkins-ephemeral.feature
│   ├── maven-agent.feature
│   ├── nodejs-agent.feature
│   ├── plugins.feature
│   ├── README.md
│   └── <font color="#0000FF"><b>steps</b></font>
│       ├── command.py
│       ├── openshift.py
│       ├── plugins.py
│       ├── project.py
│       └── steps.py
├── requirements.txt
└── <font color="#0000FF"><b>samples</b></font>
    ├── maven_pipeline.yaml
    └── nodejs_pipeline.yaml</pre>


### Run the smoke test

<pre>- oc login to/the/openshift/cluster -u username -p password --kubeconfig=kubeconfig
- export KUBECONFIG=kubeconfig
- make smoke</pre>

### Test results 
<pre>- The test results are JUnit files generated for each feature & are collected in out dir post test run is complete
</pre>
