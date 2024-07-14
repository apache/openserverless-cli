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

![example workflow](https://github.com/nuvolaris/nuv/actions/workflows/check.yml/badge.svg) ![Nuv Latest Release](https://badgen.net/github/release/nuvolaris/nuv/stable) ![License](https://badgen.net/github/license/nuvolaris/nuv) 

**WARNING: this is still work in progress**

The code may not build, there can be errors that will destroy your hard disk and so on.

Also documentation and code may not be aligned, as we first write documentation and then write code, so you may read documentation of code that does not yet exist!

You have been warned!

# `nuv`, the next generation

Nuv is the nuvolaris all-mighty build tool.

This is a rewrite of the current build tools to make it super powerful.

It is basically the [task](https://taskfile.dev) tool enhanced to support:

- a bunch of embedded commands (check tools) including `wsk` 
- the ability to download other tools
- a predefined set of tasks downloaded from github
- a way to create a hierarchy of taskfiles 
- documentation for tasks powered by [docopt](http://docopt.org/)

Note that to avoid an egg and chicken problem, `nuv` itself is built with his ancestor, `task`.

- Build it with just `task build`.
- Run tests with `task test`.

# Documentation

## Environment variables

The following environment variables allows to ovverride certain defaults.

- `DEBUG` and `TRACE` when set gives debugging and tracing informations, `EXTRA` appends extra arguments to the task invocation - useful when you need to set extra variables with a nuvopts active.
- `NUV_ROOT` is the folder where `nuv` looks for its tasks. If not defined, it follows the algorithm below to find it.
- `NUV_BIN` is the folder where `nuv` looks for binaries (external command line tools). If not defined, it defaults to the same directory where  `nuv` is located.
- `NUV_REPO` is the github repo where `nuv` downloads its tasks. If not defined, it defaults to `https://github.com/nuvolaris/olaris`.
- `NUV_BRANCH` is the branch where `nuv` looks for its tasks. The branch to use is defined at build time and it is the base version (without the patch level). For example, if `nuv` is `0.3.0-morpheus` the branch to use will be `0.3-morpheus`.
- `NUV_CMD` is the actual command executed - defaults to the absolute path of the target of the symbolic link but it can be overriden.
- `NUV_VERSION` can be defined to set nuv's version value. It is useful to override version validations when updating tasks (and you know what you are doing). 
- `NUV_NO_NUVOPTS` can be defined to disable nuvopts parsing. Useful to test hidden tasks. When this is enabled it also shows all the tasks instead of just those with a description.
- `NUV_PORT` is the port where `nuv` will run the web server. If not defined, it defaults to `9678`.
- `NUV_TMP` is a temporary folder where you can store temp files - defaults to `~/.nuv/tmp`
- `NUV_APIHOST` is the host for `nuv -login`. It is used in place of the first argument of `nuv -login`. If empty, the command will expect the first argument to be the apihost.
- `NUV_USER`: set the username for `nuv -login`. The default is `nuvolaris`. It can be overriden by passing the username as an argument to `nuv -login` or by setting the environment variable.
- `NUV_PASSWORD`: set the password for `nuv -login`. If not set, `nuv -login` will prompt for the password. It is useful for tests and non-interactive environments.
- `NUV_PWD` is the folder where `nuv` is executed (the current working directory). It is used to preserve the original working directory when `nuv` is used again in tasks (e.g. nuv -realpath to retrieve the correct path). Change it only if you know what you are doing!
- `NUV_ROOT_PLUGIN` is the folder where `nuv` looks for plugins. If not defined, it defaults to the same directory where  `nuv` is located.
- `NUV_OLARIS` holds the head commit hash of the used olaris repo. You can see the hash with `nuv -info`.
## Where `nuv` looks for binaries 
- `NUV_USE_COREUTILS` enables the use of [coreutils](https://github.com/uutils/coreutils) instead of the current unix tools implementation. They will eventually replace the current tools.

Nuv requires some binary command line tools to work with ("bins").

They are expected to be in the folder pointed by the environment variable `NUV_BIN`. 

If this environment variable is not defined, it defaults to the same folder where `nuv` itself is located. The `NUV_BIN` folder is then added to the beginning of the `PATH` before executing anything else.

Nuv is normally distributed with an installer that includes all the tools for the various operating systems (linux, windows, osx).

**NOTE**: You can download the relevant tools when you run from source code executing `task install`. This task will download the command line tools and setup a link in `/usr/local/bin` to invoke `nuv`.

## Where `nuv` looks for tasks

Nuv is an enhanced task runner. Tasks are described by [task](https://taskfile.dev) taskfiles.

Nuv is able either to run existing tasks or download them from github.

When you run `nuv [<args>...]` it will first look for its `nuv` root.

The `nuv` root is a folder with two files in it: `nuvfile.yml` (a yaml taskfile) and `nuvroot.json` (a json file with release information).

The first step is to locate the root folder. The algorithm to find the tools is the following.

If the environment variable `NUV_ROOT` is defined, it will look there first, and will check if there are the two files.

Then it will look in the current folder if there is a `nuvfile.yml`. If there is, it will also look for `nuvroot.json`. If it is not there, it will go up one level looking for a directory with `nuvfile.yml` and `nuvtools.json`, and selects it as the `nuv` root.

If there is not a `nuvfile.yml` it will look for a folder called `olaris` with both a `nuvfile.yml` and `nuvtools.json` in it and will select it as the `nuv` root.

Then it will look in `~/.nuv` if there is an `olaris` folder with `nuvfile.yml` and `nuvroot.json`.

Finally it will look in the `NUV_BIN` folder if there is an `olaris` folder with  `nuvfile.yml` and `nuvroot.json`

If everything fails, it will ask you to download some tasks with the command `nuv -update`. In this case it will download the latest version.

## Where `nuv` download tasks from GitHub

Download tasks from GitHub is triggered by the `nuv -update` command.

The repo to use is defined by the environment variable `NUV_REPO`, and defaults if it is missing to `https://github.com/nuvolaris/olaris`

The branch to use is defined at build time. It can be overriden with the enviroment variable `NUV_BRANCH`.

When you run `nuv -update`, if there is not a `~/.nuv/olaris` it will clone the current branch, otherwise it will update it.

## How `nuv` execute tasks

It will then look to the command line parameters `nuv <arg1> <arg2> <arg3>` and will consider them directory names. The list can be empty. 

If there is a directory name  `<arg1>` it will change to that directory. If there is then a subdirectory `<arg2>` it will change to that and so on until it finds a argument that is not a directory name. 

If the last argument is a directory name, will look for a `nuvopts.txt`. If it is there, it will show this. If it's not there, it will execute a `task -t nuvfile.yml -l` showing tasks with description. 

If it finds an argument not corresponding to a directory, it will consider it a task to execute, 

If there is not a `nuvopts.txt`, it will execute as a task, passing the other arguments (equivalent to `task -t nuvfile.yml <arg> -- <the-other-args>`).

If there is a `nuvopts.txt`, it will interpret it as a  [`docopt`](http://docopt.org/) to parse the remaining arguments as parameters. The result of parsing is a sequence of `<key>=<value>` that will be fed to `task`. So it is equivalent to invoking `task -t nuvfile.yml <arg> <key>=<value> <key>=<value>...`

### Example

A command like `nuv setup kubernetes install --context=k3s` will look in the folder `setup/kubernetes` in the `nuv root` if it is there, select `install` as task to execute and parse the `--context=k3s`. It is equivalent to invoke `cd setup/kubernetes ; task install -- context=k3s`.

If there is a `nuvopts.txt` with a `command <name> --flag --fl=<val>` the passed parameters will be: `_name_=<name> __flag=true __fl=true _val_=<val>`.

Note that also this will also use the downloaded tools and the embedded commands of `nuv`.

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

Currently task embeds the following tools, and you can invoke them directly prefixing them with `-`: (`nuv -task`, `nuv -basename` etc). Use `nuv -help` to list them.

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







