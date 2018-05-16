# Netm4ul
Distributed recon &amp; pentest tool

Master
[![Build Status Master](https://travis-ci.org/netm4ul/netm4ul.svg?branch=master)](https://travis-ci.org/netm4ul/netm4ul)
Develop 
[![Build Status Develop](https://travis-ci.org/netm4ul/netm4ul.svg?branch=develop)](https://travis-ci.org/netm4ul/netm4ul)

## Usage

```
git clone https://github.com/netm4ul/netm4ul
make
./netm4ul setup
# You may want to check the netm4ul.conf file to match your configuration.

./netm4ul start server # in one terminal
./netm4ul start client # in another terminal

# netm4ul is running, you can control it directly with

./netm4ul run <somedomain,ip,ip range(CIDR)>
```

See [Getting started](https://github.com/netm4ul/netm4ul/wiki/Getting-started) for more informations 


### CLI


Completion : 

bash : `source <(./netm4ul completion bash)`
zsh : `source <(./netm4ul completion zsh)`

```
Usage:
  netm4ul [flags]
  netm4ul [command]

Available Commands:
  completion  Generate autocompletion
  create      Create the requested ressource
  help        Help about any command
  list        Return all results
  run         Run scan on the defined target
  start       Start the requested service
  status      Show status of the requested service
  version     Prints version
```

Global flags : 

```
  -c, --config string   Custom config file path (default "netm4ul.conf")
  -h, --help            help for netm4ul
      --no-colors       Disable color printing
  -v, --verbose         verbose output
```

You can use -h on every subcommands.

## Contributing
[CONTRIBUTING.md](https://github.com/netm4ul/netm4ul/blob/master/CONTRIBUTING.md)

### Structure

### Core

Located in the `core` folder, all the core components are there.
The `api` folder contains all the code for the REST api on the Master node.
The `server` folder contains all the code for recieving and storing data in the DB. It's in charge of balancing all the modules on each client node.
The `client` folder contains all the code for client connection to the master node.
The `session` folder contains all the code for handling current session (loaded modules...).
The `config` is used for parsing the config files.

#### API

The api is a HTTP REST API. It only serves json results with the `Content-Type: application/json`.
It uses the following format :

```
{
	"status": "success", // only "success" or "error"
	"code": CodeOK, // see the "code list" in godocs
	"message": "Some message", // required only in ERROR code. Not mandatory on success
	"data" : // any kind of json type : object, array, string, number, ... If error, no data are returned.
}
```

All the JSON fields are **lowercase** and most of them are omitted if empty.
For more information see [HTTP API](https://github.com/netm4ul/netm4ul/wiki/API)
