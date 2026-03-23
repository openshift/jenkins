package e2e

import "fmt"

const ocVersionPipeline = `pipeline {
	agent any
	stages {
		stage('oc') {
			steps {
				sh 'oc version'
			}
		}
	}
}`

const ocVersionWithToolTemplate = `pipeline {
	agent any
	tools {
		oc '%s'
	}
	stages {
		stage('oc') {
			steps {
				sh 'oc version'
			}
		}
	}
}`

func ocVersionWithToolPipeline(ocTool string) string {
	return fmt.Sprintf(ocVersionWithToolTemplate, ocTool)
}

const smokeTestPipeline = `pipeline {
	agent any
	stages {
		stage('smoke') {
			steps {
				echo 'Jenkins smoke test passed'
				sh 'whoami'
				sh 'oc version'
			}
		}
	}
}`

const testJobXML = `<?xml version='1.0' encoding='UTF-8'?>
<project>
  <description>Smoke test job created via REST API</description>
  <keepDependencies>false</keepDependencies>
  <properties/>
  <scm class="hudson.scm.NullSCM"/>
  <canRoam>true</canRoam>
  <disabled>false</disabled>
  <blockBuildWhenDownstreamBuilding>false</blockBuildWhenDownstreamBuilding>
  <blockBuildWhenUpstreamBuilding>false</blockBuildWhenUpstreamBuilding>
  <triggers/>
  <concurrentBuild>false</concurrentBuild>
  <builders>
    <hudson.tasks.Shell>
      <command>echo "hello from smoke test job"</command>
    </hudson.tasks.Shell>
  </builders>
  <publishers/>
  <buildWrappers/>
</project>`

var basePlugins = []string{
	"credentials",
	"durable-task",
	"cloudbees-folder",
	"git",
	"git-client",
	"plain-credentials",
	"scm-api",
	"script-security",
	"structs",
}
