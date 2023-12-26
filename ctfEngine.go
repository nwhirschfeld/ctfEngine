package main

import (
	"embed"
	"flag"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/filesystem"
	"github.com/gofiber/template/html/v2"
	"github.com/russross/blackfriday"
	"html/template"
	"log"
	"math"
	"net/http"
	"strconv"
	"strings"
)

func handleError(c *fiber.Ctx, err error) error {
	fmt.Printf("[ERROR] an error occurred: %s!\n", err)
	return c.Redirect("/")
}

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

//go:embed views/*
var viewsFS embed.FS

//go:embed static/*
var staticFS embed.FS

func main() {
	var ctfLocation string
	flag.StringVar(&ctfLocation, "l", "/ctf", "load ctf from this directory")
	flag.Parse()
	ctf, err := initCTF(ctfLocation)
	if err != nil {
		fmt.Printf("Could not load CTF from path \"%s\".\n", ctfLocation)
		return
	}

	engine := html.NewFileSystem(http.FS(viewsFS), ".html")

	engine.AddFunc(
		"renderMarkdown", func(s string) template.HTML {
			return template.HTML(blackfriday.MarkdownCommon([]byte(s)))
		},
	)
	engine.AddFunc(
		"ppFilesize", func(size int64) template.HTML {
			var suffixes [5]string
			suffixes[0] = "B"
			suffixes[1] = "KB"
			suffixes[2] = "MB"
			suffixes[3] = "GB"
			suffixes[4] = "TB"

			base := math.Log(float64(size)) / math.Log(1024)
			getSize := math.Ceil(10*math.Pow(1024, base-math.Floor(base))) / 10
			getSuffix := suffixes[int(math.Floor(base))]

			return template.HTML(strconv.FormatFloat(getSize, 'f', -1, 64) + " " + getSuffix)
		},
	)
	engine.AddFunc(
		"inList", func(element string, list []string) bool {
			for _, listElement := range list {
				if strings.Compare(element, listElement) == 0 {
					return true
				}
			}
			return false
		},
	)

	app := fiber.New(fiber.Config{
		Views:        engine,
		ServerHeader: "ctfEngine",
		AppName:      "ctfEngine",
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			// TODO: add logging here
			return c.Render("views/error", fiber.Map{}, "views/layouts/main")

		},
	})

	app.Use("/static", filesystem.New(filesystem.Config{
		Root:       http.FS(staticFS),
		PathPrefix: "static",
		Browse:     true,
	}))

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

	log.Fatal(app.Listen(":3000"))
}
