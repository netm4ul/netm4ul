package postgresql

/* ========================================= CREATE DATABASE ====================================== */
const createDatabase = `
CREATE DATABASE $1;
`

/* ========================================= CREATE TABLE ========================================= */
const createTableProjects = `
CREATE TABLE IF NOT EXISTS projects (
	id serial UNIQUE,
	name text PRIMARY KEY,
	description text NULL,
	created_at timestamp with time zone DEFAULT current_timestamp,
	updated_at timestamp with time zone DEFAULT current_timestamp
);
`

const createTableUsers = `
CREATE TABLE IF NOT EXISTS users(
	id serial PRIMARY KEY,
	name varchar(50) NOT NULL,
	password varchar(255) NOT NULL,
	token varchar(255) NOT NULL,
	created_at timestamp with time zone DEFAULT current_timestamp,
	updated_at timestamp with time zone DEFAULT current_timestamp
);
`

const createTableIPs = `
CREATE TABLE IF NOT EXISTS ips(
	id serial PRIMARY KEY,
	value varchar(50) NOT NULL,
	project_name text references projects(name)
);
`

const createTablePortTypes = `
CREATE TABLE IF NOT EXISTS porttypes(
	id serial PRIMARY KEY,
	name varchar(50),
	description text
);
`

const createTablePorts = `
CREATE TABLE IF NOT EXISTS ports(
	id serial PRIMARY KEY,
	number integer NOT NULL,
	protocol varchar(10) NOT NULL,
	status varchar(10) NOT NULL,
	banner text,
	type_id serial references porttypes(id),
	ip_id serial references ips(id),
	UNIQUE(number, type_id, ip_id)
);
`

const createTableURIs = `
CREATE TABLE IF NOT EXISTS uris(
	id serial PRIMARY KEY,
	name text NOT NULL,
	code varchar(100),
	port_id serial references ports(id),
	UNIQUE(port_id, name)
);
`

const createTableRaws = `
CREATE TABLE IF NOT EXISTS raws(
	id serial PRIMARY KEY,
	module varchar(100) NOT NULL,
	project_name text references projects(name),
	data json NOT NULL,
	created_at timestamp with time zone DEFAULT current_timestamp
);
`

/* ========================================= SELECTs ========================================= */

const selectProjects = `
SELECT id, name, description, updated_at
FROM projects;
`
const selectProjectByName = `
SELECT id, name, description, updated_at
FROM projects
WHERE name = $1;
`

const selectUserByName = `
SELECT id, name, password, token, created_at, updated_at
FROM users
WHERE name = $1;
`

const selectUserByToken = `
SELECT id, name, password, token, created_at, updated_at
FROM users
WHERE token = $1;
`

/*
$1 : project name
*/
const selectIPsByProjectName = `
SELECT ips.id, ips.value
FROM ips, projects
WHERE ips.project_name = project.name
AND project.name = $1;
`

/*
$1 : project name
$2 : ip value
*/
const selectIPByProjectName = `
SELECT ips.id, ips.value
FROM ips, projects
WHERE ips.project_name = project.name
AND project.name = $1
AND ips.value = $2;
`

/*
$1 : project name
*/
const selectPortsByProjectName = `
SELECT ports.id, ports.number, ports.protocol, ports.status, ports.banner, ports.type_id, ports.ip_id
FROM ports, ips, projects
WHERE ips.project_name = projects.name
AND projects.name = $1
AND ports.ip_id = ips.id;
`

/*
$1 : project name
$2 : ip value
*/
const selectPortsByProjectNameAndIP = `
SELECT ports.id, ports.number, ports.protocol, ports.status, ports.banner, ports.type_id, ports.ip_id
FROM ports, ips, projects
WHERE ips.project_name = projects.name
AND projects.name = $1
AND ports.ip_id = ips.id
AND ips.value = $2;
`

/*
$1 : project name
$2 : ip value
$3 : port number
*/
const selectURIsByProjectNameAndIPAndPort = `
SELECT uris.id, uris.name, uris.code
FROM uris, ports, ips, projects
WHERE ips.project_name = projects.name
AND project.name = $1
AND ports.ip_id = ips.id
AND ips.value = $2
AND uris.port_id = ports.id
AND ports.number = $3;
`

/*
$1 : project name
*/
const selectRawsByProjectName = `
SELECT raws.id, raws.module, raws.project_name, raws.data, raws.created_at
FROM raws, projects
WHERE raws.project_name = projects.name
AND projects.name = $1;
`

/*
$1 : project name
$2 : module name
*/
const selectRawsByProjectNameAndModuleName = `
SELECT raws.id, raws.module, raws.project_name, raws.data, raws.created_at
FROM raws, projects
WHERE raws.project_name = projects.name
AND projects.name = $1
AND raws.module = $2;
`

/* ========================================= INSERT ========================================= */

/*
$1 : project name
$2 : project description
*/
const insertProject = `
INSERT INTO projects (name, description)
VALUES ($1, $2)
returning id;
`

/*
$1 : name
$2 : password (bcrypt please !)
$3 : token
*/
const insertUser = `
INSERT INTO users (name, password, token)
VALUES ($1, $2, $3)
returning id;
`

/*
insertIP will insert a new IP in the database
$1 : value
$2 : project name
*/
const insertIP = `
INSERT INTO ips (value, project_name)
VALUES ($1, $2)
returning id;
`

/*
insertPort will insert a new port in the db
$1 : number
$2 : protocol
$3 : status
$4 : banner
$5 : type
$6 : project name
$7 : ip value
*/
const insertPort = `
INSERT INTO ports (number, protocol, status, banner, type_id, ip_id)
VALUES ($1, $2, $3, $4,
	(
		SELECT id
		FROM porttypes
		WHERE name = $5
	),
	(
		SELECT ips.id
		FROM ips, projects
		WHERE ips.project_name = projects.name
		AND projects.name = $6
		AND ips.value = $7
	)
)
returning id;
`

/*
insertURI
$1 : uri name
$2 : uri code
$3 : port value
$4 : ip value
$5 : project value
*/
const insertURI = `
INSERT INTO uris(name, code, port_id)
VALUE ($1, $2, 
	(
		SELECT ports.id
		FROM ports, projects, ips
		WHERE ports.ip_id = ips.id
		AND projects.name = ips.project_name
		AND ports.number = $3
		AND ips.value = $4
		AND projects.name = $5
	)
)
returning id;
`

/*
insertRaw will insert raw input as JSON in the database
$1 : module name
$2 : json data (must be valid json)
$3 : project name
*/
const insertRaw = `
INSERT INTO raws(module, data, project_name)
VALUES ($1, $2, $3)
returning id;
`

/* UPDATE */

/*
$1 : token
$2 : username
*/
const updateUserTokenByUsername = `
UPDATE users SET token = $1
WHERE name = $2;
`

/*
$1 : password
$2 : username
*/
const updateUserPasswordByUsername = `
UPDATE users SET password = $1
WHERE name = $2;
`

/* ========================================= DELETE ========================================= */

/*
$1 user name
*/
const deleteUserByName = `
DELETE FROM users
WHERE name = $1;
`

/* ========================================== DROP ========================================== */
const dropTableUsers = `DROP TABLE users;`
const dropTableProjects = `DROP TABLE projects;`
const dropTableIps = `DROP TABLE ips;`
const dropTablePorttypes = `DROP TABLE porttypes;`
const dropTablePorts = `DROP TABLE ports;`
const dropTableUris = `DROP TABLE uris;`
const dropTableRaws = `DROP TABLE raws;`
