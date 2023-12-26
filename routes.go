package main

import (
	"fmt"
	"github.com/gofiber/fiber/v2"
)

func addRoutes(app *fiber.App, ctf ctf) {
	app.Get("/", func(c *fiber.Ctx) error {
		return renderWithSession(c, ctf, "index")
	})

	app.Get("/challenges", func(c *fiber.Ctx) error {
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

		return renderWithSession(c, ctf, "challenges", fiber.Map{
			"SolvedChallenges": solvedChallenges,
			"Categories":       categories,
		})
	})

	app.Get("/challenges/:challengePath", func(c *fiber.Ctx) error {
		if !ctf.loggedIn(c) {
			ctf.addToast(c, "Login needed",
				"You need to log in to view the challenges.")
			return c.Redirect("/")
		}

		return renderWithSession(c, ctf, "challenge", fiber.Map{
			"Challenge": ctf.Challenges[c.Params("challengePath")],
			"Hints":     ctf.getHints(c, c.Params("challengePath")),
			"Solved":    ctf.isSolved(c, c.Params("challengePath")),
		})
	})

	// try to solve a challenge
	app.Post("/challenges/:challengePath", func(c *fiber.Ctx) error {
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
			return renderWithSession(c, ctf, "challenge", fiber.Map{
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
	})

	// get challenge file
	app.Get("/challenges/:challengePath/files/:fileID", func(c *fiber.Ctx) error {
		if !ctf.loggedIn(c) {
			ctf.addToast(c, "Login needed",
				"You need to log in to download files.")
			return c.Redirect("/")
		}
		challenge := ctf.Challenges[c.Params("challengePath")]
		file := challenge.Files[c.Params("fileID")]
		return c.Download(file.Location)
	})

	// "buy" challenge hint
	app.Post("/challenges/:challengePath/hint", func(c *fiber.Ctx) error {
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
	})

	// delete toast
	app.Get("/toast/:toastId", func(c *fiber.Ctx) error {
		ctf.deleteToast(c, c.Params("toastId"))
		return c.Send([]byte{})
	})

	// get scoreboard
	app.Get("/score", func(c *fiber.Ctx) error {
		scores, err := ctf.scores()
		if err != nil {
			return handleError(c, err)
		}
		return renderWithSession(c, ctf, "score", fiber.Map{
			"Scores": scores,
		})
	})

	// login user
	app.Post("/login", func(c *fiber.Ctx) error {
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
	})

	// logout user
	app.Get("/logout", func(c *fiber.Ctx) error {
		err := ctf.logout(c)
		if err != nil {
			return handleError(c, err)
		}
		return c.Redirect("/")
	})

	// register new user
	app.Post("/signup", func(c *fiber.Ctx) error {
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
	})
}
