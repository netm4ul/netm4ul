package postgresql

import (
	"errors"

	"github.com/netm4ul/netm4ul/core/database/models"
)

// Raw data
func (pg *PostgreSQL) appendRawData(projectName string, raw pgRaw) error {
	res := pg.db.Create(&raw)
	if res.Error != nil {
		return errors.New("Could not insert raw : " + res.Error.Error())
	}
	return nil
}

//AppendRawData is the public wrapper to insert raw data into the database.
func (pg *PostgreSQL) AppendRawData(projectName string, raw models.Raw) error {
	praw := pgRaw{}
	praw.FromModel(raw)
	err := pg.appendRawData(projectName, praw)
	if err != nil {
		return err
	}
	return nil
}

func (pg *PostgreSQL) getRaws(projectName string) ([]pgRaw, error) {

	raws := []pgRaw{}

	res := pg.db.Find(raws)
	if res.Error != nil {
		return nil, errors.New("Could not get raws : " + res.Error.Error())
	}

	return raws, nil
}
//GetRaws is the public wrapper to get all the raw data for a project.
func (pg *PostgreSQL) GetRaws(projectName string) ([]models.Raw, error) {

	raws := []models.Raw{}

	praws, err := pg.getRaws(projectName)
	if err != nil {
		return nil, err
	}

	for _, praw := range praws {
		raws = append(raws, praw.ToModel())
	}
	return raws, nil
}

func (pg *PostgreSQL) getRawModule(projectName string, moduleName string) (map[string][]pgRaw, error) {
	raws := []pgRaw{}
	res := pg.db.
		Where("raws.name = ?", projectName).
		Where("raws.moduleName = ?", moduleName).
		Find(&raws)

	if res.Error != nil {
		return nil, errors.New("Could not get raw by module : " + res.Error.Error())
	}
	var mapOfListOfRaw map[string][]pgRaw
	mapOfListOfRaw = make(map[string][]pgRaw)

	for _, r := range raws {
		mapOfListOfRaw[r.ModuleName] = append(mapOfListOfRaw[r.ModuleName], r)
	}
	return mapOfListOfRaw, nil
}

//GetRawModule is the public wrapper to get all the raw data for a specific module name.
func (pg *PostgreSQL) GetRawModule(projectName string, moduleName string) (map[string][]models.Raw, error) {
	var mapOfListOfRaw map[string][]models.Raw
	mapOfListOfRaw = make(map[string][]models.Raw)

	rawsmap, err := pg.getRawModule(projectName, moduleName)

	if err != nil {
		return nil, err
	}

	//translate list of pgRaw to list of models.Raw
	for i, praws := range rawsmap {
		raws := []models.Raw{}
		for _, praw := range praws {
			raws = append(raws, praw.ToModel())
		}
		mapOfListOfRaw[i] = raws
	}

	return mapOfListOfRaw, nil
}
