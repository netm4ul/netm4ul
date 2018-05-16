# Netm4ul
Distributed recon &amp; pentest tool

Master
[![Build Status Master](https://travis-ci.org/netm4ul/netm4ul.svg?branch=master)](https://travis-ci.org/netm4ul/netm4ul)
Develop 
[![Build Status Develop](https://travis-ci.org/netm4ul/netm4ul.svg?branch=develop)](https://travis-ci.org/netm4ul/netm4ul)

## Usage

### Running NetM4ul with Docker (recommanded)

```
vim netm4ul.conf # change the credentials (db, api, etc...)
docker-compose up --build # you can add the "-d" option to detach your terminal, but you will need to "docker log" to see dev traces
```

This should download and run :
- 1 master node of NetM4ul (`netmaulserver`)
- 1 client node of NetM4ul (`netmaulclient`)
- 1 mongodb database with data stored in the `./data/db` directory

NOTE : If running with the provided `docker-compose.yml`, the `netm4ul.conf` **must** use the "container_name" of each service for their ip. (eg : set the database ip to "mongodb" in our case)

### Running without Docker

Requirement : 
- [Mongodb database](https://www.mongodb.com)
- [Go](https://golang.org/)
- [dep](https://github.com/golang/dep)

```
vim netm4ul.conf # change the credentials (db, api, etc...) and ip / ports
make

./netm4ul start server # in one terminal
./netm4ul start client # in another terminal

# netm4ul is running, you can control it directly with

./netm4ul run <somedomain,ip,ip range(CIDR)>
```


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


### Module

To write your module you will need to implement the Module interface (`modules/modules.go`) and put it into the `modules/<folder>` folders corresponding to your module.
All data produced by your module should write into the MongoDB database using the `WriteDB()` method.
`WriteDB()` **must** write data (at least) under its own collection and optionnaly update the global structure.

### Database

NetM4ul store all of it's data into a Mongodb database.

Database schema :

Project : 
```
{
    _id     : ObjectId("1234567890"),
    name    : "TestName",
    updated : new Date(),
    IP      : [
        ObjectID("22222222222222")
    ]
}
```

IPs :
```
{
    _id   : ObjectId("2222222222222")
    value : "10.0.0.1",
    port  : [
        ObjectId("33333333333")
    ],
    route : [

    ]
}
```

Ports :

```
{
    _id      : ObjectId("3333333333333"),
    number   : 443,
    protocol : "tcp",
    status   : "open",
    banner   : "Nginx 4.3.2.1",
    type     : "http"
    directories : [
        ObjectId("44444444444")
    ]
}
```

Directories :

```
{
    _id  : ObjectId("444444444444"),
    name : "index.php",
    code : 200
}
```

Route:

```
{
    _id : ObjectId("555555555555")
    source : "10.0.0.123",
    destionation : "10.0.0.1"
    hops : [
        {ip : 10.0.0.2, max : 10.123, min : 2.68, avg : 6.92},
        {ip : 192.168.56.1, max : 1.123, min : 0.68, avg : 0.92}
    ]
}
```

Example "final" schema (with all "joins")

```
{
    "google":{
        "IP" : [{
            "10.0.0.1": {
                "ports" : {
                    "tcp": {
                        "443": {
                            "status": "open",
                            "banner": "NGINX 4.3.2.1",
                            "type":"http"
                            "directories":{
                                "index.php":{"status":200}
                            }
                        },
                        "53": {
                            "status": "open",
                            "banner": "BIND 9",
                            "type":"dns"
                        }
                    },
                    "udp":{
                        "53": {
                            "status": "open",
                            "banner": "BIND 9",
                            "type" : "dns"
                        }
                    }
                },
                "route" :{
                    {"ip" : 10.0.0.2, "max" : 10.123, "min" : 2.68, "avg" : 6.92},
                    {"ip" : 192.168.56.1, "max" : 1.123, "min" : 0.68, "avg" : 0.92}
                }
            }
        }],
        "results": {
            "nmap":{raw},
            "traceroute":{raw},
            "dirb":{raw}
        }
    }

    "facebook":{
        "IP" : [{
            "10.0.0.1": {
                "ports" : {
                    "tcp": {
                        "443": {
                            "status": "open",
                            "banner": "NGINX 4.3.2.1",
                            "type":"http"
                            "directories":{
                                "index.php":{"status":200}
                            }
                        },
                    }
                }
            }
        }]
        "results": {
            "nmap":{raw},
            "traceroute":{raw},
            "dirb":{raw}
        }
    }
}



Database Schema :

{
    DB : netm4ul {
            Collection_1 = PROJECTS {
                Document_1 = Project_1 {
                    ID : P1_ObjectID
                    Name : "project_1"
                    IPs : [P1_IP1_ObjectID, P1_IP2_ObjectID, P1_IP3_ObjectID]
                }
                Document_2 = Project_2 {
                    ID : P2_ObjectID
                    Name : "project_2"
                    IPs : [P2_IP1_ObjectID, P2_IP2_ObjectID, P3_IP3_ObjectID]
                }
            }
            Collection_2 = IPS {
                Document_1 = P1_IP_1 {
                    ID : P1_IP1_ObjectID
                    Value : "4.4.4.4"
                    Ports : [P1_IP1_Port1_ObjectID, P1_IP1_Port2_ObjectID, P1_IP1_Port2_ObjectID]
                }
                Document_2 = P1_IP_2{
                    ID : P1_IP2_ObjectID
                    Value : "4.4.4.5"
                    Ports : [P1_IP2_Port1_ObjectID, P1_IP2_Port2_ObjectID, P1_IP2_Port3_ObjectID]
                }
            }
            Collection_3 = PORTS {
                Document_1 = P1_IP1_Port1{
                    ID : P1_IP1_Port1_ObjectID
                    Number : Port_nb
                    State : Filtered, Closed, Opened
                    Banner : Banner
                }
                Document_2 = P1_IP1_Port2{
                    ID : P1_IP1_Port2_ObjectID
                    Number : Port_nb
                    State : Filtered, Closed, Opened
                    Banner : Banner
                }
                Document_3 = P1_IP1_Port3{
                    ID : P1_IP1_Port3_ObjectID
                    Number : Port_nb
                    State : Filtered, Closed, Opened
                    Banner : Banner
                }
            }
        }
    }
```