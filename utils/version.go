package utils

import (
	"fmt"
)

// Global version infomation

const (
	APP_VERSION = "1.0.0"

	// date +%FT%T%z  // date +'%Y%m%d'
	BUILD_TIME = "2024-10-28T00:00:00+0800"

	// go version
	GO_VERSION = "1.22.0"

	APP_BANNER = `
	░▒▓██████▓▒░ ░▒▓██████▓▒░ ░▒▓██████▓▒░░▒▓███████▓▒░░▒▓███████▓▒░  
	░▒▓█▓▒░░▒▓█▓▒░▒▓█▓▒░░▒▓█▓▒░▒▓█▓▒░░▒▓█▓▒░▒▓█▓▒░░▒▓█▓▒░▒▓█▓▒░░▒▓█▓▒░ 
	░▒▓█▓▒░      ░▒▓█▓▒░░▒▓█▓▒░▒▓█▓▒░░▒▓█▓▒░▒▓█▓▒░░▒▓█▓▒░▒▓█▓▒░░▒▓█▓▒░ 
	░▒▓█▓▒▒▓███▓▒░▒▓█▓▒░░▒▓█▓▒░▒▓████████▓▒░▒▓███████▓▒░░▒▓███████▓▒░  
	░▒▓█▓▒░░▒▓█▓▒░▒▓█▓▒░░▒▓█▓▒░▒▓█▓▒░░▒▓█▓▒░▒▓█▓▒░      ░▒▓█▓▒░        
	░▒▓█▓▒░░▒▓█▓▒░▒▓█▓▒░░▒▓█▓▒░▒▓█▓▒░░▒▓█▓▒░▒▓█▓▒░      ░▒▓█▓▒░        
	 ░▒▓██████▓▒░ ░▒▓██████▓▒░░▒▓█▓▒░░▒▓█▓▒░▒▓█▓▒░      ░▒▓█▓▒░        
`
)

func Version(app string) string {
	return fmt.Sprintf(`{"app": "%s", "version": "%s", "build_time": "%s", "go_version": "%s"}`,
		app, APP_VERSION, BUILD_TIME, GO_VERSION)
}

func ShowBanner() {
	fmt.Printf("%s\n", APP_BANNER)
	fmt.Printf("goapptpl %s  Copyright (C) 2024 SOMEONELIVE\n", APP_VERSION)
}

func ShowBannerForApp(app, version, build_time string) {
	fmt.Printf("%s\n", APP_BANNER)
	fmt.Printf("Copyright (C) 2024 SOMEONELIVE\n")
	fmt.Printf("%s version %s, build on %s\n\n", app, version, build_time)
}
