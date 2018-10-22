package postgresql

import (
	"errors"

	"github.com/jinzhu/gorm"
	"github.com/netm4ul/netm4ul/core/database/models"
	log "github.com/sirupsen/logrus"
)

func (pg *PostgreSQL) createOrUpdateIP(projectName string, ip pgIP) error {
	log.Debugf("Inserting ip : %+v", ip)

	proj, err := pg.getProject(projectName)
	if err != nil {
		return errors.New("Could not find corresponding project for ip :" + err.Error())
	}

	ip.ProjectID = proj.ID

	res := pg.db.
		Where("project_id = ?", proj.ID).
		Where("value = ?", ip.Value).
		Where("network = ?", ip.Network).
		FirstOrCreate(&ip)

	if res.Error != nil {
		return errors.New("Could not save ip in the database :" + res.Error.Error())
	}
	return nil
}

//CreateOrUpdateIP is the public wrapper to create or update a new IP in the database.
func (pg *PostgreSQL) CreateOrUpdateIP(projectName string, ip models.IP) error {

	// convert to pgIP first
	pip := pgIP{}
	pip.FromModel(ip)

	err := pg.createOrUpdateIP(projectName, pip)
	if err != nil {
		return err
	}

	return nil
}

func (pg *PostgreSQL) createOrUpdateIPs(projectName string, ips []pgIP) error {
	for _, ip := range ips {
		err := pg.createOrUpdateIP(projectName, ip)
		if err != nil {
			return errors.New("Could not create or update ips : " + err.Error())
		}
	}
	return nil
}

//CreateOrUpdateIPs is the public wrapper to create or update multiple IP
// This function should be used instead of CreateOrUpdateIP during bulk inserts.
func (pg *PostgreSQL) CreateOrUpdateIPs(projectName string, ips []models.IP) error {

	pips := []pgIP{}
	for _, ip := range ips {
		// convert ip
		pip := pgIP{}
		pip.FromModel(ip)
	}

	err := pg.createOrUpdateIPs(projectName, pips)
	if err != nil {
		return err
	}
	return nil
}

func (pg *PostgreSQL) getIPs(projectName string) ([]pgIP, error) {

	pgips := []pgIP{}
	res := pg.db.Raw(`
	SELECT ips.* FROM "ips","projects"
		WHERE "ips"."deleted_at" IS NULL
		AND ips.project_id = projects.id
		AND ((projects.name = ?))
	`, projectName).Scan(&pgips)

	if res.Error != nil {
		return nil, errors.New("Could not get IPs : " + res.Error.Error())
	}

	return pgips, nil
}

//GetIPs is the public wrapper for getting all the IP for a project
func (pg *PostgreSQL) GetIPs(projectName string) ([]models.IP, error) {

	pgips, err := pg.getIPs(projectName)
	if err != nil {
		return nil, err
	}

	// convert back to the standard model
	ips := []models.IP{}
	for _, ip := range pgips {
		ips = append(ips, ip.ToModel())
	}

	return ips, nil
}

func (pg *PostgreSQL) getIP(projectName string, ip string) (pgIP, error) {

	pgip := pgIP{}
	res := pg.db.
		Where("projects.name = ?", projectName).
		Where("ips.value = ?", ip).
		Find(&pgip)

	if gorm.IsRecordNotFoundError(res.Error) {
		return pgIP{}, nil
	}
	if res.Error != nil {
		return pgIP{}, errors.New("Could not get IP : " + res.Error.Error())
	}

	return pgip, nil
}

//GetIP is the public wrapper for getting a specific IP for a project.
func (pg *PostgreSQL) GetIP(projectName string, ip string) (models.IP, error) {

	pgip, err := pg.getIP(projectName, ip)
	if err != nil {
		return models.IP{}, err
	}
	return pgip.ToModel(), nil
}

//DeleteIP TOFIX
func (pg *PostgreSQL) DeleteIP(ip models.IP) error {
	return errors.New("Not implemented yet")
}
