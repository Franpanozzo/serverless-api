package handlers

import (
	"encoding/json"
	"log"

	"github.com/aws/aws-lambda-go/events"
)

func apiResponse(status int, body interface{}) (*events.APIGatewayProxyResponse, error) {
	resp := events.APIGatewayProxyResponse{Headers: map[string]string{"Content-Type": "application/json"}}
	resp.StatusCode = status

	log.Printf("Returning status code: %v\n", status)
	stringBody, _ := json.Marshal(body)
	resp.Body = string(stringBody) //parseo los bytes a string
	return &resp, nil
}
