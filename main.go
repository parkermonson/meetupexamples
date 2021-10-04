package main

import (
	"fmt"
	"os"

	kingpin "gopkg.in/alecthomas/kingpin.v2"
)

var (
	app = kingpin.New("coolCmd", "Descriptive text that shows up whenever you type \"coolCmd help\"")

	beginOrderCmd   = app.Command("foodorder", "Command to start a food order with our restaurant service, lemon cheesy")
	teamLunchFlag   = beginOrderCmd.Flag("teamlunch", "are you eating alone or ordering for everyone?").Bool()
	deliveryTimeArg = beginOrderCmd.Arg("wait", "When do you want your food delivered? Values of '0' will be ASAP").Int()

	apocalypse      = app.Command("apocalypse", "Update the gitexampledoc.txt file with new info on the apocalypse, without changing it in the filesystem")
	apocalypseField = apocalypse.Arg("field", "Can be countdown, cause or survival_method").String()
	apocalypseText  = apocalypse.Arg("message", "Text which updates the field").String()

	kubemagic = app.Command("kubemagic", "command which runs our kubernetes port-forward example")
)

func main() {

	switch kingpin.MustParse(app.Parse(os.Args[1:])) {
	case beginOrderCmd.FullCommand():
		if deliveryTimeArg == nil {
			*deliveryTimeArg = 0
		}
		beginOrder(*teamLunchFlag, *deliveryTimeArg)
	case apocalypse.FullCommand():
		if *apocalypseField != "countdown" && *apocalypseField != "cause" && *apocalypseField != "survival_method" {
			fmt.Println("invalid apocalype tracker update field, please provide countdown, cause or survival_method as values to field argument")
			return
		}
		updateAndPush(*apocalypseField, *apocalypseText)
	case kubemagic.FullCommand():
		forwardAndWork()
	}

}
