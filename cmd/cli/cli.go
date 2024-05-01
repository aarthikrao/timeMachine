package main

import (
	"fmt"
	"os"

	"github.com/common-nighthawk/go-figure"
	"github.com/fatih/color"
	"github.com/spencerdodd/grumble"
)

var (
	// All these values will be updated when the app starts
	SeedNodeAddress      string
	LeaderAddress        string = ""
	ServerLocationLatest *ServerLocation
)

func main() {

	var app = grumble.New(&grumble.Config{
		Name:                  "timMachineCli",
		Description:           "CLI for timeMachine DB",
		HistoryFile:           os.TempDir() + "timeMachineHistory",
		PromptColor:           color.New(color.FgHiWhite),
		HelpHeadlineUnderline: true,
		ErrorColor:            color.New(color.FgHiRed),
		Flags: func(f *grumble.Flags) {
			f.String("s", "seed", "", "seed node for the timeMachine cluster")
		},
	})

	app.OnInit(func(a *grumble.App, flags grumble.FlagMap) error {
		SeedNodeAddress = flags.String("seed")
		fmt.Println("Connecting to", SeedNodeAddress)

		myFigure := figure.NewColorFigure("timeMachine", "doom", "green", false)
		myFigure.Print()
		fmt.Println()

		var err error
		LeaderAddress, err = initialise(SeedNodeAddress)
		if err != nil {
			fmt.Println("Error in initialising the client")
			os.Exit(1)
		}

		color.New(color.FgGreen, color.Bold).Printf("\n 👑 Leader : %s\n", LeaderAddress)
		a.SetPrompt(fmt.Sprintf("%s>", LeaderAddress))
		return nil
	})

	app.AddCommand(&grumble.Command{
		Name: "configure",
		Help: "Used to configure the time machine cluster",

		Args: func(a *grumble.Args) {
			a.Int("shards", "how many shards", grumble.Default(12))
			a.Int("replicas", "no of replicas", grumble.Default(3))
		},

		Run: func(c *grumble.Context) error {
			shards := c.Args["shards"].Value.(int)
			replicas := c.Args["replicas"].Value.(int)
			fmt.Printf("Configuring %d Shards and %d replicas\n", shards, replicas)

			return configure(shards, replicas)
		},
	})

	grumble.Main(app)
}
