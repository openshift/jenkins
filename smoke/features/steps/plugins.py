import os

class Plugins:

    def getPlugins(self, filepath: str)-> dict:
        plugins = {}
        with open(filepath, 'r') as reader:
            for plugin in reader.readlines():
                key, value = plugin.split(":")
                version = value.replace("\n","")
                plugins[key]= version
        return plugins

    # def traversePod(self, filepath: str) -> dict:
    #     podPlugins = {}
    #     with open(filepath, 'r') as reader:
    #         for plugin in reader.readlines():
    #             key, value = 

p = Plugins()
core_plugins_path = "./2/contrib/openshift/base-plugins.txt"
core_plugins = p.getPlugins(core_plugins_path)
# pod_plugins = p.traversePod(core_plugins,)