package main

import (
	"crypto/sha256"
	"database/sql"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"time"
)

type user struct {
	id int
	db *sql.DB
}

func getUser(id int, db *sql.DB) user {
	return user{id: id, db: db}
}

func (u *user) name() (string, error) {
	//goland:noinspection GoDirectComparisonOfErrors
	switch name, err := dbUserGetName(u.db, u.id); err {
	case sql.ErrNoRows:
		return "", fmt.Errorf("no row in table users with id %d", u.id)
	case nil:
		return name, nil
	default:
		return "", err
	}
}

func (u *user) score() (int, error) {
	//goland:noinspection GoDirectComparisonOfErrors
	switch score, err := dbUserGetScore(u.db, u.id); err {
	case sql.ErrNoRows:
		return 0, fmt.Errorf("no row in table users with id %d", u.id)
	case nil:
		return score, nil
	default:
		return 0, err
	}
}

func (ctf *ctf) setSessionKey(c *fiber.Ctx, id string, key interface{}) error {
	sess, err := ctf.Sessions.Get(c)
	if err != nil {
		return err
	}
	sess.Set(id, key)
	sess.SetExpiry(time.Second * time.Duration(ctf.Configuration.SessionTimeout))
	if err := sess.Save(); err != nil {
		return handleError(c, err)
	}
	return nil
}

func (ctf *ctf) login(c *fiber.Ctx, username, password string) (user, error) {
	var id int
	var salt string
	var hash string
	row := ctf.Storage.QueryRow(`SELECT id, password, salt FROM users WHERE name=$1;`, username)
	//goland:noinspection GoDirectComparisonOfErrors
	switch err := row.Scan(&id, &hash, &salt); err {
	case sql.ErrNoRows:
		return user{}, fmt.Errorf("cannot Login: user does not exist")
	case nil:
		hashCalculated := sha256.Sum256([]byte(password + salt))
		if hash[:] == string(hashCalculated[:]) {
			err = ctf.setSessionKey(c, "user", id)
			if err != nil {
				return user{}, err
			}
			return user{id: id, db: ctf.Storage}, nil
		}
		return user{}, fmt.Errorf("cannot Login: password is wrong")
	default:
		return user{}, err
	}
}
