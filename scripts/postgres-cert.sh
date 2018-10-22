#!/bin/bash

CA_DIR=./ca

mkdir $CA_DIR

openssl req -new -text -subj /CN=Netm4ul -out $CA_DIR/database_server.req
openssl rsa -in $CA_DIR/database_server_privkey.pem -out $CA_DIR/database_server.key
openssl req -x509 -in $CA_DIR/database_server.req -text -key $CA_DIR/database_server_privkey.key -out $CA_DIR/database_server.crt
chmod og-rwx $CA_DIR/database_server.key