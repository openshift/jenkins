# How to execute the script

### The smoke test written using behave framework for python has the below tree structure
 <pre><font color="#0087FF">.</font>
├── <font color="#0087FF">features</font>
│   ├── deployJenkinsOperator.feature
│   ├── environment.py
│   ├── README.md
│   ├── requirements.txt
│   └── <font color="#0087FF">steps</font>
│       ├── command.py
│       ├── openshift.py
│       ├── project.py
│       └── steps.py
└── logs2020-08-16.txt</pre>


### Install python dependencies
pip install -r smoke/features/requirements.txt

### Use this behave command to run the smoke/features behave code
export <kubeconfig>
behave ./smoke/features

### Logs are generated in a separate file as shown above in the tree
Example - Log file for test run on 16-AUG-2020 will be named as logs2020-08-16.txt & all the test log for the entire day will be appended to the same log file for the day and so on based on the date new log file will be created.

