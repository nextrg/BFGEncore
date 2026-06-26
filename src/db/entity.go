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

package db

import "github.com/Paficent/GoFox2X/data"

func addEntityData(db *DB, obj *data.GFSObject, entityID int) {
	e, ok := db.entityByID(entityID)
	if !ok {
		return
	}

	obj.PutInt("entity_id", entityID)
	obj.PutUtfString("name", e.Str("name"))
	obj.PutUtfString("description", e.Str("description"))
	obj.PutUtfString("entity_type", e.Str("entity_type"))

	if g := e.JSON("graphic"); g != nil {
		graphic := data.MakeGFSObject()
		putValues(graphic, g)
		obj.PutGFSObject("graphic", graphic)
	}

	buildTime := e.Int("build_time")
	obj.PutInt("size_x", e.Int("size_x"))
	obj.PutInt("size_y", e.Int("size_y"))
	obj.PutInt("level", e.Int("level"))
	obj.PutInt("buildTime", buildTime*1000)
	obj.PutInt("build_time", buildTime*1000)
	obj.PutInt("cost_coins", e.Int("cost_coins"))
	obj.PutInt("cost_eth_currency", e.Int("cost_eth_currency"))
	obj.PutInt("cost_diamonds", e.Int("cost_diamonds"))
	obj.PutInt("cost_sale", e.Int("cost_sale"))

	obj.PutUtfString("keywords", e.Str("keywords"))
	minVer := e.Str("min_server_version")
	if minVer == "" {
		minVer = "0.0"
	}
	obj.PutUtfString("min_server_version", minVer)

	obj.PutBool("movable", e.Bool("movable"))
	obj.PutBool("view_in_market", e.Bool("view_in_market"))
	obj.PutBool("premium", e.Bool("premium"))

	reqArr := data.MakeGFSArray()
	for _, req := range e.JSONArray("requirements") {
		reqObj := data.MakeGFSObject()
		if r, ok := req.(map[string]any); ok {
			reqObj.PutInt("entity", numToInt(r["entity"]))
		} else {
			reqObj.PutInt("entity", numToInt(req))
		}
		reqArr.AddSFSObject(reqObj)
	}
	obj.PutGFSArray("requirements", reqArr)

	yOffset := e.Int("y_offset")
	obj.PutLong("last_changed", nowMS())
	obj.PutInt("xp", e.Int("xp"))
	obj.PutInt("y_offset", yOffset)
	obj.PutInt("sticker_offset", yOffset)
	obj.PutUtfString("fb_object_id", "")
	obj.PutInt("tier", 1)
}
