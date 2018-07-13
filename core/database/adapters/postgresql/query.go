package postgresql

// $1 : database name
const createDatabase = `CREATE DATABASE $1;`

const dropTableUsers = `DROP TABLE IF EXISTS users;`
const dropTableDomains = `DROP TABLE IF EXISTS domains;`
const dropTableProjects = `DROP TABLE IF EXISTS projects;`
const dropTableIPs = `DROP TABLE IF EXISTS ips;`
const dropTablePortTypes = `DROP TABLE IF EXISTS porttypes;`
const dropTablePorts = `DROP TABLE IF EXISTS ports;`
const dropTableURIs = `DROP TABLE IF EXISTS uris;`
const dropTableRaws = `DROP TABLE IF EXISTS raws;`

// ew, careful to SQLI there, the name must be trusted
const dropDatabase = `DROP DATABASE IF EXISTS %s;`
