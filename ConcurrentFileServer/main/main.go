package main

import (
	"ConcurrentFileServer/api"
)

func main() {
	apiIns := api.NewAPI()
	apiIns.SetupRoutes()
}
