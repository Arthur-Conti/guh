package cli

import (
	"flag"
	"fmt"
	"os"
	"os/exec"

	"github.com/Arthur-Conti/guh/config"
	errorhandler "github.com/Arthur-Conti/guh/libs/error_handler"
	"github.com/Arthur-Conti/guh/libs/log/logger"
	projectconfig "github.com/Arthur-Conti/guh/libs/project_config"
)

const noDefaultGithubURL string = "no_default_github_url"

func Mod() error {
	fs := flag.NewFlagSet("mod", flag.ExitOnError)
	github := fs.String("github", noDefaultGithubURL, "URL to sync github")
	gin := fs.Bool("gin", false, "If true download gin package")
	help := fs.Bool("help", false, "Help with mod command")
	fs.Parse(os.Args[2:])

	if *help {
		HelpMod()
	}

	cfg, err := projectconfig.Load()
	if err != nil {
		return errorhandler.Wrap(errorhandler.KindInternal, "failed to load project config", err, errorhandler.WithOp("mod"))
	}

	if *github != noDefaultGithubURL {
		if err := syncGithub(*github); err != nil {
			config.Config.Logger.Errorf(logger.LogMessage{ApplicationPackage: "cli", Message: "Error syncing with github: %v\n", Vals: []any{err}})
			return err
		}
		if err := projectconfig.Save(cfg); err != nil {
			return errorhandler.Wrap(errorhandler.KindInternal, "failed to save project config", err, errorhandler.WithOp("mod"))
		}
		cfg.ModName = *github
	}
	if *github == noDefaultGithubURL {
		if cfg.ServiceName == "" {
			return errorhandler.New(errorhandler.KindInvalidArgument, "Run 'guh mod --github=...'; alternatively set serviceName in .guh.yaml", errorhandler.WithOp("mod"))
		}
		if err := modConfiguration(cfg.ServiceName); err != nil {
			config.Config.Logger.Errorf(logger.LogMessage{ApplicationPackage: "cli", Message: "Error configurating go mod: %v\n", Vals: []any{err}})
			return err
		}
		if err := projectconfig.Save(cfg); err != nil {
			return errorhandler.Wrap(errorhandler.KindInternal, "failed to save project config", err, errorhandler.WithOp("mod"))
		}
		cfg.ModName = cfg.ServiceName
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
		return errorhandler.Wrap(errorhandler.KindInternal, "Error running 'git init' command", err, errorhandler.WithOp("mod.syncGithub"))
	}
	cmd = exec.Command("git", "remote", "add", "origin", url)
	_, err = cmd.CombinedOutput()
	if err != nil {
		config.Config.Logger.Errorf(logger.LogMessage{ApplicationPackage: "cli", Message: "Error running 'git remote add origin %v' command: %v\n", Vals: []any{url, err}})
		return errorhandler.Wrap(errorhandler.KindInternal, fmt.Sprintf("Error running 'git remote add origin %v' command", url), err, errorhandler.WithOp("mod.syncGithub"))
	}
	cmd = exec.Command("git", "branch", "-M", "main")
	_, err = cmd.CombinedOutput()
	if err != nil {
		config.Config.Logger.Errorf(logger.LogMessage{ApplicationPackage: "cli", Message: "Error running 'git branch -M main' command: %v\n", Vals: []any{err}})
		return errorhandler.Wrap(errorhandler.KindInternal, "Error running 'git branch -M main' command", err, errorhandler.WithOp("mod.syncGithub"))
	}
	return modConfiguration(url)
}

func modConfiguration(modName string) error {
	config.Config.Logger.Infof(logger.LogMessage{ApplicationPackage: "cli", Message: "Configuring go mod with name: %v", Vals: []any{modName}})
	cmd := exec.Command("go", "mod", "init", modName)
	_, err := cmd.CombinedOutput()
	if err != nil {
		config.Config.Logger.Errorf(logger.LogMessage{ApplicationPackage: "cli", Message: "Error running 'go mod init %v' command: %v\n", Vals: []any{modName, err}})
		return errorhandler.Wrap(errorhandler.KindInternal, fmt.Sprintf("Error running 'go mod init %v' command", modName), err, errorhandler.WithOp("mod.modConfiguration"))
	}
	cmd = exec.Command("go", "clean", "-modcache")
	_, err = cmd.CombinedOutput()
	if err != nil {
		config.Config.Logger.Errorf(logger.LogMessage{ApplicationPackage: "cli", Message: "Error running 'go clean -modcache' command: %v\n", Vals: []any{err}})
		return errorhandler.Wrap(errorhandler.KindInternal, "Error running 'go clean -modcache' command", err, errorhandler.WithOp("mod.modConfiguration"))
	}
	cmd = exec.Command("go", "get", "-u", "github.com/Arthur-Conti/guh@v0.1.0")
	_, err = cmd.CombinedOutput()
	if err != nil {
		config.Config.Logger.Errorf(logger.LogMessage{ApplicationPackage: "cli", Message: "Error running 'go get -u github.com/Arthur-Conti/guh@v0.1.0' command: %v\n", Vals: []any{err}})
		return errorhandler.Wrap(errorhandler.KindInternal, "Error running 'go get -u github.com/Arthur-Conti/guh@v0.1.0' command", err, errorhandler.WithOp("mod.modConfiguration"))
	}
	cmd = exec.Command("go", "mod", "tidy")
	_, err = cmd.CombinedOutput()
	if err != nil {
		config.Config.Logger.Errorf(logger.LogMessage{ApplicationPackage: "cli", Message: "Error running 'go mod tidy' command: %v\n", Vals: []any{err}})
		return errorhandler.Wrap(errorhandler.KindInternal, "Error running 'go mod tidy' command", err, errorhandler.WithOp("mod.modConfiguration"))
	}

	return nil
}

func ginDownload() error {
	config.Config.Logger.Info(logger.LogMessage{ApplicationPackage: "cli", Message: "Downloading gin"})
	cmd := exec.Command("go", "get", "-u", "github.com/gin-gonic/gin")
	_, err := cmd.CombinedOutput()
	if err != nil {
		config.Config.Logger.Errorf(logger.LogMessage{ApplicationPackage: "cli", Message: "Error running 'go get -u github.com/gin-gonic/gin' command: %v\n", Vals: []any{err}})
		return errorhandler.Wrap(errorhandler.KindInternal, "Error running 'go get -u github.com/gin-gonic/gin' command", err, errorhandler.WithOp("mod.ginDownload"))
	}
	return nil
}

func HelpMod() {
	fmt.Println(`mod - The mod command help you starting your go mod, connecting it to github and downloanding the GUH files you need 

Usage:
  guh mod [flags]

Flags:
  --github         Provide the github url to initializes your git directory locally, connect it to github and init your go mod with the github url
  --modName        If not connecting to git you can provide a modName to init you go mod
  --gin            Download the gin package

Examples:
  guh mod --github=github/user-example/my-project --gin
  guh mod --modName=my-project

For more information, visit: https://github.com/Arthur-Conti/guh`)
	os.Exit(0)
}
