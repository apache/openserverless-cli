cd "$(dirname $0)"
git ls-remote https://github.com/sciabarracom/task | awk '/refs\/heads\/openserverless/{print $1}' >_task.hash
git ls-remote https://github.com/sciabarracom/openwhisk-cli | awk '/refs\/heads\/openserverless/{print $1}' >_wsk.hash
cd ..
go get github.com/sciabarracom/task/v3@$(cat patches/_task.hash)
go get github.com/sciabarracom/openwhisk-cli@$(cat patches/_wsk.hash)
go mod tidy
