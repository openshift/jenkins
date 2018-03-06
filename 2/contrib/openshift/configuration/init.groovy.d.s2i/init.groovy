import jenkins.model.Jenkins
Jenkins.getInstance().getPluginManager().doCheckUpdatesServer()
Jenkins.getInstance().getUpdateCenter().getCoreSource().getData()
