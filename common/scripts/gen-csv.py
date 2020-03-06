#!/usr/bin/env python

from ruamel.yaml import YAML
import argparse 
  
yaml=YAML(typ='safe') 
# Initialize parser 
parser = argparse.ArgumentParser() 
  
# Adding optional argument 
parser.add_argument("-v", "--version", help = "Version") 
  
# Read arguments from command line 
args = parser.parse_args()   
version = args.version

workdir = "deploy/olm-catalog/ibm-auditlogging-operator/" + version + "/"
csv = workdir + "ibm-auditlogging-operator.v3.5.0.clusterserviceversion.yaml"

with open(csv, 'r') as stream:
    try:
        loaded = yaml.load(stream)
    except yaml.YAMLError as exc:
        print(exc)

# Modify the fields from the dict
print(loaded['metadata']['annotations']['alm-examples'])
loaded['spec']['customresourcedefinitions']['owned'].append(
    {
        'description': 'AuditPolicy is the Schema for the auditpolicies API', 
        'kind': 'AuditPolicy',
        'name': 'auditpolicies.audit.policies.ibm.com',
        'version': 'v1alpha1',
        'displayName': 'AuditPolicy'
    }
)

# Save it again
with open(csv, 'w') as stream:
    try:
        stream.write(yaml.dump(loaded, default_flow_style=False))
    except yaml.YAMLError as exc:
        print(exc)
