# netm4ul
Distributed recon &amp; pentest tool

## Usage

### Master node

Run one master node instance

```
make
./netm4ul -server
```

### Clients node

Run N client

```
vim netm4ul.conf # add the master node (ip, port) in the server section
make
./netm4ul
```

## Contributing

### Module

To write your module you will need to implement the Module interface (`modules/modules.go`) and put it into the `modules/<folder>` folders corresponding to your module.
All data produced by your module should write into the MongoDB database using the `WriteDB()` method.
`WriteDB()` must write data (at least) under its own collection and optionnaly update the global structure.

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