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

import (
	"encoding/json"
	"fmt"

	"github.com/Paficent/GoFox2X/data"
)

func addValue(arr *data.GFSArray, v any) {
	switch x := v.(type) {
	case json.Number:
		if isIntNumber(x) {
			i, _ := x.Int64()
			arr.AddInt(int(i))
		} else {
			arr.AddUtfString(string(x))
		}
	case string:
		arr.AddUtfString(x)
	default:
		arr.AddUtfString(fmt.Sprint(v))
	}
}

func putJSONValue(obj *data.GFSObject, k string, v any) {
	switch val := v.(type) {
	case json.Number:
		if isIntNumber(val) {
			i, _ := val.Int64()
			obj.PutInt(k, int(i))
		}
	case string:
		obj.PutUtfString(k, val)
	case []any:
		sub := data.MakeGFSArray()
		for _, item := range val {
			addValue(sub, item)
		}
		obj.PutGFSArray(k, sub)
	}
}

func getQuests(db *DB) *data.GFSArray {
	return buildArray(db.Table("quests"), func(r Row) *data.GFSObject {
		id := r.Int("id")
		log := questLogObject(id, "false", 0, r.Int("initial"))
		return questEntryWrap(id, log, questStaticObject(r))
	})
}

func loadQuestStatics(db *DB) (map[int]*data.GFSObject, []int) {
	statics := map[int]*data.GFSObject{}
	var order []int
	for _, r := range db.Table("quests") {
		id := r.Int("id")
		statics[id] = questStaticObject(r)
		order = append(order, id)
	}
	return statics, order
}

func questLogObject(id int, status string, collected, isNew int) *data.GFSObject {
	return data.MakeGFSObject().
		PutInt("id", id).
		PutInt("quest_id", id).
		PutInt("user", 0).
		PutUtfString("status", status).
		PutInt("collected", collected).
		PutInt("new", isNew)
}

func questEntryWrap(id int, log, static *data.GFSObject) *data.GFSObject {
	entry := data.MakeGFSArray()
	entry.AddSFSObject(log)
	entry.AddSFSObject(static)

	return data.MakeGFSObject().
		PutGFSArray("new", entry).
		PutLong("id", int64(id))
}

func questStaticObject(r Row) *data.GFSObject {
	id := r.Int("id")
	static := data.MakeGFSObject().
		PutInt("id", id).
		PutUtfString("name", r.Str("name")).
		PutUtfString("description", r.Str("description")).
		PutUtfString("type", r.Str("type"))

	goals := data.MakeGFSArray()
	for _, g := range r.JSONArray("goals") {
		gm, ok := g.(map[string]any)
		if !ok {
			continue
		}
		goal := data.MakeGFSObject()
		for k, v := range gm {
			putJSONValue(goal, k, v)
		}
		goals.AddSFSObject(goal)
	}
	static.PutGFSArray("goals", goals)

	next := data.MakeGFSArray()
	for _, item := range r.JSONArray("next") {
		if name, ok := item.(string); ok {
			next.AddSFSObject(data.MakeGFSObject().PutUtfString("quest", name))
		}
	}
	static.PutGFSArray("next", next)

	rewards := data.MakeGFSObject()
	if arr := r.JSONArray("rewards"); len(arr) > 0 {
		if m, ok := arr[0].(map[string]any); ok {
			for k, v := range m {
				putJSONValue(rewards, k, v)
			}
		}
	}
	static.PutGFSObject("rewards", rewards)

	static.PutUtfString("sheet", r.Str("sheet")).
		PutUtfString("image", r.Str("image")).
		PutInt("visible", r.Int("visible")).
		PutUtfString("min_server_version", r.Str("min_server_version"))
	if c := r.Str("comment"); c != "" {
		static.PutUtfString("comment", c)
	}
	return static
}

func getTimedEvents(db *DB) *data.GFSArray {
	now := nowMS()
	const oneYearMS = int64(60*60*24*365) * 1000
	endDate := now + oneYearMS

	return buildArray(db.Table("entities"), func(r Row) *data.GFSObject {
		if r.Int("view_in_market") == 1 {
			return nil
		}
		eid := r.Int("entity_id")
		eventData := data.MakeGFSArray()
		eventData.AddSFSObject(data.MakeGFSObject().PutInt("entity", eid))

		return data.MakeGFSObject().
			PutLong("end_date", endDate).
			PutLong("last_updated", now).
			PutUtfString("event_type", "EntityStoreAvailability").
			PutInt("event_id", 3).
			PutGFSArray("data", eventData).
			PutLong("id", int64(200000+eid)).
			PutLong("start_date", now)
	})
}
