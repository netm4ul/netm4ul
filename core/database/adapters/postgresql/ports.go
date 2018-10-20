package postgresql

import (
	"errors"

	"github.com/netm4ul/netm4ul/core/database/models"
)

// TOFIX : doing an actual join/relation insert instead of 3 f requests
func (pg *PostgreSQL) createOrUpdatePort(projectName string, ip string, port pgPort) error {
	proj := pgProject{}
	res := pg.db.Where("name = ?", projectName).First(&proj)
	if res.Error != nil {
		return errors.New("Could not corresponding project for port : " + res.Error.Error())
	}

	pip := pgIP{}
	res = pg.db.Where("value = ?", ip).Where("project_id = ?", proj.ID).First(&pip)
	if res.Error != nil {
		return errors.New("Could not corresponding ip for port : " + res.Error.Error())
	}

	port.IPId = pip.ID
	res = pg.db.
		Where("ip_id = ?", pip.ID).
		Where("number = ?", port.Number).
		Where("protocol = ?", port.Protocol).
		FirstOrCreate(&port)
	if res.Error != nil {
		return errors.New("Could not create or update port : " + res.Error.Error())
	}

	return nil
}

func (pg *PostgreSQL) CreateOrUpdatePort(projectName string, ip string, port models.Port) error {

	pgp := pgPort{}
	pgp.FromModel(port)
	err := pg.createOrUpdatePort(projectName, ip, pgp)
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

func (pg *PostgreSQL) GetPort(projectName string, ip string, port string) (models.Port, error) {
	pgp, err := pg.getPort(projectName, ip, port)
	if err != nil {
		return models.Port{}, nil
	}
	return pgp.ToModel(), nil
}

func (pg *PostgreSQL) DeletePort(projectName string, ip string, port models.Port) error {
	return errors.New("Not implemented yet")
}
