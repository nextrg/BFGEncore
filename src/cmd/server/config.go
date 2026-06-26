/* Project Encore: BFG - Localized Private Game Restoration Server
 * Copyright (C) 2026 Paficent <paficent@tutamail.com> & Contributors
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program.  If not, see <https://www.gnu.org/licenses/>.
 */

package main

import (
	"encoding/json"
	"os"
)

type config struct {
	Key        string `json:"key"`
	IV         string `json:"iv"`
	MaxPlayers int    `json:"max_players"`
	ServerIP   string `json:"server_ip"`

	GameAddr   string `json:"game_addr"`
	AuthAddr   string `json:"auth_addr"`
	DBPath     string `json:"db_path"`
	SavePath   string `json:"save_path"`
	DLCPath    string `json:"dlc_path"`
	UsersPath  string `json:"users_path"`
	LogPath    string `json:"log_path"`
	RefreshLog bool   `json:"refresh_log"`
	Debug      bool   `json:"debug"`
}

func loadConfig(path string) (config, error) {
	var c config
	raw, err := os.ReadFile(path)
	if err != nil {
		return c, err
	}
	return c, json.Unmarshal(raw, &c)
}
