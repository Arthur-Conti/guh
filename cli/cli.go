package cli

func Handle(args []string) error {
	switch args[1] {
	case "compose":
		return Compose()
	case "config":
		return Config()
	case "structure":
		return Structure()
	case "mod":
		return Mod()
	case "api":
		return Api()
	case "help":
		Help()
	case "-v":
		PrintVersion()
	}
	return nil
}