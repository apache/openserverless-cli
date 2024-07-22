cd "$(dirname $0)"/..
#git ls-remote https://github.com/sciabarracom/task | awk '/refs\/heads\/openserverless/{print $1}' >_task.hash
#git ls-remote https://github.com/sciabarracom/openwhisk-cli | awk '/refs\/heads\/openserverless/{print $1}' >_wsk.hash
#go get github.com/sciabarracom/task/v3@$(cat _task.hash)
#go get github.com/sciabarracom/openwhisk-cli@$(cat _wsk.hash)
#go mod tidy


go clean -cache -modcache
git ls-remote https://github.com/sciabarracom/sh | awk '/refs\/heads\/openserverless/{print $1}' | tee _sh.hash
go get github.com/sciabarracom/sh/v3@$(cat _sh.hash)

git ls-remote https://github.com/sciabarracom/task | awk '/refs\/heads\/openserverless/{print $1}' | tee _task.hash
go get github.com/sciabarracom/task/v3@$(cat _task.hash)

git ls-remote https://github.com/sciabarracom/openwhisk-wskdeploy | awk '/refs\/heads\/openserverless/{print $1}' | tee _wskdeploy.hash
go get github.com/sciabarracom/openwhisk-wskdeploy@$(cat _wskdeploy.hash)

git ls-remote https://github.com/sciabarracom/openwhisk-cli | awk '/refs\/heads\/openserverless/{print $1}' | tee _wsk.hash

go get github.com/sciabarracom/openwhisk-cli@$(cat _wsk.hash)

go mod tidy
