package cli

import (
	"flag"
	"fmt"
	"os"
	"os/exec"

	"github.com/Arthur-Conti/guh/config"
	errorhandler "github.com/Arthur-Conti/guh/libs/error_handler"
	"github.com/Arthur-Conti/guh/libs/log/logger"
)

const noDefaultGithubURL string = "no_default_github_url"
const noModName string = "no_mod_name"

func Mod() error {
	fs := flag.NewFlagSet("mod", flag.ExitOnError)
	github := fs.String("github", noDefaultGithubURL, "URL to sync github")
	gin := fs.Bool("gin", false, "If true download gin package")
	modName := fs.String("modName", noModName, "Mod name if not syncing with github")
	fs.Parse(os.Args[2:])

	if *github != noDefaultGithubURL {
		if err := syncGithub(*github); err != nil {
			config.Config.Logger.Errorf(logger.LogMessage{ApplicationPackage: "cli", Message: "Error syncing with github: %v\n", Vals: []any{err}})
			return err
		}
	}
	if *github == noDefaultGithubURL && *modName != noModName {
		if err := modConfiguration(*modName); err != nil {
			config.Config.Logger.Errorf(logger.LogMessage{ApplicationPackage: "cli", Message: "Error configurating go mod: %v\n", Vals: []any{err}})
			return err
		}
	}
	if *gin {
		if err := ginDownload(); err != nil {
			config.Config.Logger.Errorf(logger.LogMessage{ApplicationPackage: "cli", Message: "Error downloading gin: %v\n", Vals: []any{err}})
			return err
		}
	}

	return nil
}

func syncGithub(url string) error {
	config.Config.Logger.Infof(logger.LogMessage{ApplicationPackage: "cli", Message: "Syncing with github url : %v", Vals: []any{url}})
	cmd := exec.Command("git", "init")
	_, err := cmd.CombinedOutput()
	if err != nil {
		config.Config.Logger.Errorf(logger.LogMessage{ApplicationPackage: "cli", Message: "Error running 'git init' command: %v\n", Vals: []any{err}})
		return errorhandler.Wrap("InternalServerError", "Error running 'git init' command", err)
	}
	cmd = exec.Command("git", "remote", "add", "origin", url)
	_, err = cmd.CombinedOutput()
	if err != nil {
		config.Config.Logger.Errorf(logger.LogMessage{ApplicationPackage: "cli", Message: "Error running 'git remote add origin %v' command: %v\n", Vals: []any{url, err}})
		return errorhandler.Wrap("InternalServerError", fmt.Sprintf("Error running 'git remote add origin %v' command", url), err)
	}
	cmd = exec.Command("git", "branch", "-M", "main", url)
	_, err = cmd.CombinedOutput()
	if err != nil {
		config.Config.Logger.Errorf(logger.LogMessage{ApplicationPackage: "cli", Message: "Error running 'git branch -M main' command: %v\n", Vals: []any{err}})
		return errorhandler.Wrap("InternalServerError", "Error running 'git branch -M main' command", err)
	}
	return modConfiguration(url)
}

func modConfiguration(modName string) error {
	config.Config.Logger.Infof(logger.LogMessage{ApplicationPackage: "cli", Message: "Configuring go mod with name: %v", Vals: []any{modName}})
	cmd := exec.Command("go", "mod", "init", modName)
	_, err := cmd.CombinedOutput()
	if err != nil {
		config.Config.Logger.Errorf(logger.LogMessage{ApplicationPackage: "cli", Message: "Error running 'go mod init %v' command: %v\n", Vals: []any{modName, err}})
		return errorhandler.Wrap("InternalServerError", fmt.Sprintf("Error running 'go mod init %v' command", modName), err)
	}
	cmd = exec.Command("go", "clean", "-modcache")
	_, err = cmd.CombinedOutput()
	if err != nil {
		config.Config.Logger.Errorf(logger.LogMessage{ApplicationPackage: "cli", Message: "Error running 'go clean -modcache' command: %v\n", Vals: []any{err}})
		return errorhandler.Wrap("InternalServerError", "Error running 'go clean -modcache' command", err)
	}
	cmd = exec.Command("go", "get", "-u", "github.com/Arthur-Conti/guh")
	_, err = cmd.CombinedOutput()
	if err != nil {
		config.Config.Logger.Errorf(logger.LogMessage{ApplicationPackage: "cli", Message: "Error running 'go get -u github.com/Arthur-Conti/guh' command: %v\n", Vals: []any{err}})
		return errorhandler.Wrap("InternalServerError", "Error running 'go get -u github.com/Arthur-Conti/guh' command", err)
	}
	cmd = exec.Command("go", "mod", "tidy")
	_, err = cmd.CombinedOutput()
	if err != nil {
		config.Config.Logger.Errorf(logger.LogMessage{ApplicationPackage: "cli", Message: "Error running 'go mod tidy' command: %v\n", Vals: []any{err}})
		return errorhandler.Wrap("InternalServerError", "Error running 'go mod tidy' command", err)
	}

	return nil
}

func ginDownload() error {
	config.Config.Logger.Info(logger.LogMessage{ApplicationPackage: "cli", Message: "Downloading gin"})
	cmd := exec.Command("go", "get", "-u", "github.com/gin-gonic/gin")
	_, err := cmd.CombinedOutput()
	if err != nil {
		config.Config.Logger.Errorf(logger.LogMessage{ApplicationPackage: "cli", Message: "Error running 'go get -u github.com/gin-gonic/gin' command: %v\n", Vals: []any{err}})
		return errorhandler.Wrap("InternalServerError", "Error running 'go get -u github.com/gin-gonic/gin' command", err)
	}
	return nil
}
