package postgresql

import (
	"errors"
	"github.com/netm4ul/netm4ul/core/events"

	"github.com/jinzhu/gorm"
	"github.com/netm4ul/netm4ul/core/database/models"
	log "github.com/sirupsen/logrus"
)

func (pg *PostgreSQL) createOrUpdateProject(project pgProject) error {
	var p models.Project

	res := pg.db.Raw(selectProjectByName, project.Name).Scan(&p)
	// The project doesn't exist yet
	if gorm.IsRecordNotFoundError(res.Error) {
		res := pg.db.Create(&project)
		if res.Error != nil {
			return errors.New("Could not insert project : " + res.Error.Error())
		}
		return nil
	}

	// handle other errors
	if res.Error != nil {
		return errors.New("Could not select project : " + res.Error.Error())
	}

	// update if the project was found
	res = pg.db.Model(&project).Where("name = ?", project.Name).Update(project)
	if res.Error != nil {
		return errors.New("Could not save project in the database :" + res.Error.Error())
	}

	return nil
}

//CreateOrUpdateProject is the public wrapper to create or update a new project in the database.
func (pg *PostgreSQL) CreateOrUpdateProject(project models.Project) error {
	log.Debugf("CreateOrUpdateProject : %s", project)

	p := pgProject{}
	p.FromModel(project)

	err := pg.createOrUpdateProject(p)
	if err != nil {
		return err
	}
	return nil
}

//CreateProject is the public wrapper to create a new Project in the database.
func (pg *PostgreSQL) CreateProject(project models.Project) error {
	res := pg.db.Exec(insertProject, project.Name, project.Description)
	if res.Error != nil {
		return res.Error
	}
	events.NewEventProject(project)
	return nil
}

//UpdateProject is the public wrapper to update a new Project in the database.
func (pg *PostgreSQL) UpdateProject(project models.Project) error {
	res := pg.db.Model(&project).Update(project)
	if res.Error != nil {
		return errors.New("Could not update project : " + res.Error.Error())
	}
	return nil
}

func (pg *PostgreSQL) getProjects() ([]pgProject, error) {
	var projects []pgProject
	res := pg.db.Find(&projects)
	if res.Error != nil {
		return nil, errors.New("Could not select projects : " + res.Error.Error())
	}

	return projects, nil
}

//GetProjects is the public wrapper for getting all the project available
func (pg *PostgreSQL) GetProjects() ([]models.Project, error) {
	var projects []models.Project

	ps, err := pg.getProjects()
	if err != nil {
		return nil, err
	}

	// convert to model
	for _, p := range ps {
		projects = append(projects, p.ToModel())
	}
	return projects, nil
}

func (pg *PostgreSQL) getProject(projectName string) (pgProject, error) {
	var project pgProject

	res := pg.db.Where("name = ?", projectName).Find(&project)
	if res.Error != nil {
		return project, errors.New("Could not select project : " + res.Error.Error())
	}

	return project, nil
}

//GetProject is the public wrapper for getting all the informations on a specific project
func (pg *PostgreSQL) GetProject(projectName string) (models.Project, error) {
	p, err := pg.getProject(projectName)
	if err != nil {
		return models.Project{}, err
	}

	return p.ToModel(), nil
}

//DeleteProject TOFIX
func (pg *PostgreSQL) DeleteProject(project models.Project) error {
	return errors.New("Not implemented yet")
}
