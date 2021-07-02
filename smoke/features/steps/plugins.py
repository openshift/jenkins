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

