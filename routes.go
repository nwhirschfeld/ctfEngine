package main

import (
	"fmt"
	"github.com/gofiber/fiber/v2"
)

func renderWithSession(c *fiber.Ctx, ctf ctf, name string, additionalEnv ...fiber.Map) error {
	sess, err := ctf.session(c)
	if err != nil {
		return handleError(c, err)
	}

	env := fiber.Map{
		"CTF":     ctf,
		"Path":    c.Path(),
		"Session": sess,
	}

	for _, i := range additionalEnv {
		for k, v := range i {
			env[k] = v
		}
	}

	return c.Render(fmt.Sprintf("views/%s", name), env, "views/layouts/main")
}

func rGetChallenges(c *fiber.Ctx, ctf *ctf) error {
	if !ctf.loggedIn(c) {
		ctf.addToast(c, "Login needed",
			"You need to log in to view the challenges.")
		return c.Redirect("/")
	}

	solvedChallenges, err := ctf.solvedChallenges(c)
	if err != nil {
		solvedChallenges = []string{}
	}

	categories, err := ctf.categories()
	if err != nil {
		return handleError(c, err)
	}

	return renderWithSession(c, *ctf, "challenges", fiber.Map{
		"SolvedChallenges": solvedChallenges,
		"Categories":       categories,
	})
}

func rGetChallenge(c *fiber.Ctx, ctf *ctf) error {
	if !ctf.loggedIn(c) {
		ctf.addToast(c, "Login needed",
			"You need to log in to view the challenges.")
		return c.Redirect("/")
	}

	return renderWithSession(c, *ctf, "challenge", fiber.Map{
		"Challenge": ctf.Challenges[c.Params("challengePath")],
		"Hints":     ctf.getHints(c, c.Params("challengePath")),
		"Solved":    ctf.isSolved(c, c.Params("challengePath")),
	})
}

func rPostChallenge(c *fiber.Ctx, ctf *ctf) error {
	if !ctf.loggedIn(c) {
		ctf.addToast(c, "Login needed",
			"You need to log in to solve the challenges.")
		return c.Redirect("/")
	}
	payload := struct {
		Flag string `form:"flag"`
	}{}

	if err := c.BodyParser(&payload); err != nil {
		return err
	}

	challenge := ctf.Challenges[c.Params("challengePath")]

	if ctf.coolDownActive(c) {
		ctf.addToast(c, "Committed too many false flags",
			"You committed too many false flags. Try again in a few seconds.")
		return renderWithSession(c, *ctf, "challenge", fiber.Map{
			"Challenge": challenge,
		})
	}

	points, err := ctf.solve(c, payload.Flag)
	if err != nil {
		ctf.addToast(c, "Flag incorrect",
			"The flag was not correct. Try again!")
	} else {
		ctf.addToast(c,
			fmt.Sprintf("Challenge \"%s\" solved", challenge.Title),
			fmt.Sprintf("You solved challenge \"%s\" and earned %d points.", challenge.Title, points))

	}

	return c.Redirect(fmt.Sprintf("/challenges/%s", c.Params("challengePath")))
}

func rGetChallengeFile(c *fiber.Ctx, ctf *ctf) error {
	if !ctf.loggedIn(c) {
		ctf.addToast(c, "Login needed",
			"You need to log in to download files.")
		return c.Redirect("/")
	}
	challenge := ctf.Challenges[c.Params("challengePath")]
	file := challenge.Files[c.Params("fileID")]
	return c.Download(file.Location)
}

func rPostChallengeHint(c *fiber.Ctx, ctf *ctf) error {
	if !ctf.loggedIn(c) {
		ctf.addToast(c, "Login needed",
			"You need to log in to get hints.")
		return c.Redirect("/")
	}
	payload := struct {
		HintID string `form:"hintid"`
	}{}

	if err := c.BodyParser(&payload); err != nil {
		return err
	}

	err := ctf.buyHint(c, c.Params("challengePath"), payload.HintID)
	if err != nil {
		return err
	}

	return c.Redirect(fmt.Sprintf("/challenges/%s", c.Params("challengePath")))
}

func rGetDeleteToast(c *fiber.Ctx, ctf *ctf) error {
	ctf.deleteToast(c, c.Params("toastId"))
	return c.Send([]byte{})
}

func rGetScore(c *fiber.Ctx, ctf *ctf) error {
	scores, err := ctf.scores()
	if err != nil {
		return handleError(c, err)
	}
	return renderWithSession(c, *ctf, "score", fiber.Map{
		"Scores": scores,
	})
}

func rPostLogin(c *fiber.Ctx, ctf *ctf) error {
	payload := struct {
		Username string `form:"username"`
		Password string `form:"password"`
	}{}

	if err := c.BodyParser(&payload); err != nil {
		return err
	}
	_, err := ctf.login(c, payload.Username, payload.Password)
	if err != nil {
		ctf.addToast(c, "Login failed",
			"Something went wrong, please try again.")
		return handleError(c, err)
	}
	return c.Redirect("/")
}

func rPostLogout(c *fiber.Ctx, ctf *ctf) error {
	err := ctf.logout(c)
	if err != nil {
		return handleError(c, err)
	}
	return c.Redirect("/")
}

func rPostSignup(c *fiber.Ctx, ctf *ctf) error {
	payload := struct {
		Username  string `form:"username"`
		Password  string `form:"password"`
		Password2 string `form:"password2"`
	}{}

	if err := c.BodyParser(&payload); err != nil {
		return err
	}

	_, err := ctf.register(c, payload.Username, payload.Password, payload.Password2)
	switch err {
	case nil:
		ctf.addToast(c, "Registration completed",
			fmt.Sprintf("You have been registered sucessfully as \"%s\"!", payload.Username))
	default:
		ctf.addToast(c, "Registration failed",
			"Something went wrong, please try again.")
	}
	return c.Redirect("/")
}

func addRoutes(app *fiber.App, ctf *ctf) {
	app.Get("/", func(c *fiber.Ctx) error {

		return renderWithSession(c, *ctf, "index")
	})

	app.Get("/challenges", func(c *fiber.Ctx) error {
		return rGetChallenges(c, ctf)
	})

	app.Get("/challenges/:challengePath", func(c *fiber.Ctx) error {
		return rGetChallenge(c, ctf)
	})

	// try to solve a challenge
	app.Post("/challenges/:challengePath", func(c *fiber.Ctx) error {
		return rPostChallenge(c, ctf)
	})

	// get challenge file
	app.Get("/challenges/:challengePath/files/:fileID", func(c *fiber.Ctx) error {
		return rGetChallengeFile(c, ctf)
	})

	// "buy" challenge hint
	app.Post("/challenges/:challengePath/hint", func(c *fiber.Ctx) error {
		return rPostChallengeHint(c, ctf)
	})

	// delete toast
	app.Get("/toast/:toastId", func(c *fiber.Ctx) error {
		return rGetDeleteToast(c, ctf)
	})

	// get scoreboard
	app.Get("/score", func(c *fiber.Ctx) error {
		return rGetScore(c, ctf)
	})

	// login user
	app.Post("/login", func(c *fiber.Ctx) error {
		return rPostLogin(c, ctf)
	})

	// logout user
	app.Get("/logout", func(c *fiber.Ctx) error {
		return rPostLogout(c, ctf)
	})

	// register new user
	app.Post("/signup", func(c *fiber.Ctx) error {
		return rPostSignup(c, ctf)
	})
}
