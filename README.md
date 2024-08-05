<!--
  ~ Licensed to the Apache Software Foundation (ASF) under one
  ~ or more contributor license agreements.  See the NOTICE file
  ~ distributed with this work for additional information
  ~ regarding copyright ownership.  The ASF licenses this file
  ~ to you under the Apache License, Version 2.0 (the
  ~ "License"); you may not use this file except in compliance
  ~ with the License.  You may obtain a copy of the License at
  ~
  ~   http://www.apache.org/licenses/LICENSE-2.0
  ~
  ~ Unless required by applicable law or agreed to in writing,
  ~ software distributed under the License is distributed on an
  ~ "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
  ~ KIND, either express or implied.  See the License for the
  ~ specific language governing permissions and limitations
  ~ under the License.
  ~
-->

# `ops`, the Apache OpenServerless all-mighty CLI tool.

Quick install in Linux, MacOSX and Windows with WSL or GitBash:

```
curl -sL bit.ly/get-ops | bash
````

To be implemented: quick install in Windows with PowerShell 

```
irm bit.ly/get-ops-exe | iex
````

# What is `ops`?

Well, as you can guess it helps with operations: `ops` is the OpenServerless CLI.

It is a task executor on steroids. 

- it embeds [`task`](https://taskfile.dev))
- it embeds [`wsk`](https://github.com/apache/openwhisk-cli)
- it embeds a lot of other utility commands (check with `ops -help`)
- automatically download and updated commad line tools he neeeds as prerequisites
- automatically download and updated a predefined set of tasks  from github 
- taskfiles are organized in commands and subcommands, hiearachally
- taskfiles have options powered by [docopt](http://docopt.org/)
- it supports plugins

The predefined set of tasks are all you need to install and manage an OpenServerless cluster.

You can also use for your own purposes, if you want.

# Where to look for external documetation

Because `ops` is built on `task`, `docopt` and `wsk` you have to consult the following website for: 

- informations on the format of `opsfile.yml`: [taskfiles](https://taskfile.dev)
- informations on the format of `docopts.txt`: [docopts](http://docopt.org/)
- informations on the OpenWhisk cli (the underlying engine for OpenServerless): [wsk](https://github.com/apache/openwhisk-cli)

## Where `ops` looks for tasks

Ops is able either to run local tasks or download them from github. 

Tasks also can contains prerequisite binaries, and task will download them automatically and put then in the PATH.

Se below for details

### Local Taks

When you run `ops [<args>...]` it will first look for its `ops` root.  

The `olaris` root is a folder with two files in it: `opsfile.yml` (a yaml taskfile) and `opsroot.json` (a json file with release information).

The first step is to locate the root folder. The algorithm to find the tools is the following.

If the environment variable `OPS_ROOT` is defined, it will look there first, and will check if there are the two files.

Then it will look in the current folder if there is a `opsfile.yml`. If there is, it will also look for `opsroot.json`. If it is not there, it will go up one level looking for a directory with `opsfile.yml` and `opstools.json`, and selects it as the `ops` root.

If there is not a `opsfile.yml` it will look for a folder called `olaris` with both a `opsfile.yml` and `opstools.json` in it and will select it as the `ops` root.

Then it will look in `~/.ops` if there is an `olaris` folder with `opsfile.yml` and `opsroot.json`.

### Download Tasks

If the local task resolution fails, it will download its tasks from github. 

It is the same process that you can trigger manually with the command `ops -update`. 

## Where `ops` download tasks from GitHub

The repo to use is defined by the environment variable `OPS_REPO`, and defaults to `https://github.com/apache/openserverless-task`

The branch to use is defined at build time. It is noramlly named as the base version of the CLI.  It can be overriden with the enviroment variable `OPS_BRANCH`.

When you run `ops -update`, if there is not a `~/.ops/<branch>/olaris` it will clone the current branch, otherwise it will update it.

## How `ops` execute tasks

`Ops` will  look to the command line parameters `ops <arg1> <arg2> <arg3>` and will consider them as directory names. The list can be empty. 

If there is a directory name  `<arg1>` it will change to that directory. If there is then a subdirectory `<arg2>` it will change to that and so on until it finds a argument that is not a directory name. 

If the last argument is a directory name, will look for a `docopts.txt`. If it is there, it will show this. If it's not there, it will execute a `ops -task -t opsfile.yml -l` showing tasks with description. 

If it finds an argument not corresponding to a directory, it will consider it a task to execute, 

If there is not a `docopts.txt`, it will execute as a task, passing the other arguments (equivalent to `ops -task -t opsfile.yml <arg> -- <the-other-args>`).

If there is a `docopts.txt`, it will interpret it as a  [`docopt`](http://docopt.org/) to parse the remaining arguments as parameters. The result of parsing is a sequence of `<key>=<value>` that will be fed to `task`. So it is equivalent to invoking `task -t opsfile.yml <arg> <key>=<value> <key>=<value>...`

### Example

A command like `ops setup kubernetes install --context=k3s` will look in the folder `setup/kubernetes` in the ops root, if it is there, then select `install` as task to execute and parse the `--context=k3s`. It is equivalent to invoke `cd setup/kubernetes ; task install -- context=k3s`.

If there is a `docopts.txt` with a `command <name> --flag --fl=<val>` the passed parameters will be: `_name_=<name> __flag=true __fl=true _val_=<val>`.

Note that also this will also use the downloaded tools and the embedded commands of `ops`.

## Embedded tools

Currently task embeds the following tools, and you can invoke them directly prefixing them with `-`: (`ops -task`, `ops -basename` etc). Use `ops -help` to list them.

This is the list of the tools (it could be outdated, check with `ops -help`):

```
Available tools:
-awk
-base64
-config
-datefmt
-die
-echoif
-echoifempty
-echoifexists
-empty
-envsubst
-extract
-filetype
-gron
-help
-info
-jj
-jq
-login
-needupdate
-plugin
-random
-remove
-rename
-replace
-retry
-serve
-sh
-task
-update
-urlenc
-validate
-version
-wsk
```

## Environment variables for tasks

As a convenience, the system sets the following variables but you cannot override them:

- `OPS` is the actual command you are using, so you can refer to ops itself in opsfiles as `$OPS`.
- `OPS_PWD` is the folder where `ops` is executed (the current working directory). It is used to preserve the original working directory because `ops` changes to the acutal folder where the tasks are for execution.

The following environment variables are always set and you can  ovverride them.

- `OPS_CMD` is the actual command executed - defaults to the absolute path of the target of the symbolic link but it can be overriden. `OPS` will take this value.
- `OPS_VERSION` can be defined to set ops's version value. It is useful to override version validations when updating tasks (and you know what you are doing).  Current value is defined at build time and stored in sources in version.txt
- `OPS_HOME` is the home dir, defaults to `~/.ops` 
- `OPS_REPO` is the github repo where `ops` downloads its tasks. If not defined, it defaults to `https://github.com/apache/openserverless-task`.
- `OPS_BRANCH` is the branch where `ops` looks for its tasks. The branch to use is defined at build time and it is ususally the base version (without the patch level). Check `branch.txt` for the current value
- `OPS_ROOT` is the folder where `ops` looks for its tasks. If not defined, if will follow the algirithm described before to finding it locally. Otherwise download it from github, git clones or git updates the `$OPS_REPO` in the `$OPS_BRANCH` and store it is `$OPS_HOME/$OPS_BRANCH/olaris`

- `OPS_BIN` is the folder where `ops` looks for binaries (external command line tools). If not defined, it defaults to `~/.ops/{{.OS}}-{{.ARCH}}/bin`. All the prerequisites are downloaded in this directory. It is automatically added to the PATH at the beginning when executing opsfiles.
- `OPS_TMP` is a temporary folder where you can store temp files - defaults to `~/.ops/tmp` 
- `OPS_APIHOST` is the host for `ops -login`. It is used in place of the first argument of `ops -login`. If empty, the command will expect the first argument to be the apihost.
- `OPS_USER` is set the username for `ops -login`. The default is `nuvolaris`. It can be overriden by passing the username as an argument to `ops -login` or by setting the environment variable.
- `OPS_PASSWORD`: set the password for `ops -login`. If not set, `ops -login` will prompt for the password. It is useful for tests and non-interactive environments.
- `OPS_ROOT_PLUGIN` is the folder where `ops` looks for plugins. If not defined, it defaults to the same directory where  `ops` is located.
- `OPS_PORT` is the port where `ops` will run embedded web server for the configurator. If not defined, it defaults to `9678`.
- `OPS_OLARIS` holds the head commit hash of the used olaris repo. If it is a local version its value is `<local>`. You can see the hash with `ops -info`.

## Special purpose environment variables

The following variables have a special purpose and used for test and debug

- `DEBUG` if set enable debugging messages
- `TRACE` when set gives mode detailes tracing informations, also enable DEBUG=1
- `EXTRA` appends extra arguments to the task invocation - useful when you need to set extra variables with a docopts active.
- `__OS` overrides the value of the detected operating system - useful to test prereq scripts
- `__ARCH` overrides the value of the detected architecture - useful to test prereq scripts
- `OPS_RUNTIMES_JSON` is used for the values of the runtimes json if the system is unable to read in other ways the current runtimes json. It is normally compiled in when you buid from the current version of the runtimes.json. It can be overriden
- `OPS_COREUTILS` is a string, a space separated list, which lists all the commands the coreutils binary provided. It should be kept updated with the values of the current version used. It can be overriden
- `OPS_TOOLS` is a string, a space separated list, which lists all the commands provided as internal tool by the ops binary. It shold be kept updated with the current list of tools provided. It can be overriden defining externally
- `OPS_NO_DOCOPTS` can be defined to disable docopts parsing. Useful to test hidden tasks. When this is enabled it also shows all the tasks instead of just those with a description.
- `OPS_NO_PREREQ` disable downloading of prerequisites - you have to ensure at least coreutils is in the path to make things work
