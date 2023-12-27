package main

import (
	"crypto/sha256"
	"database/sql"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/session"
	"math/rand"
	"slices"
	"strconv"
	"strings"
	"time"
)

type ctf struct {
	Storage       *sql.DB
	Challenges    map[string]challenge
	Sessions      *session.Store
	Configuration configuration
}

func initCTF(path string) (ctf, error) {
	var ctf ctf

	configuration, err := readConfiguration(path)
	if err != nil {
		return ctf, err
	}
	ctf.Configuration = configuration

	sessionStorage := initDB(fmt.Sprintf("%s/data.sqlite", path))

	err = dbInitStorage(sessionStorage.Conn())
	if err != nil {
		return ctf, err
	}
	ctf.Storage = sessionStorage.Conn()

	challenges, _ := readChallenges(path)
	ctf.Challenges = challenges

	sessions := session.New(session.Config{
		Storage:    sessionStorage,
		Expiration: time.Second * time.Duration(ctf.Configuration.SessionTimeout),
		KeyLookup:  "cookie:ctf_session",
	})
	ctf.Sessions = sessions

	return ctf, nil
}

func (ctf *ctf) logout(c *fiber.Ctx) error {
	sess, err := ctf.Sessions.Get(c)
	if err != nil {
		return err
	}
	err = sess.Destroy()
	if err != nil {
		return err
	}
	return nil
}

func (ctf *ctf) register(c *fiber.Ctx, username, password, password2, token string) (user, error) {
	if username == "" {
		return user{}, fmt.Errorf("empty username cannot be used for registration")
	}

	if password != password2 {
		return user{}, fmt.Errorf("passwords does not match")
	}

	if ctf.Configuration.RegistrationToken {
		state, err := dbHasSignupToken(ctf.Storage, token)
		if err != nil {
			return user{}, err
		}
		if !state {
			return user{}, fmt.Errorf("signup token required but not accepted")
		}
	}

	salt := strconv.Itoa(rand.New(rand.NewSource(time.Now().UnixNano())).Int())
	hash := sha256.Sum256([]byte(password + salt))

	id, err := dbUserRegister(ctf.Storage, username, string(hash[:]), salt)
	if ctf.Configuration.RegistrationToken {
		_ = dbDeleteSignupToken(ctf.Storage, token)
	}
	if err != nil {
		return user{}, err
	}

	sessionUser := user{id: id, db: ctf.Storage}

	if ctf.setSessionKey(c, "user", id) != nil {
		return user{}, err
	}

	return sessionUser, nil
}

type score struct {
	User   string
	Points int
}

func (ctf *ctf) categories() ([]string, error) {
	var categories []string

	for _, c := range ctf.Challenges {
		if !slices.Contains(categories, c.Category) {
			categories = append(categories, c.Category)
		}
	}
	slices.Sort(categories)
	return categories, nil
}

func (ctf *ctf) scores() ([]score, error) {
	var scores []score

	rows, err := dbGetScoreboard(ctf.Storage)
	if err != nil {
		return nil, err
	}

	for _, row := range rows {
		scores = append(scores, score{User: row[1].(string), Points: row[2].(int)})
	}

	return scores, nil
}

func (ctf *ctf) session(c *fiber.Ctx) (map[string]interface{}, error) {
	sess, err := ctf.Sessions.Get(c)
	if err != nil {
		return nil, err
	}

	var userName string
	var score int
	test := sess.Get("user")
	if test != nil {
		user := getUser(sess.Get("user").(int), ctf.Storage)
		userName, err = (&user).name()
		if err != nil {
			return nil, err
		}
		score, err = (&user).score()
		if err != nil {
			return nil, err
		}
	}

	sess.SetExpiry(time.Second * time.Duration(ctf.Configuration.SessionTimeout))
	if err := sess.Save(); err != nil {
		return nil, err
	}

	toasts, err := ctf.getToasts(c)
	if err != nil {
		toasts = []toast{}
	}

	return fiber.Map{
		"Session":      sess,
		"LoggedIn":     ctf.loggedIn(c),
		"UserName":     userName,
		"Score":        score,
		"session-user": sess.Get("user"),
		"Toasts":       toasts,
	}, nil
}

func (ctf *ctf) ensureLoggedIn(c *fiber.Ctx) (user, error) {
	sess, err := ctf.Sessions.Get(c)
	if err != nil {
		return user{}, err
	}
	if sess.Get("user") != nil {
		sessionUser := user{id: sess.Get("user").(int), db: ctf.Storage}
		_, err := sessionUser.name()
		if err == nil {
			return sessionUser, nil
		}
	}
	return user{}, fmt.Errorf("user not logged in")
}

func (ctf *ctf) loggedIn(c *fiber.Ctx) bool {
	_, err := ctf.ensureLoggedIn(c)
	if err == nil {
		return true
	}
	return false
}

func (ctf *ctf) addSignupToken(token string) error {
	return dbInsertSignupToken(ctf.Storage, token)
}

func (ctf *ctf) solve(c *fiber.Ctx, flag string) (int, error) {
	user, err := ctf.ensureLoggedIn(c)
	if err != nil {
		return 0, err
	}

	challenge := ctf.Challenges[c.Params("challengePath")]

	if strings.Compare(strings.TrimSpace(challenge.Flag), strings.TrimSpace(flag)) == 0 {
		hintCost, err := dbHintGetCost(ctf.Storage, user.id, c.Params("challengePath"))
		if err != nil {
			return 0, err
		}
		points := challenge.Points - hintCost
		err = dbChallengeAddSolve(ctf.Storage, user.id, c.Params("challengePath"), points)
		if err != nil {
			return 0, err
		}
		return points, nil
	}

	ctf.failedTriesAdd(c)
	return 0, fmt.Errorf("submitted flag was wrong")
}

func (ctf *ctf) solvedChallenges(c *fiber.Ctx) ([]string, error) {
	user, err := ctf.ensureLoggedIn(c)
	if err != nil {
		return []string{}, err
	}

	return dbChallengeGetSolved(ctf.Storage, user.id)
}

func (ctf *ctf) failedTriesAdd(c *fiber.Ctx) {
	sess, err := ctf.Sessions.Get(c)
	if err != nil {
		return
	}

	var tries []int64
	value := sess.Get("tries")
	if value != nil {
		tries = value.([]int64)
	}

	tries = append(tries, time.Now().UnixNano())
	_ = ctf.setSessionKey(c, "tries", tries)
}

func (ctf *ctf) coolDownActive(c *fiber.Ctx) bool {
	sess, err := ctf.Sessions.Get(c)
	if err != nil {
		return true
	}
	value := sess.Get("tries")
	if value == nil {
		return false
	}
	savedTries := value.([]int64)
	var tries []int64

	for _, attempt := range savedTries {
		attemptTime := time.Unix(0, attempt)
		passed := time.Now().Sub(attemptTime).Seconds()
		if passed < float64(ctf.Configuration.CoolDown) {
			tries = append(tries, attempt)
		}
	}

	tries = append(tries, time.Now().UnixNano())
	_ = ctf.setSessionKey(c, "tries", tries)

	if len(tries) > 3+1 {
		return true
	}

	return false
}

func (ctf *ctf) buyHint(c *fiber.Ctx, challengeID, hintID string) error {
	sess, err := ctf.Sessions.Get(c)
	if err != nil {
		return err
	}

	user := sess.Get("user")
	if user == nil {
		return fmt.Errorf("user not logged in")
	}

	challenge := ctf.Challenges[challengeID]

	for _, hint := range challenge.Hints {
		if hint.UID == hintID {
			err := dbInsertHint(ctf.Storage, sess.Get("user").(int), challengeID, hintID, hint.Cost)
			if err != nil {
				return err
			}
			return nil
		}
	}

	return fmt.Errorf("hint dows not exist")
}

func (ctf *ctf) getHints(c *fiber.Ctx, challengeID string) []string {
	var hintIDs []string

	sess, err := ctf.Sessions.Get(c)
	if err != nil {
		return hintIDs
	}

	user := sess.Get("user")
	if user == nil {
		return hintIDs
	}

	hintIDs, err = dbHintGetBoughtIDs(ctf.Storage, sess.Get("user").(int), challengeID)
	if err != nil {
		return hintIDs
	}

	return hintIDs
}

func (ctf *ctf) isSolved(c *fiber.Ctx, challengeID string) bool {
	sess, err := ctf.Sessions.Get(c)
	if err != nil {
		return false
	}

	user := sess.Get("user")
	if user == nil {
		return false
	}

	solved, err := ctf.solvedChallenges(c)
	if err != nil {
		return false
	}

	for _, c := range solved {
		if c == challengeID {
			return true
		}
	}
	return false
}
