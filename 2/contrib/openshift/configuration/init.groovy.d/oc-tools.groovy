import jenkins.model.Jenkins
import com.openshift.jenkins.plugins.OpenShiftClientTools

println "oc tools: configuring ..."

def descriptor = Jenkins.instance.getDescriptor(OpenShiftClientTools)

def toolsToAdd = [
    "oc-4.14": "/usr/share/openshift/bin/oc-414",
    "oc-4.13": "/usr/share/openshift/bin/oc-413",
    "oc-4.12": "/usr/share/openshift/bin/oc-412"
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
