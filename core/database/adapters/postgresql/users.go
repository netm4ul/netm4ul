package postgresql

import (
	"errors"

	"github.com/jinzhu/gorm"
	"github.com/netm4ul/netm4ul/core/database/models"
	"github.com/netm4ul/netm4ul/core/security"
	log "github.com/sirupsen/logrus"
)

//TODO : refactor this.

//CreateOrUpdateUser will create or update a new user in the database.
//If the user already exist,
func (pg *PostgreSQL) CreateOrUpdateUser(user models.User) error {

	inDbUser, err := pg.getUser(user.Name)
	if err != nil {
		return errors.New("Could not get user : " + err.Error())
	}

	// transform user into pgUser model !
	pguser := pgUser{}
	pguser.FromModel(user)
	log.Debugf("pguser : %+v", &pguser)
	log.Debugf("user : %+v", &user)

	//user doesn't exist, create it and exit.
	if inDbUser.Name == "" {
		res := pg.db.Where("name = ?", inDbUser.Name).FirstOrCreate(&pguser)
		if res.Error != nil {
			return errors.New("Could not insert user in the database : " + res.Error.Error())
		}
		return nil
	}

	//if the in-database user doesn't have a token, create one : it might be a security issues without one.
	if inDbUser.Token == "" {
		inDbUser.Token = security.GenerateNewToken()
	}

	// The user exist, check if the password is correct before updating anything (except the token just above)
	if !security.ComparePassword(inDbUser.Password, pguser.Password) {
		log.Debug("Updating password for user : ", pguser.Name)
		return errors.New("Could not update user : the provided password doesn't match the stored one")
	}

	if pguser.Token != "" && inDbUser.Token != pguser.Token {
		log.Debug("Updating token for user : ", pguser.Name)
		inDbUser.Token = pguser.Token
	}

	log.Debugf("Writing tmp user : %+v", inDbUser)
	res := pg.db.Model(&inDbUser).Update(&inDbUser)

	if res.Error != nil {
		return errors.New("Could not update user : " + res.Error.Error())
	}
	return nil
}

func (pg *PostgreSQL) getUser(username string) (pgUser, error) {

	pguser := pgUser{}
	res := pg.db.Where("name = ?", username).First(&pguser)

	// Accept empty rows !
	if res.Error != nil && !gorm.IsRecordNotFoundError(res.Error) {
		return pgUser{}, errors.New("Could not get user by name : " + res.Error.Error())
	}
	return pguser, nil
}

func (pg *PostgreSQL) GetUser(username string) (models.User, error) {
	pguser, err := pg.getUser(username)
	if err != nil {
		return models.User{}, err
	}
	return pguser.ToModel(), err
}

func (pg *PostgreSQL) getUserByToken(token string) (pgUser, error) {

	pguser := pgUser{}
	res := pg.db.Where("token = ?", token).First(&pguser)
	// Accept empty rows !
	if res.Error != nil && !gorm.IsRecordNotFoundError(res.Error) {
		return pgUser{}, errors.New("Could not get user by token : " + res.Error.Error())
	}
	return pguser, nil
}

func (pg *PostgreSQL) GetUserByToken(token string) (models.User, error) {
	pguser, err := pg.getUserByToken(token)
	if err != nil {
		return models.User{}, err
	}
	return pguser.ToModel(), err
}

/*
GenerateNewToken generates a new token and save it in the database.
It uses the function GenerateNewToken provided by the `models` class
*/
func (pg *PostgreSQL) GenerateNewToken(user models.User) error {

	user.Token = security.GenerateNewToken()
	err := pg.CreateOrUpdateUser(user)
	if err != nil {
		return errors.New("Could not generate a new token : " + err.Error())
	}
	return nil
}

//DeleteUser remove the user from the database (using its ID)
func (pg *PostgreSQL) DeleteUser(user models.User) error {
	res := pg.db.Delete(user)
	if res.Error != nil {
		return errors.New("Could not delete user from the database : " + res.Error.Error())
	}
	return nil
}
