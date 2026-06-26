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
	"fmt"
	"sort"
	"strings"

	"github.com/Paficent/GoFox2X/data"
)

func orDefault(s, def string) string {
	if s == "" {
		return def
	}
	return s
}

func getStoreGroups(db *DB) *data.GFSArray {
	now := nowMS()
	return buildArray(db.Table("store_groups"), func(r Row) *data.GFSObject {
		id := r.Int("storegroup_id")
		return data.MakeGFSObject().
			PutInt("id", id).
			PutInt("storegroup_id", id).
			PutUtfString("name", r.Str("group_name")).
			PutInt("currency", r.Int("currency")).
			PutUtfString("group_title", r.Str("group_title")).
			PutLong("last_changed", now).
			PutUtfString("min_server_version", r.Str("min_server_version"))
	})
}

func getStoreCurrencies(db *DB) *data.GFSArray {
	now := nowMS()
	return buildArray(db.Table("store_currency"), func(r Row) *data.GFSObject {
		id := r.Int("storecur_id")
		return data.MakeGFSObject().
			PutInt("storecur_id", id).
			PutInt("id", id).
			PutUtfString("currency_name", r.Str("currency_name")).
			PutInt("starting_amount", r.Int("starting_amount")).
			PutLong("last_changed", now).
			PutUtfString("min_server_version", r.Str("min_server_version"))
	})
}

func getStoreItems(db *DB) *data.GFSArray {
	items := buildArray(db.Table("store_data"), func(r Row) *data.GFSObject {
		cur := r.Str("currency")
		if cur != "coins" && cur != "diamonds" && cur != "food" && cur != "ethereal" {
			return nil
		}
		if strings.Contains(strings.ToLower(r.Str("item_title")), "warm") {
			return nil
		}
		if r.Str("min_server_version") != "0.0" {
			return nil
		}
		maxVal := -1
		if r.Str("max") != "" {
			maxVal = r.Int("max")
		}
		id := r.Int("storeitem_id")
		return data.MakeGFSObject().
			PutInt("id", id).
			PutInt("item_id", id).
			PutUtfString("item_name", r.Str("item_name")).
			PutUtfString("item_title", r.Str("item_title")).
			PutUtfString("item_desc", r.Str("item_desc")).
			PutInt("price", r.Int("price")).
			PutInt("consumable", r.Int("consumable")).
			PutInt("amount", r.Int("amount")).
			PutInt("max", maxVal).
			PutInt("group_id", r.Int("group_id")).
			PutInt("sale_amount", 0).
			PutInt("currency_id", r.Int("currency_id")).
			PutUtfString("sheet_id", r.Str("sheet_id"))
	})

	now := nowMS()

	var castles []Row
	for _, s := range db.Table("structures") {
		if s.Str("structure_type") == "castle" {
			castles = append(castles, s)
		}
	}
	sort.Slice(castles, func(i, j int) bool {
		return castles[i].Int("structure_id") < castles[j].Int("structure_id")
	})
	for _, s := range castles {
		addPermanentStructureItem(db, items, s, now)
	}

	for _, s := range db.Table("structures") {
		if s.Int("structure_id") == 2 {
			addPermanentStructureItem(db, items, s, now)
			break
		}
	}

	addPermanentMonsterItem(db, items, 82, "001_E_rare.bin", now, 1)
	return items
}

func addPermanentStructureItem(db *DB, items *data.GFSArray, s Row, now int64) {
	const offset = 300000
	entity := s.Int("entity")
	e, ok := db.entityByID(entity)
	if !ok {
		return
	}
	storeID := offset + s.Int("structure_id")
	name := e.Str("name")
	itemName := orDefault(name, fmt.Sprintf("STRUCTURE_%d", s.Int("structure_id")))

	item := data.MakeGFSObject().
		PutInt("id", storeID).
		PutInt("item_id", storeID).
		PutUtfString("item_name", itemName).
		PutUtfString("item_title", name).
		PutUtfString("item_desc", e.Str("description"))

	cc, cd, ce := e.Int("cost_coins"), e.Int("cost_diamonds"), e.Int("cost_eth_currency")
	switch {
	case cc > 0:
		item.PutInt("price", cc).PutUtfString("currency", "coins")
	case cd > 0:
		item.PutInt("price", cd).PutUtfString("currency", "diamonds")
	case ce > 0:
		item.PutInt("price", ce).PutUtfString("currency", "ethereal")
	default:
		item.PutInt("price", 0).PutUtfString("currency", "coins")
	}

	item.PutInt("consumable", 0).
		PutInt("amount", 1).
		PutInt("max", 1).
		PutInt("group_id", 1).
		PutInt("sale_amount", 0).
		PutInt("currency_id", 1).
		PutUtfString("sheet_id", "").
		PutUtfString("image_id", "").
		PutUtfString("ios_platform_id", "").
		PutUtfString("android_platform_id", "").
		PutUtfString("amazon_platform_id", "").
		PutLong("last_changed", now).
		PutInt("enabled", 1).
		PutUtfString("min_server_version", orDefault(e.Str("min_server_version"), "0.0"))

	addEntityData(db, item, entity)
	item.PutUtfString("structure_type", s.Str("structure_type")).PutInt("upgrades_to", s.Int("upgrades_to"))
	items.AddSFSObject(item)
}

func addPermanentMonsterItem(db *DB, items *data.GFSArray, monsterID int, binsID string, now int64, maxLimit int) {
	const offset = 100000

	var monster Row
	for _, m := range db.Table("monsters") {
		if m.Int("monster_id") == monsterID {
			monster = m
			break
		}
	}
	if monster == nil {
		return
	}
	entity := monster.Int("entity")
	e, ok := db.entityByID(entity)
	if !ok {
		return
	}
	storeID := offset + monsterID
	name := e.Str("name")
	itemName := orDefault(name, fmt.Sprintf("Monster_%d", monsterID))

	item := data.MakeGFSObject().
		PutInt("id", storeID).
		PutInt("item_id", storeID).
		PutUtfString("item_name", itemName).
		PutUtfString("item_title", name).
		PutUtfString("item_desc", e.Str("description"))

	cc, cd := e.Int("cost_coins"), e.Int("cost_diamonds")
	switch {
	case cc > 0:
		item.PutInt("price", cc).PutInt("currency_id", 1)
	case cd > 0:
		item.PutInt("price", cd).PutInt("currency_id", 2)
	default:
		item.PutInt("price", 0).PutInt("currency_id", 1)
	}

	item.PutInt("consumable", 0).
		PutInt("amount", 1).
		PutInt("max", maxLimit).
		PutInt("group_id", 1).
		PutInt("sale_amount", 0).
		PutUtfString("sheet_id", "").
		PutUtfString("image_id", "").
		PutUtfString("ios_platform_id", "").
		PutUtfString("android_platform_id", "").
		PutUtfString("amazon_platform_id", "").
		PutLong("last_changed", now).
		PutInt("enabled", 1).
		PutUtfString("min_server_version", orDefault(e.Str("min_server_version"), "0.0"))

	if binsID != "" {
		item.PutUtfString("bins_id", binsID).PutUtfString("bin_id", binsID)
	}

	addEntityData(db, item, entity)
	items.AddSFSObject(item)
}
