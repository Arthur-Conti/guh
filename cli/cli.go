package cli

func Handle(args []string) error {
	switch args[1] {
	case "compose":
		return Compose()
	}
	return nil
}