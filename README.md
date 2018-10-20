# Netm4ul
Distributed recon &amp; pentest tool

Master
[![Build Status Master](https://travis-ci.org/netm4ul/netm4ul.svg?branch=master)](https://travis-ci.org/netm4ul/netm4ul)
Develop 
[![Build Status Develop](https://travis-ci.org/netm4ul/netm4ul.svg?branch=develop)](https://travis-ci.org/netm4ul/netm4ul)

[![GoDoc](https://godoc.org/github.com/netm4ul/netm4ul?status.svg)](https://godoc.org/github.com/netm4ul/netm4ul)

## Usage

To understand a little bit more of the project, you might want to look at the `documentation/Architecture.md` file.

There is a single binary for the Controller, Agent, CLI client, API...

### Running NetM4ul with Docker (recommended)

You can install Docker on common distribution (ubuntu / debian) with the script at `https://get.docker.com`.
The full procedure can be found [here](https://docs.docker.com/install/)

You will need docker-compose to simplify the launch of all parts of the application. You can find the installation instruction [here](https://docs.docker.com/compose/install/)

```
git clone https://github.com/netm4ul/netm4ul
cd netm4ul
vim netm4ul.conf # change the credentials (db, api, etc...)
docker-compose up --build # you can add the "-d" option to detach your terminal, but you will need to "docker log" to see dev traces
```

This should download and run locally :
- 1 Controller node of NetM4ul (named `netmaulserver`)
- 1 Agent node of NetM4ul (named `netmaulclient`)
- 1 Database with data stored in the `./data/db` directory

NOTES :
- If running with the provided `docker-compose.yml`, the `netm4ul.conf` **must** use the "container_name" of each service for their ip. (eg : set the database ip to "postgres" in our case) 
  
- We have another repository to enables deployement with Ansible.
  
- We might add a sample k8s deployment file soon.


### Running without Docker
In the following example, a "terminal" could be a single host or multiple. You will need to download the same binary for each of them. Just modify the `netm4ul.conf` accordingly

Requirement : 
- Download the executable from the release section of Github.
  
Optionnal database:
- [PostgreSQL database](https://www.postgresql.org/)

```
# !!! ensure the database is running if you plan to use one. !!!

./netm4ul setup # It will help setup all the things.

./netm4ul start server # in one terminal
./netm4ul start client # in another terminal

# netm4ul is running, you can control it directly with

./netm4ul run <somedomain,ip,ip range(CIDR)>
```

See [Getting started](https://github.com/netm4ul/netm4ul/wiki/Getting-started) for more informations 


### CLI

To help newcomers we provide a complete autocompletion.

To enable it, you will need to execute one of the following command.

for bash : `source <(./netm4ul completion bash)`

for zsh : `source <(./netm4ul completion zsh)`

```
netm4ul : Distributed recon made easy

Usage:
  netm4ul [flags]
  netm4ul [command]

Available Commands:
  completion  Generate autocompletion
  create      Create the requested ressource
  help        Help about any command
  list        Return all results
  report      Generate a new report
  run         Run scan on the defined target
  setup       NetM4ul setup
  start       Start the requested service
  status      Show status of the requested service
  version     Prints version

Flags:
  -c, --config string    Custom config file path (default "netm4ul.conf")
  -h, --help             help for netm4ul
      --no-colors        Disable color printing
  -p, --project string   Uses the provided project name
  -v, --verbose          verbose output

Use "netm4ul [command] --help" for more information about a command.

```

You can use `-h` or `--help` on every subcommands.

## Structure

The general architecture of Netm4ul looks like this :
![general architecture schema](documentations/schema/general.png?raw=true)

### Core

All the core components are located in the `core` directory.

The `api/` folder contains all the code for the REST api on the Controller node.

The `server/` folder contains all the code for recieving and storing data in the DB. It's in charge of balancing all the modules on each client node.

The `client/` folder contains all the code for Agent connection to the Controller node.

The `session/` folder contains all the code for handling current session (loaded modules...).

The `config/` is used for parsing the config files.

### API

The api is a HTTP REST API. It only serves JSON results with the `Content-Type: application/json`.
It uses the following format :

```
{
	"status": "success", // only "success" or "error"
	"code": CodeOK, // see the "code list" in godocs
	"message": "Some message", // required only in ERROR code. Not mandatory on success
	"data" : // any kind of JSON type : object, array, string, number, ... If error, no data are returned.
}
```

All the JSON fields are **lowercase** and most of them are omitted if empty. (You can see all the field and possible result in the `core/database/models/models.go`)

The status code list is available in the `core/api/codes.go` file. If you plan to develop something, you might want to use the constant value representing them.


### Module

We are open to as many module as possible. Recon, Report, Exploit... are all welcome.

You can enable/disable each modules directly in the config file.

#### Recon

We currently support `nmap`, `masscan`, `traceroute` recon modules. `dns`, `shodan` are WIP and should soon be added.

#### Report

One report mode is available. More a to come. We aim to provides :

- [ ] Textual reports
- [ ] PDF
- [ ] Docs
- [ ] HTML

### Database

Netm4ul support multiple data storage backend.

For the moment : **PostgreSQL** is the prefered one.

Each database support is called an `adapter` and can be found in the `core/database/adapters` folder.

We currently support : `PostgreSQL`, storing to JSON file (`JsonDB` adapter) and a `testadapter` is provided for testing purpose only (it will not store anything).

The support for `MongoDB` is wanted but will need help to support it.

You can create a new adapters using the `netm4ul create adapter` command. It will generate all the boiler plate and place all the code in the good place. For more information see Developers.

#### PostgreSQL

The PostgreSQL backend is the recommended database backend for Netm4ul.
You can use the provided docker-compose to set it up easily.

If you want to just start the postgres server (without dockerized netm4ul Controller / Agents) with the docker-compose command, run `docker-compose up -d postgres`

Please, before using it, change the password writen in the `docker-compose.yml` file. (`POSTGRES_PASSWORD: password`)


## Developers

Netm4ul tries to be developer friend.

If you want to hack on it, you must download [Go](https://golang.org)
You will also need to get `dep` by running the command `got get -u https://github.com/golang/dep` (for other installation type, see [dep](https://github.com/golang/dep) )

We provide a `Makefile` to download dependencies, build and test the application.

You can run `make` in the netm4ul repository to build and install it.
You can run tests by running `make test`.

If you want to write a new module (Recon, Report, Exploit) or a new database adapter, you should use the command `netm4ul create`.
It will generate all the boilerplate needed to efficiently write new code.

For more information, follow the [CONTRIBUTING.md](https://github.com/netm4ul/netm4ul/blob/develop/CONTRIBUTING.md)


## Contributting

You can contribute to Netm4ul by openning pull requests, issues.

Contribution are not only for code. Non-code submission are *strongly* appreciated. (Spelling error? Missing documentation ? Missing example ? Better schema ?)

If you see a bug, please fill up an issues. We will try to fix it as soon as possible.

For more information, follow the [CONTRIBUTING.md](https://github.com/netm4ul/netm4ul/blob/develop/CONTRIBUTING.md)
