/*
 * ● YukkiMusic
 * ○ A high-performance engine for streaming music in Telegram voicechats.
 *
 * Copyright (C) 2026 TheTeamVivek
 *
 * This program is free software: you can redistribute it and/or modify it under the
 * terms of the GNU General Public License as published by the Free Software Foundation,
 * either version 3 of the License, or (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful, but WITHOUT ANY
 * WARRANTY; without even the implied warranty of MERCHANTABILITY or FITNESS FOR A
 * PARTICULAR PURPOSE. See the GNU General Public License for more details.
 *
 * Repository: https://github.com/TheTeamVivek/YukkiMusic
 */

package utils

import (
	"fmt"
	"time"

	"github.com/Laky-64/gologging"
	"resty.dev/v3"
)

const batbinBaseURL = "https://batbin.me/"

var httpClient = resty.New().SetTimeout(10 * time.Second)

type batbinResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

func CreatePaste(content string) (string, error) {
	var result batbinResponse

	resp, err := httpClient.R().
		SetBody(content).
		SetResult(&result).
		Post(batbinBaseURL + "api/v2/paste")
	if err != nil {
		gologging.Error("batbin request error: " + err.Error())
		return "", err
	}

	if resp.StatusCode() != 200 {
		gologging.Error("batbin bad response: " + resp.String())
		return "", fmt.Errorf("batbin returned status %d", resp.StatusCode())
	}

	if !result.Success {
		err := fmt.Errorf("batbin paste failed")
		gologging.Error(err.Error())
		return "", err
	}

	return batbinBaseURL + result.Message, nil
}
