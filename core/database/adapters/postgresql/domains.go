package postgresql

import (
	"errors"

	"github.com/netm4ul/netm4ul/core/database/models"
	log "github.com/sirupsen/logrus"
)

func (pg *PostgreSQL) createOrUpdateDomain(projectName string, domain pgDomain) error {
	project := pgProject{}
	res := pg.db.Where("name = ?", projectName).First(&project)
	if res.Error != nil {
		return errors.New("Could not find assiociated projet : " + res.Error.Error())
	}
	log.Debugf("Project with name : %s : %+v", projectName, project)

	domain.Project = project
	log.Debugf("Saving domain : %+v", domain)
	res = pg.db.Where("project_id = ?", project.ID).Where("name = ?", domain.Name).FirstOrCreate(&domain)
	if res.Error != nil {
		return errors.New("Could not save or update domain : " + res.Error.Error())
	}

	return nil
}

func (pg *PostgreSQL) CreateOrUpdateDomain(projectName string, domain models.Domain) error {

	d := pgDomain{}
	d.FromModel(domain)

	err := pg.createOrUpdateDomain(projectName, d)
	if err != nil {
		return err
	}

	return nil
}

func (pg *PostgreSQL) createOrUpdateDomains(projectName string, domains []pgDomain) error {
	for _, domain := range domains {
		err := pg.createOrUpdateDomain(projectName, domain)
		if err != nil {
			return errors.New("Could not save or update domains : " + err.Error())
		}
	}
	return nil
}

func (pg *PostgreSQL) CreateOrUpdateDomains(projectName string, domains []models.Domain) error {

	pgds := []pgDomain{}
	for _, domain := range domains {
		pgd := pgDomain{}
		pgd.FromModel(domain)
		pgds = append(pgds, pgd)
	}

	err := pg.createOrUpdateDomains(projectName, pgds)
	if err != nil {
		return err
	}

	return nil
}

func (pg *PostgreSQL) getDomains(projectName string) ([]pgDomain, error) {
	domains := []pgDomain{}
	res := pg.db.Raw(`
	SELECT *
		FROM domains, projects
		WHERE projects.name = ?
		AND projects.id = domains.project_id
		`,
		projectName,
	).Scan(&domains)

	if res.Error != nil {
		return nil, errors.New("Could not get domains : " + res.Error.Error())
	}

	return domains, nil
}

func (pg *PostgreSQL) GetDomains(projectName string) ([]models.Domain, error) {

	domains := []models.Domain{}
	pgds, err := pg.getDomains(projectName)
	if err != nil {
		return nil, err
	}

	for _, d := range pgds {
		domains = append(domains, d.ToModel())
	}

	return domains, nil
}

func (pg *PostgreSQL) getDomain(projectName string, domainName string) (pgDomain, error) {
	domain := pgDomain{}

	res := pg.db.
		Where("projects.name = ?", projectName).
		Where("domains.name = ?", domainName).
		Find(&domain)
	if res.Error != nil {
		return pgDomain{}, errors.New("Could not get domain : " + res.Error.Error())
	}

	return domain, nil
}

func (pg *PostgreSQL) GetDomain(projectName string, domainName string) (models.Domain, error) {
	d, err := pg.getDomain(projectName, domainName)
	if err != nil {
		return models.Domain{}, err
	}

	return d.ToModel(), nil
}

func (pg *PostgreSQL) DeleteDomain(projectName string, domain models.Domain) error {
	return errors.New("Not implemented yet")
}
