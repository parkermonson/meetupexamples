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

	requestFoodCmd = app.Command("requestfood", "hey can you get spot me? Swear I'll make it up to you...")
	foodArg        = requestFoodCmd.Arg("food", "name of what you want").String()
	orderIdArg     = requestFoodCmd.Arg("orderId", "hashed id of an existing order. Can be found in the slack message").String()
)

func main() {

	switch kingpin.MustParse(app.Parse(os.Args[1:])) {
	case beginOrderCmd.FullCommand():
		if deliveryTimeArg == nil {
			*deliveryTimeArg = 0
		}
		beginOrder(*teamLunchFlag, *deliveryTimeArg)
	case requestFoodCmd.FullCommand():
		if foodArg == nil {
			fmt.Println("missing food, unable to request nothing")
			return
		}
		requestFood(*foodArg, *orderIdArg)
	}

}
