package main

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
)

const (
	slackAppOauth      = "xoxb-1706344177943-2561742371700-Ysvs4l4Zfq0NeBVvMnvCoUsk"
	slackWebhookURLRaw = "https://hooks.slack.com/services/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"
	channelID          = ""
)

func beginOrder(teamlunch bool, deliveryTime int) {
	rawText := []byte("{\"text\": \"An order has been submitted to the sweet treats diner\"}")

	req, err := http.NewRequest("POST", slackWebhookURLRaw, bytes.NewBuffer(rawText))
	if err != nil {
		log.Panicln("unable to create http request")
	}

	req.Header.Add("Content-type", "application/json")
	req.Header.Add("Authorization", "Bearer "+slackAppOauth)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {

	}
	defer resp.Body.Close()

	dump, _ := httputil.DumpResponse(resp, true)

	fmt.Printf("%+v\n", string(dump))
}

func requestFood(food, orderId string) {
	rawText := []byte(`{
		"channel": "",
		"thread_ts": "",
		"text": "Someone has requested a: "` + food + `
	}`)

	req, err := http.NewRequest("POST", slackWebhookURLRaw, bytes.NewBuffer(rawText))
	if err != nil {
		log.Panicln("unable to create http request")
	}

	req.Header.Add("Content-type", "application/json")
	req.Header.Add("Authorization", "Bearer "+slackAppOauth)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {

	}
	defer resp.Body.Close()

}
