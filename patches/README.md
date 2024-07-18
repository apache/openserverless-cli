# Patches:

- changes to mvdan/sh:
   - rename to sciabarracom/sh
   - use coreutils as builtins
   - use ops tools as builtin
- changes to taskfile: 
   - rename to sciabarracom/task
    - default taskfile is opsfile.yml 
    - use the patched mvdan shell
- changes to openwhisk-wskdeploy:
    - rename to sciabarracom/openwhisk-wskdeploy
    - use a different path for retrieving the configuration (api/info instead of the root of the APIHOST)
    - wire a different default runtimesjson
- changes to openwhisk-wsk
    - rename to sciabarracom/openwhisk-wskdeploy
    - generate and add to the sources the translation resources to build as a library
    - use the patched openwhisk-wskdeploy
       
# Procedure

This is the procedure we followed to build the patched versions of:
 
1. fork mvdan.cc/sh in github.com/sciabarracom/sh
   fork github.com/go-task/task in github.com/sciabarracom/task
   fork github.com/apache/openwhisk-cli in github.com/sciabarracom/openwhisk-cli
   fork github.com/apache/openwhisk-wskdeploy in github.com/sciabarracom/openwhisk-wskdeploy

3. git submodule update --init, 
   then add remote orig-auth ian authentication token so you can push it  back
   then add remote upstream to the original repos to fetch tags
   then git fetch --all

3. execute the patch scripts in order

   bash patch-sh.sh
   bash patch-task.sh
   bash patch-wskdeploy.sh
   bash patch-wsk.sh
