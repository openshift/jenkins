import jenkins.model.Jenkins
import hudson.security.ACL;
import jenkins.security.NotReallyRoleSensitiveCallable;

ACL.impersonate(ACL.SYSTEM, new NotReallyRoleSensitiveCallable<Void, Exception>() {
    public Void call() throws Exception {
        Thread thread = new Thread(){
            public void run(){
                ACL.impersonate(ACL.SYSTEM, new NotReallyRoleSensitiveCallable<Void, Exception>() {
                    public Void call() throws Exception {
                            try {
                                Jenkins.getInstance().getPluginManager().doCheckUpdatesServer();
                            } catch (IOException e) {
                                e.printStackTrace();
                            }
                            Jenkins.getInstance().getUpdateCenter().getCoreSource().getData();
                        return null;
                    }
                });
            }
        };
        
        thread.start();
        return null;
    }
});

