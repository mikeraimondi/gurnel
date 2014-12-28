package main

func root() gurnelCmd {
	return gurnelCmd{
		f: func(args []string) (err error) {
			displayHelpToc()
			return
		},
	}
}
