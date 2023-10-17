package e2e

const (
	simpleSuccessfulPipeline = `
        node() {
          sh 'exit 0'
        }
`
	simpleFailedPipeline = `
        node() {
          sh 'exit 1'
        }
`
	pipelineWithEnvs = `
          node() {
            echo "FOO1 is ${env.FOO1}"
            echo "FOO2 is ${env.FOO2}"
          }
`
	samplepipeline = `
          try {
             timeout(time: 20, unit: 'MINUTES') {
                stage('build') {
                  openshift.withCluster() {
                     openshift.withProject() {
                        echo "Using project ${openshift.project()} in cluster with url ${openshift.cluster()}"
                     }
                  }
                }

             }
          } catch (err) {
             echo "in catch block"
             echo "Caught: ${err}"
             currentBuild.result = 'FAILURE'
             throw err
          }
`
	simplemaven2 = `
          try {
             timeout(time: 20, unit: 'MINUTES') {

                node("POD_TEMPLATE_NAME") {
                  container("java") {
                    sh "mvn --version"
                  }
                }

             }
          } catch (err) {
             echo "in catch block"
             echo "Caught: ${err}"
             currentBuild.result = 'FAILURE'
             throw err
          }
`

	simplemaven1 = `
         try {
            timeout(time: 20, unit: 'MINUTES') {

               node("POD_TEMPLATE_NAME") {
                  sh "mvn --version"
               }

            }
         } catch (err) {
            echo "in catch block"
            echo "Caught: ${err}"
            currentBuild.result = 'FAILURE'
            throw err
         }
`

	javabuilder = `
try {
  timeout(time: 20, unit: 'MINUTES') {

    node("java-builder") {
      container("java") {
        sh "mvn --version"
      }
    }
  }
} catch (err) {
  echo "in catch block"
  echo "Caught: ${err}"
  currentBuild.result = 'FAILURE'
  throw err
}
`
	nodejsbuilder = `
try {
  timeout(time: 20, unit: 'MINUTES') {

    node("nodejs-builder") {
      container("nodejs") {
        sh "npm --version"
      }
    }
  }
} catch (err) {
  echo "in catch block"
  echo "Caught: ${err}"
  currentBuild.result = 'FAILURE'
  throw err
}
`
	nodejsDeclarative = `
        // path of the template to use
        def templatePath = 'nodejs-postgresql-example'
        // name of the template that will be created
        def templateName = 'nodejs-postgresql-example'
        // NOTE, the "pipeline" directive/closure from the declarative pipeline syntax needs to include, or be nested outside,
        // and "openshift" directive/closure from the OpenShift Client Plugin for Jenkins.  Otherwise, the declarative pipeline engine
        // will not be fully engaged.
        pipeline {
            agent {
              node {
                // spin up a node.js slave pod to run this build on
                label 'nodejs-builder'
              }
            }
            options {
                // set a timeout of 20 minutes for this pipeline
                timeout(time: 20, unit: 'MINUTES')
            }

            stages {
                stage('preamble') {
                    steps {
                        script {
                            openshift.withCluster() {
                                openshift.withProject() {
                                    echo "Using project: ${openshift.project()}"
                                }
                            }
                        }
                    }
                }
                stage('cleanup') {
                    steps {
                        script {
                            openshift.withCluster() {
                                openshift.withProject() {
                                    // delete everything with this template label
                                    openshift.selector("all", [ template : templateName ]).delete()
                                    // delete any secrets with this template label
                                    if (openshift.selector("secrets", templateName).exists()) {
                                        openshift.selector("secrets", templateName).delete()
                                    }
                                }
                            }
                        } // script
                    } // steps
                } // stage
                stage('create') {
                    steps {
                        script {
                            openshift.withCluster() {
                                openshift.withProject() {
                                    // create a new application from the templatePath
                                    openshift.newApp(templatePath, "-p", "APPLICATION_DOMAIN=rails-%s.ocp.io")
                                }
                            }
                        } // script
                    } // steps
                } // stage
                stage('build') {
                    steps {
                        script {
                            openshift.withCluster() {
                                openshift.withProject() {
                                    def builds = openshift.selector("bc", templateName).related('builds')
                                    builds.untilEach(1) {
                                        return (it.object().status.phase == "Complete")
                                    }
                                }
                            }
                        } // script
                    } // steps
                } // stage
                stage('deploy') {
                    steps {
                        script {
                            openshift.withCluster() {
                                openshift.withProject() {
                                    def rm = openshift.selector("dc", templateName).rollout()
                                    openshift.selector("dc", templateName).related('pods').untilEach(1) {
                                        return (it.object().status.phase == "Running")
                                    }
                                }
                            }
                        } // script
                    } // steps
                } // stage
                stage('tag') {
                    steps {
                        script {
                            openshift.withCluster() {
                                openshift.withProject() {
                                    // if everything else succeeded, tag the ${templateName}:latest image as ${templateName}-staging:latest
                                    // a pipeline build config for the staging environment can watch for the ${templateName}-staging:latest
                                    // image to change and then deploy it to the staging environment
                                    openshift.tag("${templateName}:latest", "${templateName}-staging:latest")
                                }
                            }
                        } // script
                    } // steps
                } // stage
            } // stages
        } // pipeline
`
	bluegreenTemplateYAML = `
apiVersion: v1
kind: Template
labels:
  template: bluegreen-pipeline
message: A Jenkins server must be instantiated in this project to manage
  the Pipeline BuildConfig created by this template.  You will be able to log in to
  it using your OpenShift user credentials.
metadata:
  annotations:
    description: This example showcases a blue green deployment using a Jenkins
      pipeline that pauses for approval.
    iconClass: icon-jenkins
    tags: instant-app,jenkins
  name: bluegreen-pipeline
objects:
- apiVersion: v1
  kind: BuildConfig
  metadata:
    annotations:
      pipeline.alpha.openshift.io/uses: '[{"name": "${NAME}", "namespace": "", "kind": "DeploymentConfig"}]'
    creationTimestamp: null
    labels:
      name: bluegreen-pipeline
    name: bluegreen-pipeline
  spec:
    strategy:
      jenkinsPipelineStrategy:
        jenkinsfile: |-
          try {
             timeout(time: 20, unit: 'MINUTES') {
                def appName="${NAME}"
                def project=""
                def tag="blue"
                def altTag="green"
                def verbose="${VERBOSE}"

                node {
                  project = env.PROJECT_NAME
                  stage("Initialize") {
                    sh "oc get route ${appName} -n ${project} -o jsonpath='{ .spec.to.name }' --loglevel=4 > activeservice"
                    activeService = readFile('activeservice').trim()
                    if (activeService == "${appName}-blue") {
                      tag = "green"
                      altTag = "blue"
                    }
                    sh "oc get route ${tag}-${appName} -n ${project} -o jsonpath='{ .spec.host }' --loglevel=4 > routehost"
                    routeHost = readFile('routehost').trim()
                  }

                  openshift.withCluster() {
                    openshift.withProject() {
                      stage("Build") {
                        echo "building tag ${tag}"
                        def bld = openshift.startBuild("${appName}")
                        bld.untilEach {
                          return it.object().status.phase == "Running"
                        }
                        bld.logs('-f')
                      }

                      stage("Deploy Test") {
                        openshift.tag("${appName}:latest", "${appName}:${tag}")
                        def dc = openshift.selector('dc', "${appName}-${tag}")
                        dc.rollout().status()
                      }

                      stage("Test") {
                        input message: "Test deployment: http://${routeHost}. Approve?", id: "approval"
                      }

                      stage("Go Live") {
                        sh "oc set -n ${project} route-backends ${appName} ${appName}-${tag}=100 ${appName}-${altTag}=0 --loglevel=4"
                      }

                    }
                  }
                }
             }
          } catch (err) {
             echo "in catch block"
             echo "Caught: ${err}"
             currentBuild.result = 'FAILURE'
             throw err
          }
      type: JenkinsPipeline
    triggers:
    - github:
        secret: "${GITHUB_WEBHOOK_SECRET}"
      type: GitHub
    - generic:
        secret: "${GENERIC_WEBHOOK_SECRET}"
      type: Generic
- apiVersion: v1
  kind: Secret
  metadata:
    name: ${NAME}
  stringData:
    database-admin-password: ${DATABASE_ADMIN_PASSWORD}
    database-password: ${DATABASE_PASSWORD}
    database-user: ${DATABASE_USER}
- apiVersion: v1
  kind: Route
  metadata:
    name: blue-${NAME}
  spec:
    host: blue-${APPLICATION_DOMAIN}
    to:
      kind: Service
      name: ${NAME}-blue
- apiVersion: v1
  kind: Route
  metadata:
    name: green-${NAME}
  spec:
    host: green-${APPLICATION_DOMAIN}
    to:
      kind: Service
      name: ${NAME}-green
- apiVersion: v1
  kind: Route
  metadata:
    name: ${NAME}
  spec:
    alternateBackends:
    - name: ${NAME}-green
      weight: 0
    host: n-p-e-${APPLICATION_DOMAIN}
    to:
      kind: Service
      name: ${NAME}-blue
      weight: 100
- apiVersion: v1
  kind: ImageStream
  metadata:
    annotations:
      description: Keeps track of changes in the application image
    name: ${NAME}
- apiVersion: v1
  kind: BuildConfig
  metadata:
    annotations:
      description: Defines how to build the application
    name: ${NAME}
  spec:
    output:
      to:
        kind: ImageStreamTag
        name: ${NAME}:latest
    postCommit:
      script: npm test
    source:
      contextDir: ${CONTEXT_DIR}
      git:
        ref: ${SOURCE_REPOSITORY_REF}
        uri: ${SOURCE_REPOSITORY_URL}
      type: Git
    strategy:
      sourceStrategy:
        env:
        - name: NPM_MIRROR
          value: ${NPM_MIRROR}
        from:
          kind: ImageStreamTag
          name: nodejs:latest
          namespace: ${NAMESPACE}
      type: Source
    triggers:
    - github:
        secret: ${GITHUB_WEBHOOK_SECRET}
      type: GitHub
    - generic:
        secret: ${GENERIC_WEBHOOK_SECRET}
      type: Generic
- apiVersion: v1
  kind: Service
  metadata:
    annotations:
      service.alpha.openshift.io/dependencies: '[{"name": "${DATABASE_SERVICE_NAME}", "namespace": "", "kind": "Service"}]'
    name: ${NAME}-blue
  spec:
    ports:
    - name: web
      port: 8080
      targetPort: 8080
    selector:
      name: ${NAME}-blue
- apiVersion: v1
  kind: DeploymentConfig
  metadata:
    annotations:
      description: Defines how to deploy the application server
    name: ${NAME}-blue
  spec:
    replicas: 1
    selector:
      name: ${NAME}-blue
    strategy:
      type: Rolling
    template:
      metadata:
        labels:
          name: ${NAME}-blue
        name: ${NAME}-blue
      spec:
        containers:
        - env:
          - name: DATABASE_SERVICE_NAME
            value: ${DATABASE_SERVICE_NAME}
          - name: MYSQL_USER
            value: ${DATABASE_USER}
          - name: MYSQL_PASSWORD
            value: ${DATABASE_PASSWORD}
          - name: MYSQL_DATABASE
            value: ${DATABASE_NAME}
          - name: MYSQL_ROOT_PASSWORD
            value: ${DATABASE_ROOT_PASSWORD}
          image: ' '
          livenessProbe:
            httpGet:
              path: /pagecount
              port: 8080
            initialDelaySeconds: 30
            timeoutSeconds: 3
          name: nodejs-postgresql-example
          ports:
          - containerPort: 8080
          readinessProbe:
            httpGet:
              path: /pagecount
              port: 8080
            initialDelaySeconds: 3
            timeoutSeconds: 3
          resources:
            limits:
              memory: ${MEMORY_LIMIT}
    triggers:
    - imageChangeParams:
        automatic: true
        containerNames:
        - nodejs-postgresql-example
        from:
          kind: ImageStreamTag
          name: ${NAME}:blue
      type: ImageChange
    - type: ConfigChange
- apiVersion: v1
  kind: Service
  metadata:
    annotations:
      service.alpha.openshift.io/dependencies: '[{"name": "${DATABASE_SERVICE_NAME}", "namespace": "", "kind": "Service"}]'
    name: ${NAME}-green
  spec:
    ports:
    - name: web
      port: 8080
      targetPort: 8080
    selector:
      name: ${NAME}-green
- apiVersion: v1
  kind: DeploymentConfig
  metadata:
    annotations:
      description: Defines how to deploy the application server
    name: ${NAME}-green
  spec:
    replicas: 1
    selector:
      name: ${NAME}-green
    strategy:
      type: Rolling
    template:
      metadata:
        labels:
          name: ${NAME}-green
        name: ${NAME}-green
      spec:
        containers:
        - env:
            - name: POSTGRESQL_USER
              valueFrom:
                secretKeyRef:
                  key: database-user
                  name: ${NAME}
            - name: POSTGRESQL_PASSWORD
              valueFrom:
                secretKeyRef:
                  key: database-password
                  name: ${NAME}
            - name: POSTGRESQL_DATABASE
              value: ${DATABASE_NAME}
            - name: POSTGRESQL_ADMIN_PASSWORD
              valueFrom:
                secretKeyRef:
                  key: database-admin-password
                  name: ${NAME}
          image: ' '
          livenessProbe:
            httpGet:
              path: /pagecount
              port: 8080
            initialDelaySeconds: 30
            timeoutSeconds: 3
          name: nodejs-postgresql-example
          ports:
          - containerPort: 8080
          readinessProbe:
            httpGet:
              path: /pagecount
              port: 8080
            initialDelaySeconds: 3
            timeoutSeconds: 3
          resources:
            limits:
              memory: ${MEMORY_LIMIT}
    triggers:
    - imageChangeParams:
        automatic: true
        containerNames:
        - nodejs-postgresql-example
        from:
          kind: ImageStreamTag
          name: ${NAME}:green
      type: ImageChange
    - type: ConfigChange
- apiVersion: v1
  kind: Service
  metadata:
    annotations:
      description: Exposes the database server
    name: ${DATABASE_SERVICE_NAME}
  spec:
    ports:
    - name: postgresql
      port: 5432
      targetPort: 5432
    selector:
      name: ${DATABASE_SERVICE_NAME}
- apiVersion: v1
  kind: DeploymentConfig
  metadata:
    annotations:
      description: Defines how to deploy the database
    name: ${DATABASE_SERVICE_NAME}
  spec:
    replicas: 1
    selector:
      name: ${DATABASE_SERVICE_NAME}
    strategy:
      type: Recreate
    template:
      metadata:
        labels:
          name: ${DATABASE_SERVICE_NAME}
        name: ${DATABASE_SERVICE_NAME}
      spec:
        containers:
        - env:
            - name: POSTGRESQL_USER
              valueFrom:
                secretKeyRef:
                  key: database-user
                  name: ${NAME}
            - name: POSTGRESQL_PASSWORD
              valueFrom:
                secretKeyRef:
                  key: database-password
                  name: ${NAME}
            - name: POSTGRESQL_DATABASE
              value: ${DATABASE_NAME}
            - name: POSTGRESQL_ADMIN_PASSWORD
              valueFrom:
                secretKeyRef:
                  key: database-admin-password
                  name: ${NAME}
          image: ' '
          livenessProbe:
            initialDelaySeconds: 30
            tcpSocket:
              port: 5432
            timeoutSeconds: 1
          name: postgresql
          ports:
          - containerPort: 5432
          resources:
            limits:
              memory: ${MEMORY_MYSQL_LIMIT}
          volumeMounts:
          - mountPath: /var/lib/mysql/data
            name: ${DATABASE_SERVICE_NAME}-data
        volumes:
        - emptyDir:
            medium: ""
          name: ${DATABASE_SERVICE_NAME}-data
    triggers:
    - imageChangeParams:
        automatic: true
        containerNames:
        - postgresql
        from:
          kind: ImageStreamTag
          name: postgresql:${POSTGRESQL_VERSION}
          namespace: ${NAMESPACE}
      type: ImageChange
    - type: ConfigChange
parameters:
- description: The name assigned to all of the frontend objects defined in this template.
  displayName: Name
  name: NAME
  required: true
  value: nodejs-postgresql-example
- description: The exposed hostname that will route to the Node.js service, if left
    blank a value will be defaulted.
  displayName: Application Hostname
  name: APPLICATION_DOMAIN
- description: The URL of the repository with your application source code.
  displayName: Git Repository URL
  name: SOURCE_REPOSITORY_URL
  required: true
  value: https://github.com/openshift/nodejs-ex.git
- description: The reference of the repository with your application source code.
  displayName: Git Repository Ref
  name: SOURCE_REPOSITORY_REF
  required: true
  value: master
- description: Password for the database admin user.
  displayName: Database Administrator Password
  from: '[a-zA-Z0-9]{16}'
  generate: expression
  name: DATABASE_ADMIN_PASSWORD
- displayName: Database Name
  name: DATABASE_NAME
  required: true
  value: sampledb
- description: Username for postgresql user that will be used for accessing the database.
  displayName: postgresql Username
  from: user[A-Z0-9]{3}
  generate: expression
  name: DATABASE_USER
- description: Password for the postgresql user.
  displayName: postgresql Password
  from: '[a-zA-Z0-9]{16}'
  generate: expression
  name: DATABASE_PASSWORD
- description: Maximum amount of memory the Node.js container can use.
  displayName: Memory Limit
  name: MEMORY_LIMIT
  required: true
  value: 512Mi
- description: Maximum amount of memory the postgresql container can use.
  displayName: Memory Limit (postgresql)
  name: MEMORY_MYSQL_LIMIT
  required: true
  value: 512Mi
- displayName: Database Service Name
  name: DATABASE_SERVICE_NAME
  required: true
  value: postgresql
- description: Password for the database admin user.
  displayName: Database Administrator Password
  from: '[a-zA-Z0-9]{16}'
  generate: expression
  name: DATABASE_ROOT_PASSWORD
- description: Set this to the relative path to your project if it is not in the root
    of your repository.
  displayName: Context Directory
  name: CONTEXT_DIR
- description: Github trigger secret.  A difficult to guess string encoded as part of the webhook URL.  Not encrypted.
  displayName: GitHub Webhook Secret
  from: '[a-zA-Z0-9]{40}'
  generate: expression
  name: GITHUB_WEBHOOK_SECRET
- description: A secret string used to configure the Generic webhook.
  displayName: Generic Webhook Secret
  from: '[a-zA-Z0-9]{40}'
  generate: expression
  name: GENERIC_WEBHOOK_SECRET
- description: The custom NPM mirror URL
  displayName: Custom NPM Mirror URL
  name: NPM_MIRROR
- description: The OpenShift Namespace where the NodeJS and postgresql ImageStreams reside.
  displayName: Namespace
  name: NAMESPACE
  required: true
  value: openshift
- description: Whether to enable verbose logging of Jenkinsfile steps in pipeline
  displayName: Verbose
  name: VERBOSE
  required: true
  value: "false"`
)
