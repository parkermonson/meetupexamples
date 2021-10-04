package main

import (
	"bytes"
	"log"
	"net/http"
)

const (
	slackAppOauth      = "YOUR_SLACK_TOKEN_HERE"
	slackWebhookURLRaw = "https://hooks.slack.com/services/YOUR_SLACK/WEBHOOK_PATH"
)

//This function just pushes a message to slack using their webhook API
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

}
