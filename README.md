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

**WARNING: this is still work in progress**

The code may not build, there can be errors, and it can even destroy your hard disk or send you in another dimension.

Documentation is also work in progress....

# `ops`, the next generation

`ops` is the OpenServerless all-mighty CLI tool.

It is basically the [task](https://taskfile.dev) tool enhanced to support:

- a bunch of embedded commands (check tools) including `wsk` 
- the ability to download other tools
- a predefined set of tasks downloaded from github
- a way to create a hierarchy of taskfiles 
- documentation for tasks powered by [docopt](http://docopt.org/)

Note that to avoid an egg and chicken problem, `ops` itself is built with his ancestor, `task`.

- Build it with just `task build`.
- Run tests with `task test`.

# Documentation

## Environment variables

The following environment variables allows to ovverride certain defaults.

- `OPS_HOME` is the home dir, defaults to `~/.nuv` if not defined
- `NUV_ROOT` is the folder where `ops` looks for its tasks. If not defined, if will first look in current directory for an `olaris` folder otherwise download it from githut fron the `NUV_REPO` with a git clone
- `OPS_BIN` is the folder where `ops` looks for binaries (external command line tools). If not defined, it defaults to `~/.nuv/<os>-<arch>/bin`. All the prerequisites are downloaded in this directory
- `OPS_CMD` is the actual command executed - defaults to the absolute path of the target of the symbolic link but it can be overriden.
- `NUV_REPO` is the github repo where `ops` downloads its tasks. If not defined, it defaults to `https://github.com/apache/openserverless-task`.
- `NUV_BRANCH` is the branch where `nuv` looks for its tasks. The branch to use is defined at build time and it is the base version (without the patch level). Chech branch.txt for the current value
- `NUV_VERSION` can be defined to set nuv's version value. It is useful to override version validations when updating tasks (and you know what you are doing).  Current value is in version.txt
- `NUV_TMP` is a temporary folder where you can store temp files - defaults to `~/.nuv/tmp` 
- `NUV_APIHOST` is the host for `nuv -login`. It is used in place of the first argument of `nuv -login`. If empty, the command will expect the first argument to be the apihost.
- `NUV_USER`: set the username for `nuv -login`. The default is `nuvolaris`. It can be overriden by passing the username as an argument to `nuv -login` or by setting the environment variable.
- `NUV_PASSWORD`: set the password for `nuv -login`. If not set, `nuv -login` will prompt for the password. It is useful for tests and non-interactive environments.
- `NUV_PWD` is the folder where `ops` is executed (the current working directory). It is used to preserve the original working directory when `ops` is used again in tasks (e.g. nuv -realpath to retrieve the correct path). Change it only if you know what you are doing!
- `NUV_ROOT_PLUGIN` is the folder where `nuv` looks for plugins. If not defined, it defaults to the same directory where  `ops` is located.
- `NUV_OLARIS` holds the head commit hash of the used olaris repo. You can see the hash with `nuv -info`.
- `NUV_PORT` is the port where `ops` will run embedde web server for the configurator. If not defined, it defaults to `9678`.
- `NUV_NO_NUVOPTS` can be defined to disable nuvopts parsing. Useful to test hidden tasks. When this is enabled it also shows all the tasks instead of just those with a description.
- `OPS_NO_PREREQ` disable downloading of prerequisites - you have to ensure at least coreutils is in the path to make things work

## Special variables used for tests and debugging

## Where `nuv` looks for binaries 

Nuv requires some binary command line tools to work with. Those binaries are automatically downloaded for the supported platforms.

## Where `nuv` looks for tasks

Ops is an enhanced task runner that execute **taskfiles**.  How tasks are written is  described in [task](https://taskfile.dev) documentation.

Nuv is able either to run existing tasks or download them from github.

When you run `ops [<args>...]` it will first look for its `ops` root. 

The `ops` root is a folder with two files in it: `nuvfile.yml` (a yaml taskfile) and `nuvroot.json` (a json file with release information).

The first step is to locate the root folder. The algorithm to find the tools is the following.

If the environment variable `NUV_ROOT` is defined, it will look there first, and will check if there are the two files.

Then it will look in the current folder if there is a `nuvfile.yml`. If there is, it will also look for `nuvroot.json`. If it is not there, it will go up one level looking for a directory with `nuvfile.yml` and `nuvtools.json`, and selects it as the `nuv` root.

If there is not a `nuvfile.yml` it will look for a folder called `olaris` with both a `nuvfile.yml` and `nuvtools.json` in it and will select it as the `nuv` root.

Then it will look in `~/.nuv` if there is an `olaris` folder with `nuvfile.yml` and `nuvroot.json`.

If everything fails, it will ask you to download some tasks with the command `nuv -update`. In this case it will download the latest version.

## Where `ops` download tasks from GitHub

Download tasks from GitHub is triggered by the `ops -update` command.

The repo to use is defined by the environment variable `NUV_REPO`, and defaults if it is missing to `https://github.com/apache/openserverless-task`

The branch to use is defined at build time. It can be overriden with the enviroment variable `NUV_BRANCH`.

When you run `ops -update`, if there is not a `~/.nuv/<branch>/olaris` it will clone the current branch, otherwise it will update it.

## How `ops` execute tasks

It will then look to the command line parameters `ops <arg1> <arg2> <arg3>` and will consider them directory names. The list can be empty. 

If there is a directory name  `<arg1>` it will change to that directory. If there is then a subdirectory `<arg2>` it will change to that and so on until it finds a argument that is not a directory name. 

If the last argument is a directory name, will look for a `nuvopts.txt`. If it is there, it will show this. If it's not there, it will execute a `ops -t nuvfile.yml -l` showing tasks with description. 

If it finds an argument not corresponding to a directory, it will consider it a task to execute, 

If there is not a `nuvopts.txt`, it will execute as a task, passing the other arguments (equivalent to `task -t nuvfile.yml <arg> -- <the-other-args>`).

If there is a `nuvopts.txt`, it will interpret it as a  [`docopt`](http://docopt.org/) to parse the remaining arguments as parameters. The result of parsing is a sequence of `<key>=<value>` that will be fed to `task`. So it is equivalent to invoking `task -t nuvfile.yml <arg> <key>=<value> <key>=<value>...`

### Example

A command like `ops setup kubernetes install --context=k3s` will look in the folder `setup/kubernetes` in the ops root, if it is there, then select `install` as task to execute and parse the `--context=k3s`. It is equivalent to invoke `cd setup/kubernetes ; task install -- context=k3s`.

If there is a `nuvopts.txt` with a `command <name> --flag --fl=<val>` the passed parameters will be: `_name_=<name> __flag=true __fl=true _val_=<val>`.

Note that also this will also use the downloaded tools and the embedded commands of `ops`.

## Saving state

If you want to save values from a precedent execution to be provided as variables, simply write a file with a name starting and eding with `_`. 

Nuv will read all the `_*_` files assuming they are in a format `<key>=<value>`, will skip any line starting with `#` and add to the command line invoking task.

So if you have  a file `_server_` with:

```
# the server
_SERVER=myserver
# the user
_USER=myuser
```

at the end of `task` invocation there will be `_SERVER=myserver` `_USER=myuser`

## Embedded tools

Currently task embeds the following tools, and you can invoke them directly prefixing them with `-`: (`ops -task`, `nuv -basename` etc). Use `nuv -help` to list them.

- [task](https://taskfile.dev) the Task build tool
- [wsk](https://github.com/apache/openwhisk-cli) the OpenWhisk cli 

Basic unix like tools (`nuv -<tool> -help for details`):

- basename
- cat
- cp
- dirname
- grep
- gunzip
- gzip
- head
- ls
- mv
- pwd
- rm
- sleep
- tail
- tar
- tee
- touch
- tr
- unzip
- wc
- which
- zip

## Internal environment variables

Those variables are special purpose and used for test and debug

- `DEBUG` if set enable debugging messages
- `TRACE` when set gives mode detailes tracing informations, also enable DEBUG=1
- `EXTRA` appends extra arguments to the task invocation - useful when you need to set extra variables with a nuvopts active.
- `__OS` overrides the value of the detected operating system - useful to test prereq scripts
-  `__ARCH` overrides the value of the detected architecture - useful to test prereq scripts
- `OPS_RUNTIMES_JSON` is used for the values of the runtimes json if the system is unable to read in other ways the current runtimes json. It is normally compiled in when you buid from the current version of the runtimes.json. It can be overriden

- `OPS_COREUTILS` is a string, a space separated list, which lists all the commands the coreutils binary provided. It should be kept updated with the values of the current version used. It can be overriden

- `OPS_TOOLS` is a string, a space separated list, which lists all the commands provided as internal tool by the ops binary. It shold be kept updated with the current list of tools provided. It can be overriden defining externally

