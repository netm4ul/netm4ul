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
./netm4ul -server # in one terminal
./netm4ul -client # in  another terminal
# netm4ul is running, you can control it directly with
./netm4ul -targets <somedomain,ip,ip range(CIDR)>
```


### CLI

```
  -client  : Set the node as client
  -config  : Custom config file path (default "netm4ul.conf")
  -mode    : Mode of execution. Simple alias to list of module. See the config file (default "stealth")
  -modules : List of modules executed
  -server  : Set the node as server
  -targets : List of targets, comma separated
  -verbose : Enable verbose output
  -version : Print the version
```

Example :

`./netm4ul -client -config netm4ul.custom.conf -verbose`

`./netm4ul -version`

`./netm4ul -targets 192.168.1.1,192.168.2.0/24,localhost.localdomain -modules traceroute`

## Contributing

### Structure

```
.
├── cmd
│   ├── api
│   │   └── api.go
│   ├── client
│   │   └── client.go
│   ├── config
│   │   ├── config.go
│   │   └── config_test.go
│   ├── handler.go
│   └── server
│       ├── database
│       │   ├── database.go
│       │   └── helpers.go
│       └── server.go
├── config
│   ├── dorks.conf
│   ├── sqlmap.conf
│   └── traceroute.conf
├── docker-compose.yml
├── Dockerfile
├── Gopkg.lock
├── Gopkg.toml
├── LICENSE
├── main.go
├── Makefile
├── modules
│   ├── exploit
│   ├── modules.go
│   ├── recon
│   │   └── traceroute
│   │       └── traceroute.go
│   └── report
├── netm4ul
├── netm4ul.conf
└── README.md
```

### Core

Located in the `cmd` folder, all the core components are there.
The `api` folder contains all the code for the REST api on the Master node.
The `server` folder contains all the code for recieving and storing data in the DB. It's in charge of balancing all the modules on each client node.
The `client` folder contains all the code for client connection to the master node.

#### API

The api is a HTTP REST API. It only serves json results with the `Content-Type: application/json`.
It uses the following format :

```
{
	"status": "success", // only "success" or "error"
	"code": 200, // see the "code list" below
	"message": "Some message", // required only in ERROR code. Not mandatory on success
	"data" : // any kind of json type : object, array, string, number, ... If error, no data are returned.
}
```

All the JSON fields are **lowercase** and most of them are omitted if empty.

**Code list**

| Number | Meaning             | Description                                    | Status  |
| ------ | ------------------- | ---------------------------------------------- | ------- |
| 200    | OK                  | Expected result                                | success |
| 404    | Not Found           | The requested item was not found on the server | error   |
| 998    | Database Error      | Some error occured with the database           | error   |
| 999    | Not Implemented Yet | This endpoint is not available yet             | error   |


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
```