import jenkins.model.Jenkins
import com.openshift.jenkins.plugins.OpenShiftClientTools

println "oc tools: configuring ..."

def descriptor = Jenkins.instance.getDescriptor(OpenShiftClientTools)

def toolsToAdd = [
    "oc-4.16": "/usr/share/openshift/bin/oc-416",
    "oc-4.17": "/usr/share/openshift/bin/oc-417",
    "oc-4.18": "/usr/share/openshift/bin/oc-418",
    "oc-4.19": "/usr/share/openshift/bin/oc-419"
]

def existing = descriptor.getInstallations() as List

toolsToAdd.each { name, path ->
    if (!existing.any { it.name == name }) {
        existing << new OpenShiftClientTools(name, path, [])
        println "oc tools: added: $name -> $path"
    } else {
        println "oc tools: skipped (already exists): $name"
    }
}

descriptor.setInstallations(existing.toArray(new OpenShiftClientTools[0]))
descriptor.save()

println "oc tools: all done!"
