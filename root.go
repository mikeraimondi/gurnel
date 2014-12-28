package main

func rootCmd() gurnelCmd {
	return gurnelCmd{
		f: root,
	}
}

func root(args []string) (err error) {
	displayHelpToc()
	return
}
