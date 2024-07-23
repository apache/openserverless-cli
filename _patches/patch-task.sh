#!/bin/sh
cd "$(dirname $0)"
# any suggestion how to avoid this rename and use just replaces in go.mod is welcome
HERE=$PWD
SHVER=$(git ls-remote https://github.com/sciabarracom/sh | awk '/refs\/heads\/openserverless/{print $1}')
STAG="v3.38.0"
DTAG="v3.38.6"
cd taskfile
go clean -cache -modcache
git checkout "$STAG" -B openserverless
mkdir -p cmd/taskmain
cat cmd/task/task.go \
| sed -e 's/package main/package taskmain/' \
| sed -e 's/func main()/func _main()/' \
| tee cmd/taskmain/task.go
sed -i -e 's/func init/func FlagInit/' internal/flags/flags.go
cat   <<EOF >>cmd/taskmain/task.go

func Task(_args []string) (int, error) {
   os.Args = _args
   flags.FlagInit()
	if err := run(); err != nil {
		l := &logger.Logger{
			Stdout:  os.Stdout,
			Stderr:  os.Stderr,
			Verbose: flags.Verbose,
			Color:   flags.Color,
		}
		if err, ok := err.(*errors.TaskRunError); ok && flags.ExitCode {
			l.Errf(logger.Red, "%v\n", err)
			return err.TaskExitCode(), err
		}
		if err, ok := err.(errors.TaskError); ok {
			l.Errf(logger.Red, "%v\n", err)
			return err.Code(), err
		}
		l.Errf(logger.Red, "%v\n", err)
		return errors.CodeUnknown, err
	}
	return errors.CodeOk, nil
}
EOF
#cp $HERE/task.go cmd/taskmain/task.go
git add cmd/taskmain/task.go
find . -name \*.go  | while read file 
do echo $file 
   sed -i "s!go-task/task!sciabarracom/task!" $file
   sed -i 's!mvdan.cc/sh!github.com/sciabarracom/sh!' $file
   sed -i 's!"Taskfile.!"opsfile.!' $file
   sed -i 's!task: !ops: !' $file
done
sed -i -e 's/go-task\/task/sciabarracom\/task/' go.mod
sed -i -e '/mvdan.cc/g' go.mod
go get github.com/sciabarracom/sh/v3@$SHVER
go mod tidy
git commit -m "patching sh for ops" -a
git tag $DTAG
go build
git push origin-auth openserverless -f --tags
VER=$(git rev-parse HEAD)
GOBIN=$HERE go install github.com/sciabarracom/task/v3/cmd/task@$VER

