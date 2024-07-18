
This is the procedure we followed to build the patched versions of:

 github.com/go-task/task -> github.com/sciabarracom/task
 mvdan.cc/sh -> github.com/sciabarracom/sh
 
1. fork mvdan.cc/sh in github.com/sciabarracom/sh
   fork github.com/go-task/task in github.com/sciabarracom/task

2. checkout v3.8.0 for sh and v3.38.0 for task as openserverless branch and set openserverless as the default branch

4. git submodule update --init, then add orig-auth to both with an authentication token so you can push it

5. bash patch-sh.sh
   bash patch-task.sk

to patch and push
