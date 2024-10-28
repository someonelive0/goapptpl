package main

import (
	"fmt"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/log"
)

func main() {
	app := fiber.New(fiber.Config{
		CaseSensitive: true,
		StrictRouting: true,
		Immutable:     true,
		ServerHeader:  "gofiber3",
		AppName:       "Test App v1.0.1",
		ReadTimeout:   30 * time.Second,
		WriteTimeout:  30 * time.Second,
		ProxyHeader:   fiber.HeaderXForwardedFor,
		UnescapePath:  false, // default false
	})

	log.SetLevel(log.LevelInfo)
	log.Info("🚀 Server started")

	// Uer Middleware
	// Match any route
	app.Use(func(c fiber.Ctx) error {
		fmt.Println("🥇 Any handler, 匹配任何路由" + c.Path())
		return c.Next()
	})

	// // Match all routes starting with /api
	// app.Use("/api", func(c fiber.Ctx) error {
	// 	fmt.Println("🥈 Second handler")
	// 	return c.Next()
	// })

	// // Match request starting with /api
	// app.Use("/api", func(c fiber.Ctx) error {
	// 	fmt.Println("🥈 third handler")
	// 	return c.Next()
	// })

	// // Match requests starting with /api or /home (multiple-prefix support)
	// app.Use([]string{"/api", "/home"}, func(c fiber.Ctx) error {
	// 	fmt.Println("🥈 Fourth handler")
	// 	return c.Next()
	// })

	// // Attach multiple handlers
	// app.Use("/api", func(c fiber.Ctx) error {
	// 	c.Set("X-Custom-Header", random.String(32))
	// 	fmt.Println("🥈 Fifth handler")
	// 	return c.Next()
	// }, func(c fiber.Ctx) error {
	// 	fmt.Println("🥈 Fifth 2 handler")
	// 	return c.Next()
	// })

	// // API routes
	// // GET /api/register
	// app.Get("/api/*", func(c fiber.Ctx) error {
	// 	msg := fmt.Sprintf("✋ %s", c.Params("*"))
	// 	return c.SendString(msg) // => ✋ register
	// })

	AddMinioHandler(app)
	// AddMinioHandler1(app.Group("/minio"))
	AddMysqlHandler(app.Group("/mysql"))

	// // GET /flights/LAX-SFO
	// app.Get("/flights/:from-:to", func(c fiber.Ctx) error {
	// 	msg := fmt.Sprintf("💸 From: %s, To: %s", c.Params("from"), c.Params("to"))
	// 	return c.SendString(msg) // => 💸 From: LAX, To: SFO
	// })

	// // GET /dictionary.txt
	// app.Get("/:file.:ext", func(c fiber.Ctx) error {
	// 	msg := fmt.Sprintf("📃 %s.%s", c.Params("file"), c.Params("ext"))
	// 	return c.SendString(msg) // => 📃 dictionary.txt
	// })

	// // GET /john/75
	// app.Get("/:name/:age/:gender?", func(c fiber.Ctx) error {
	// 	msg := fmt.Sprintf("👴 %s is %s years old", c.Params("name"), c.Params("age"))
	// 	return c.SendString(msg) // => 👴 john is 75 years old
	// })

	// // GET /john
	// app.Get("/:name", func(c fiber.Ctx) error {
	// 	msg := fmt.Sprintf("Hello, %s 👋!", c.Params("name"))
	// 	return c.SendString(msg) // => Hello john 👋!
	// })

	// Or extend your config for customization
	// Assign the middleware to /metrics
	// and change the Title to `MyService Metrics Page`
	// app.Get("/metrics", monitor.New())

	// data, _ := json.MarshalIndent(app.Stack(), "", "  ")
	// fmt.Println(string(data))
	// data, _ = json.MarshalIndent(app.Config(), "", "  ")
	// fmt.Printf("config: %s\n", data)

	log.Fatal(app.Listen("[::]:3000", fiber.ListenConfig{
		CertFile:    "cert.pem",
		CertKeyFile: "key.pem",
	}))
}
