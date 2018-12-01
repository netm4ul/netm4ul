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

/* ========================================= SELECTs ========================================= */

const selectProjects = `
SELECT id, name, description, updated_at
FROM projects;
`

/*
$1 : project name
*/
const selectProjectByName = `
SELECT id, name, description, updated_at
FROM projects
WHERE name = $1;
`

/*
$1 : domain name
$2 : project name
*/
const selectDomain = `
SELECT id, name, created_at, updated_at
FROM domains, projects
WHERE domains.project_id = projects.id
AND domains.name = $1
AND projects.name = $2;
`

/*
$1 : project name
*/
const selectDomains = `
SELECT domains.name, domains.created_at, domains.updated_at
FROM domains, projects
WHERE projects.name = $1
AND domains.project_id = projects.id; 
`

/*
Will return all subdomains too !
$1 : domain name
$2 : project name
*/
const selectDomainsAndSubdomain = `
SELECT id, name, created_at, updated_at
FROM domains, projects
WHERE name LIKE ('%.' || $1)
AND projects.name = $2
AND domains.project_id = projects.id; 
`

/*
$1 : user name
*/
const selectUserByName = `
SELECT id, name, password, token, created_at, updated_at
FROM users
WHERE name = $1;
`

/*
$1 : user token
*/
const selectUserByToken = `
SELECT id, name, password, token, created_at, updated_at
FROM users
WHERE token = $1;
`

/*
$1 : project name
*/
const selectIPsByProjectName = `
SELECT ips.id, ips.value, ips.network, ips.created_at, ips.updated_at
FROM ips, projects
WHERE ips.project_id = projects.ips
AND projects.name = $1;
`

/*
$1 : project name
$2 : ip value
*/
const selectIPByProjectName = `
SELECT ips.id, ips.value
FROM ips, projects
WHERE ips.project_id = projects.id
AND projects.name = $1
AND ips.value = $2;
`

/*
$1 : project name
*/
const selectPortsByProjectName = `
SELECT ports.id, ports.number, ports.protocol, ports.status, ports.banner, ports.type_id, ports.ip_id
FROM ports, ips, projects
WHERE ips.project_id = projects.id
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
WHERE ips.project_id = projects.id
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
WHERE ips.project_id = projects.id
AND projects.name = $1
AND ports.ip_id = ips.id
AND ips.value = $2
AND uris.port_id = ports.id
AND ports.number = $3;
`

/*
$1 : project name
*/
const selectRawsByProjectName = `
SELECT raws.id, raws.module, raws.project_id, raws.data, raws.created_at
FROM raws, projects
WHERE raws.project_id = projects.id
AND projects.name = $1;
`

/*
$1 : project name
$2 : module name
*/
const selectRawsByProjectNameAndModuleName = `
SELECT raws.id, raws.module, raws.project_id, raws.data, raws.created_at
FROM raws, projects
WHERE raws.project_id = projects.id
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
$2 : project_name
*/
const insertDomain = `
INSERT INTO domains (name, project_id)
VALUES ($1, (SELECT id FROM projects WHERE name = $2))
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
INSERT INTO ips (project_id, value, network)
VALUES (
	(SELECT id FROM projects WHERE name = $1),
	$2,
	$3
)
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
		WHERE ips.project_id = projects.id
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
		AND projects.id = ips.project_id
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
INSERT INTO raws(module, data, project_id)
VALUES ($1, $2, (SELECT id FROM projects WHERE name = $3))
returning id;
`

/* ========================================= UPDATE ========================================= */

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

/*
$1 ip value
$2 ip network
*/
const updateIP = `
UPDATE ips SET updated_at = now()
WHERE value = $1
AND network = $2;
`

/* ========================================= DELETE ========================================= */

/*
$1 user name
*/
const deleteUserByName = `
DELETE FROM users
WHERE name = $1;
`
