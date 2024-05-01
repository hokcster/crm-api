package main

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/labstack/echo/v4"
)

type requestLogFormat struct {
	Time     time.Time `json:"time"`
	Level    string    `json:"level"`
	Request  string    `json:"request"`
	Response string    `json:"response"`
}

func logJrpcRequest(c echo.Context, reqBody, resBody []byte) {
	requestBody := string(reqBody)
	responseBody := string(resBody)
	rb := requestBody
	if len(requestBody) > 512 {
		rb = requestBody[:512]
	}
	pb := responseBody
	if len(responseBody) > 512 {
		pb = responseBody[:512]
	}

	log := &requestLogFormat{
		Time:     time.Now(),
		Level:    "debug",
		Request:  rb,
		Response: pb,
	}

	logjson, _ := json.Marshal(log)
	fmt.Println(string(logjson))
}
