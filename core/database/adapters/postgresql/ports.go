package postgresql

import (
	"errors"
	"github.com/netm4ul/netm4ul/core/events"

	"github.com/jinzhu/gorm"
	"github.com/netm4ul/netm4ul/core/database/models"
	log "github.com/sirupsen/logrus"
)

// TOFIX : doing an actual join/relation insert instead of 3 f requests
func (pg *PostgreSQL) createOrUpdatePort(projectName string, ip string, port pgPort) error {
	// res := pg.db.Raw(insertPort, port.Number, port.Protocol, port.Status, port.Banner, port.Type, projectName, ip)

	// if res.Error != nil {
	// 	return errors.New("Could not create or update port : " + res.Error.Error())
	// }

	log.Debugf("createOrUpdate Port : %+v", port)

	var foundPort pgPort
	res := pg.db.Raw(selectPortsByProjectNameAndIP, projectName, ip).Scan(&foundPort)

	// insert port if it doesn't exist
	if gorm.IsRecordNotFoundError(res.Error) {
		return pg.createPort(projectName, ip, port)
	}

	// handle other errors
	if res.Error != nil {
		return errors.New("Could not select port : " + res.Error.Error())
	}

	err := pg.updatePort(projectName, ip, port)
	if err != nil {
		return errors.New("Could not update port : " + err.Error())
	}

	return nil
}

//CreateOrUpdatePort is the public wrapper to create or update a new port in the database.
func (pg *PostgreSQL) CreateOrUpdatePort(projectName string, ip string, port models.Port) error {

	pgp := pgPort{}
	pgp.FromModel(port)
	err := pg.createOrUpdatePort(projectName, ip, pgp)
	if err != nil {
		return err
	}

	return nil
}

func (pg *PostgreSQL) createPort(projectName string, ip string, port pgPort) error {
	//TOFIX
	res := pg.db.Debug().Create(&port)
	if res.Error != nil {
		return errors.New("Could not insert ip : " + res.Error.Error())
	}
	// res := pg.db.Exec(insertPort, port.Number, port.Protocol, port.Status, port.Banner, port.Type, projectName, ip)
	// if res.Error != nil {
	// 	return res.Error
	// }

	events.NewEventPort(port.ToModel())
	return nil
}

//CreatePort is the public wrapper to create a new port in the database.
func (pg *PostgreSQL) CreatePort(projectName string, ip string, port models.Port) error {
	pgp := pgPort{}
	pgp.FromModel(port)
	err := pg.createPort(projectName, ip, pgp)
	if err != nil {
		return err
	}
	return nil
}

func (pg *PostgreSQL) updatePort(projectName string, ip string, port pgPort) error {
	res := pg.db.Model(&port).Update(port)
	if res.Error != nil {
		return res.Error
	}
	return nil
}

//UpdatePort is the public wrapper to update a new port in the database.
func (pg *PostgreSQL) UpdatePort(projectName string, ip string, port models.Port) error {
	pgp := pgPort{}
	pgp.FromModel(port)
	err := pg.updatePort(projectName, ip, pgp)
	if err != nil {
		return err
	}
	return nil
}

func (pg *PostgreSQL) createOrUpdatePorts(projectName string, ip string, ports []pgPort) error {
	for _, port := range ports {
		err := pg.createOrUpdatePort(projectName, ip, port)
		if err != nil {
			return err
		}
	}
	return nil
}

//CreateOrUpdatePorts is the public wrapper to create or update multiple Port
// This function should be used instead of CreateOrUpdatePort during bulk inserts.
func (pg *PostgreSQL) CreateOrUpdatePorts(projectName string, ip string, ports []models.Port) error {
	for _, port := range ports {
		pgp := pgPort{}
		pgp.FromModel(port)
		err := pg.createOrUpdatePort(projectName, ip, pgp)
		if err != nil {
			return err
		}
	}
	return nil
}

func (pg *PostgreSQL) getPorts(projectName string, ip string) ([]pgPort, error) {

	ports := []pgPort{}
	res := pg.db.Raw(`
	SELECT ports.*
		FROM ports, ips, projects
		WHERE projects.name = ?
		AND ips.value = ?
		AND ports.ip_id = ips.id`,
		projectName,
		ip,
	).Scan(&ports)
	if res.Error != nil {
		return nil, errors.New("Could not get ports : " + res.Error.Error())
	}

	return ports, nil
}

//GetPorts is the public wrapper for getting all the ports for a project and a specific IP
func (pg *PostgreSQL) GetPorts(projectName string, ip string) ([]models.Port, error) {

	ports, err := pg.getPorts(projectName, ip)
	if err != nil {
		return nil, err
	}

	res := []models.Port{}
	for _, p := range ports {
		res = append(res, p.ToModel())
	}

	return res, nil
}

func (pg *PostgreSQL) getPort(projectName string, ip string, port string) (pgPort, error) {
	var p pgPort
	res := pg.db.Raw(`
	SELECT ports.*
		FROM ports, ips, projects
		WHERE projects.name = ?
		AND ips.value = ?
		AND ports.ip_id = ips.id
		AND ports.number = ?`,
		projectName,
		ip,
		port,
	).Scan(&p)

	if res.Error != nil {
		return p, errors.New("Could not get port : " + res.Error.Error())
	}

	return p, nil
}

//GetPort is the public wrapper for getting a specific port for a project based on the IP and port.
func (pg *PostgreSQL) GetPort(projectName string, ip string, port string) (models.Port, error) {
	pgp, err := pg.getPort(projectName, ip, port)
	if err != nil {
		return models.Port{}, nil
	}
	return pgp.ToModel(), nil
}

//DeletePort TOFIX
func (pg *PostgreSQL) DeletePort(projectName string, ip string, port models.Port) error {
	return errors.New("Not implemented yet")
}
