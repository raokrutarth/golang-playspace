package main

import (
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	"github.com/raokrutarth/golang-playspace/templates"

	// https://betterstack.com/community/guides/logging/logging-in-go/
	"golang.org/x/exp/slog"

	"crypto/rand"
	"crypto/rsa"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/compress"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/csrf"
	"github.com/gofiber/fiber/v2/middleware/favicon"
	"github.com/gofiber/fiber/v2/middleware/filesystem"
	"github.com/gofiber/fiber/v2/middleware/idempotency"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/fiber/v2/middleware/requestid"
	"github.com/gofiber/fiber/v2/utils"

	jwtware "github.com/gofiber/jwt/v3"
	"github.com/golang-jwt/jwt/v4"
)

var (
	// Obviously, this is just a test example. Do not do this in production.
	// In production, you would have the private key and public key pair generated
	// in advance. NEVER add a private key to any GitHub repo.
	privateKey *rsa.PrivateKey
)

func main() {
	err := godotenv.Load("dev.env")
	logger := slog.New(slog.NewJSONHandler(os.Stdout))

	slog.SetDefault(logger)

	if err != nil {
		log.Fatal(err)
	}

	app := fiber.New()

	// Define a template struct to hold the data for rendering the template
	type Template struct {
		Title   string
		Heading string
		Content string
	}

	// Define a route to render the index template
	app.Get("/", func(c *fiber.Ctx) error {
		// Define the data for the template
		data := Template{
			Title:   "My Title",
			Heading: "My Heading",
			Content: "My Content",
		}

		// Render the template with the data
		return templates.Resources.ExecuteTemplate(c.Response().BodyWriter(), "index.html", data)
	})

	// Just as a demo, generate a new private/public key pair on each run. See note above.
	rng := rand.Reader
	privateKey, err = rsa.GenerateKey(rng, 2048)
	if err != nil {
		log.Fatalf("rsa.GenerateKey: %v", err)
	}

	app.Use(requestid.New())
	// app.Use(logger.New(logger.Config{
	// 	// For more options, see the Config section
	// 	Format: "${pid} ${locals:requestid} ${status} - ${method} ${path}​\n",
	// }))

	// Or extend your config for customization
	app.Use(filesystem.New(filesystem.Config{
		Root:         http.Dir("./assets"),
		Browse:       true,
		Index:        "index.html",
		NotFoundFile: "404.html",
		MaxAge:       3600,
	}))

	// Default middleware config
	app.Use(recover.New())
	// Provide a minimal config
	app.Use(favicon.New())

	// Or extend your config for customization
	app.Use(favicon.New(favicon.Config{
		File: "./favicon.ico",
		URL:  "/favicon.ico",
	}))

	app.Use(idempotency.New(idempotency.Config{
		Lifetime: 42 * time.Minute,
		// ...
	}))

	// Default middleware config
	app.Use(compress.New())

	// Provide a custom compression level
	app.Use(compress.New(compress.Config{
		Level: compress.LevelBestSpeed, // 1
	}))

	// Skip middleware for specific routes
	app.Use(compress.New(compress.Config{
		Next: func(c *fiber.Ctx) bool {
			return c.Path() == "/dont_compress"
		},
		Level: compress.LevelBestSpeed, // 1
	}))

	// Default config
	app.Use(cors.New())

	// Or extend your config for customization
	app.Use(cors.New(cors.Config{
		AllowOrigins: "https://gofiber.io, https://gofiber.net",
		AllowHeaders: "Origin, Content-Type, Accept",
	}))

	// Initialize default config
	app.Use(csrf.New())

	// Or extend your config for customization
	app.Use(csrf.New(csrf.Config{
		KeyLookup:      "header:X-Csrf-Token",
		CookieName:     "csrf_",
		CookieSameSite: "Strict",
		Expiration:     1 * time.Hour,
		KeyGenerator:   utils.UUID,
		// Extractor:      func(c *fiber.Ctx) (string, error) { ... },
	}))

	// Login route
	app.Post("/login", login)

	// Unauthenticated route
	app.Get("/", accessible)

	// JWT Middleware
	app.Use(jwtware.New(jwtware.Config{
		SigningMethod: "RS256",
		SigningKey:    privateKey.Public(),
	}))

	// Restricted Routes
	app.Get("/restricted", restricted)

	err = app.Listen(":3000")
	if err != nil {
		log.Fatal(err)
	}
}

func login(c *fiber.Ctx) error {
	user := c.FormValue("user")
	pass := c.FormValue("pass")

	// Throws Unauthorized error
	if user != "john" || pass != "doe" {
		return c.SendStatus(fiber.StatusUnauthorized)
	}

	// Create the Claims
	claims := jwt.MapClaims{
		"name":  "John Doe",
		"admin": true,
		"exp":   time.Now().Add(time.Hour * 72).Unix(),
	}

	// Create token
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)

	// Generate encoded token and send it as response.
	t, err := token.SignedString(privateKey)
	if err != nil {
		log.Printf("token.SignedString: %v", err)
		return c.SendStatus(fiber.StatusInternalServerError)
	}

	return c.JSON(fiber.Map{"token": t})
}

func accessible(c *fiber.Ctx) error {
	return c.SendString("Accessible")
}

func restricted(c *fiber.Ctx) error {
	user := c.Locals("user").(*jwt.Token)
	claims := user.Claims.(jwt.MapClaims)
	name := claims["name"].(string)
	return c.SendString("Welcome " + name)
}
