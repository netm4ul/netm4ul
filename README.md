# netm4ul
Distributed recon &amp; pentest tool

## Usage

### Master node

Run one master node instance

```
go build
./netm4ul -server
```

### Clients node

Run N client

```
vim netm4ul.conf # add the master node (ip, port) in the server section
go build # if necessary
./netm4ul
```
