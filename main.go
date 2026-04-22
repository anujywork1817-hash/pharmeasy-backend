package main

import (
	"pharmeasy-backend/config"
	"pharmeasy-backend/routes"
	"pharmeasy-backend/seeder"
)

func main() {
	config.ConnectDB()
	seeder.SeedMedicines(config.DB) // ✅ seeds only if table is empty
	r := routes.SetupRoutes()
	r.Run(":8080")
}
