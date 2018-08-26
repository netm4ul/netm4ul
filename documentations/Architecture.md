# Architecture

- modulable
    - recon
    - exploit
    - report
- distributed
    - Docker containers
- extensible
    - postgres
    - json files
    - mongodb (partial implementation : need help)

- Nodes (same binary, just different cli argument)
    - Controller
        - Database
        - Load balancer
        - API 
    - Agent
        - Modules

Netm4ul is designed in a modulable way. The global architecture looks like this : 
![general architecture schema](schema/general.png?raw=true)

You can run multiple agent on the same host, and run the same kind agent on multiple host. They will automatically connect to the host (given that you correctly set up the ip/port and details in the config files)


## Agent <-> Controller protocol

- The connection is initiated by the Agent. The Agent has the controller IP in its config file. The server doesn't have the list of Agents before they connect to it.
- The Agent send a first packet (the "hello packet") following the `communication.Node` structure to the Controller. (file `core/communication/abi.go`)
- The Controller recieves it and store it in a map.
- The Controller MAY send requests to execute commands following the `communication.Command` structure. (file `core/communication/abi.go`)
- The Agent MAY now send some data packet with 

## REST API

> [WIP, missing features]

There is an HTTP API available on port 8080 by default (you can change it in the `netm4ul.conf`).
The rest API is the central control point. You must use it to launch scan and change the state of the application.

You can use the API to retrieve any data from the database. It will helps in the build of Web interface for example.

The API has an authentification mechanism : 
- Create a login / password combo. The application stores it, generate and return a random token.
- The client recieves and stores the token.
- Every request maybe by the client **MUST** provide the token with the header `X-Session-Token:longtokenvalue`. There are (at the moment) only 3 routes that accept unauthentificated request : `/`,`/users/create`,`/users/login`. (To be sure, up to date info : `core/api/middleware.go`).
- The server check if the token is valid (found in the database)

The flow of execution for creating an user has been schematized and looks like this :

![users creation schema](schema/create-user.png?raw=true)
