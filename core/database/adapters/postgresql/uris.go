package postgresql

import (
	"errors"

	"github.com/netm4ul/netm4ul/core/database/models"
)

// URI (directory and files)
func (pg *PostgreSQL) createOrUpdateURI(projectName string, ip string, port string, uri pgURI) error {
	//TOFIX : do real join / relation / whatever. Stop doing 4 request.
	proj := pgProject{}
	res := pg.db.Where("name = ?", projectName).First(&proj)
	if res.Error != nil {
		return errors.New("Could not match corresponding project for port : " + res.Error.Error())
	}

	pip := pgIP{}
	res = pg.db.Where("value = ?", ip).Where("project_id = ?", proj.ID).First(&pip)
	if res.Error != nil {
		return errors.New("Could not match corresponding ip for port : " + res.Error.Error())
	}

	pport := pgPort{}
	res = pg.db.Where("ip_id = ?", pip.ID).First(&pport)
	if res.Error != nil {
		return errors.New("Could not match corresponding ip for port : " + res.Error.Error())
	}

	uri.PortID = pport.ID
	res = pg.db.Where("port_id = ?", pport.ID).FirstOrCreate(&uri)
	if res.Error != nil {
		return res.Error
	}
	return nil
}

//CreateOrUpdateURI is the public wrapper to create or update a new URI for a project, ip and port.
func (pg *PostgreSQL) CreateOrUpdateURI(projectName string, ip string, port string, uri models.URI) error {

	puris := pgURI{}
	puris.FromModel(uri)

	err := pg.createOrUpdateURI(projectName, ip, port, puris)
	if err != nil {
		return err
	}

	return nil
}

//CreateURI is the public wrapper to create a new URI in the database.
func (pg *PostgreSQL) CreateURI(projectName string, ip string, port string, URI models.URI) error {
	return errors.New("Not implemented yet")
}

//UpdateURI is the public wrapper to update a new URI in the database.
func (pg *PostgreSQL) UpdateURI(projectName string, ip string, port string, URI models.URI) error {
	return errors.New("Not implemented yet")
}

func (pg *PostgreSQL) createOrUpdateURIs(projectName string, ip string, port string, uris []pgURI) error {
	// TOFIX
	// bulk insert!
	for _, uri := range uris {
		err := pg.createOrUpdateURI(projectName, ip, port, uri)
		if err != nil {
			return err
		}
	}
	return nil
}

//CreateOrUpdateURIs is the bulk insert version of CreateOrUpdateURI
func (pg *PostgreSQL) CreateOrUpdateURIs(projectName string, ip string, port string, uris []models.URI) error {
	puris := []pgURI{}
	for _, uri := range uris {
		puri := pgURI{}
		puri.FromModel(uri)
		puris = append(puris, puri)
	}

	err := pg.createOrUpdateURIs(projectName, ip, port, puris)
	if err != nil {
		return err
	}
	return nil
}

func (pg *PostgreSQL) getURIs(projectName string, ip string, port string) ([]pgURI, error) {

	uris := []pgURI{}

	res := pg.db.
		Where("ips.value = ?", ip).
		Where("ports.number = ?", port).
		Find(&uris)
	if res.Error != nil {
		return uris, errors.New("Could not get URIs : " + res.Error.Error())
	}

	return nil, nil
}

//GetURIs will return all the available URI in the database from a project ip and port combo. This function is a wrapper around `getURIs`
func (pg *PostgreSQL) GetURIs(projectName string, ip string, port string) ([]models.URI, error) {

	uris := []models.URI{}

	puris, err := pg.getURIs(projectName, ip, port)
	if err != nil {
		return nil, err
	}

	for _, puri := range puris {
		uris = append(uris, puri.ToModel())
	}
	return uris, nil
}

func (pg *PostgreSQL) getURI(projectName string, ip string, port string, dir string) (pgURI, error) {

	uri := pgURI{}

	res := pg.db.
		Where("ips.value = ?", ip).
		Where("ports.number = ?", port).
		First(&uri)

	if res.Error != nil {
		return pgURI{}, errors.New("Could not get URIs : " + res.Error.Error())
	}

	return uri, nil
}

//GetURI will return one URI in the database from a project ip and port combo. This function is a wrapper around `getURI`
func (pg *PostgreSQL) GetURI(projectName string, ip string, port string, dir string) (models.URI, error) {

	uri, err := pg.getURI(projectName, ip, port, dir)
	if err != nil {
		return models.URI{}, err
	}

	return uri.ToModel(), err
}

//DeleteURI TOFIX
func (pg *PostgreSQL) DeleteURI(projectName string, ip string, port string, dir models.URI) error {
	return errors.New("Not implemented yet")
}
