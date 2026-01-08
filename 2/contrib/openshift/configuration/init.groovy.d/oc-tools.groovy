import jenkins.model.Jenkins
import hudson.slaves.EnvironmentVariablesNodeProperty
import hudson.slaves.EnvironmentVariablesNodeProperty.Entry
import com.openshift.jenkins.plugins.OpenShiftClientTools
import jenkins.model.Jenkins

println "Configuring node property ..."

def instance = Jenkins.instance

def nodeProps = instance.nodeProperties
def envProp = nodeProps.get(EnvironmentVariablesNodeProperty)

if (envProp == null) {
    def newEnv = new EnvironmentVariablesNodeProperty(
        new Entry("PATH+OC", "/usr/share/openshift/bin/oc-417"),
    )
    nodeProps.add(newEnv)
} else {
    def envVars = envProp.envVars

    envVars.put("PATH+OC", "/usr/share/openshift/bin/oc-417")
}

instance.save()

println "Environment variables updated."

println "Configuring oc tools ..."

def descriptor = Jenkins.instance.getDescriptor(OpenShiftClientTools)

def toolsToAdd = [
    "oc-412": "/usr/share/openshift/bin/oc-412",
    "oc-413": "/usr/share/openshift/bin/oc-413",
    "oc-414": "/usr/share/openshift/bin/oc-414",
    "oc-415": "/usr/share/openshift/bin/oc-415",
    "oc-416": "/usr/share/openshift/bin/oc-416",
    "oc-417": "/usr/share/openshift/bin/oc-417"
]

def existing = descriptor.getInstallations() as List

toolsToAdd.each { name, path ->
    if (!existing.any { it.name == name }) {
        existing << new OpenShiftClientTools(name, path, [])
        println "Added: $name -> $path"
    } else {
        println "Skipped (already exists): $name"
    }
}

descriptor.setInstallations(existing.toArray(new OpenShiftClientTools[0]))
descriptor.save()

println "All done!"
