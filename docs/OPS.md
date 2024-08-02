# Ops

Ops is the OpenServerless cli tool.

It embeds [Task](https://taskfile.dev) and  [OpenWhisk wsk](https://github.com/apache/openwhisk-cli).  Note that Task actually embeds a [shell interpreter](https://github.com/mvdan/sh) so most of the tasks are actully implemented as shell scripts embedded in taskfiles (YAML).

Most of its capabilites are implemented as a hiearchy of taskfiles, managed as described below. There is a preprocessing done on the command line allowing to find the correct taskfile, and parameters of each task are described using [docpts](http://docopt.org/)

To be able to develop ops you need either to know [posix shell scripting](https://pubs.opengroup.org/onlinepubs/9699919799/utilities/V3_chap02.html), [taskfile definition](https://taskfile.dev/usage/) and [docpts](http://docopt.org/).

The shell is modified to be able to test some commands in a special way to simplify writing the scripts in the taskfiles.

In particular, all the commonly used Unix utilites are delegated to an utility that embeds all of then called `coreutils`.  So in a task all the commands listed as `coreutils` like `rm`, `ls` are actually translated in an invocation of `coreutils rm`, `coreutils ls` and so all. All the commands delegated to coreutils can be seen with `ops -help`. We are using [uutils](https://github.com/uutils/coreutils) a rust portable rewrite of core unix coreutils.

Furthrmore some commands are implemented in `ops` itself. For example it embeds an `awk` intepreter (actually [gowak](https://github.com/benhoyt/goawk) and a JQ intepreter (actually [gojq](https://github.com/itchyny/gojq)). The shell is modified to be able to use those tooks `awk` and `jq` directly.

Some commands are actually external so there is support for automatically downloading prequisite binaries.

All the CLI commands are a hierarchy of taskfiles downloaded automatically from [https://github.com/apache/openserverless-task](https://github.com/apache/openserverless-task)

It then parse the command line to locate the correct taskfile and execute it.

Before executing each command, there is a prerequisite management system that is executed before actually running the commands, so the correct version of the utils are downloaded

# Ops command line

TODO describe the command line 

# Ops Preqrequisites

TODO describe prereq.yaml

