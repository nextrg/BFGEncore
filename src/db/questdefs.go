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

type QuestGoal struct {
	Key  string
	Int  int
	Str  string
	List []int
	Eval string
	Num  int
}

type QuestReward struct {
	Coins    int
	Diamonds int
	Food     int
	XP       int
	Eth      int
}

type QuestDef struct {
	ID      int
	Name    string
	Goals   []QuestGoal
	Next    []string
	Rewards QuestReward
	Initial bool
	Visible bool
}

func loadQuestDefs(db *DB) (defs []*QuestDef, byName map[string]*QuestDef, byID map[int]*QuestDef) {
	byName = map[string]*QuestDef{}
	byID = map[int]*QuestDef{}
	for _, r := range db.Table("quests") {
		def := &QuestDef{
			ID:      r.Int("id"),
			Name:    r.Str("name"),
			Initial: r.Int("initial") == 1,
			Visible: !r.Has("visible") || r.Int("visible") != 0,
		}
		for _, g := range r.JSONArray("goals") {
			gm, ok := g.(map[string]any)
			if !ok {
				continue
			}
			goal := QuestGoal{Eval: "==", Num: 1}
			for k, v := range gm {
				switch k {
				case "eval":
					if s, ok := v.(string); ok {
						goal.Eval = s
					}
				case "num":
					goal.Num = numToInt(v)
				default:
					goal.Key = k
					switch val := v.(type) {
					case string:
						goal.Str = val
					case []any:
						for _, it := range val {
							goal.List = append(goal.List, numToInt(it))
						}
					default:
						goal.Int = numToInt(v)
					}
				}
			}
			if goal.Num <= 0 {
				goal.Num = 1
			}
			def.Goals = append(def.Goals, goal)
		}
		for _, n := range r.JSONArray("next") {
			if s, ok := n.(string); ok {
				def.Next = append(def.Next, s)
			}
		}
		for _, rw := range r.JSONArray("rewards") {
			m, ok := rw.(map[string]any)
			if !ok {
				continue
			}
			def.Rewards.Coins += numToInt(m["coins"])
			def.Rewards.Diamonds += numToInt(m["diamonds"])
			def.Rewards.Food += numToInt(m["food"])
			def.Rewards.XP += numToInt(m["xp"])
			def.Rewards.Eth += numToInt(m["ethereal_currency"])
		}
		defs = append(defs, def)
		byName[def.Name] = def
		byID[def.ID] = def
	}
	return defs, byName, byID
}

func loadMonsterEntity(db *DB) map[int]int {
	out := map[int]int{}
	for _, m := range db.Table("monsters") {
		out[m.Int("monster_id")] = m.Int("entity")
	}
	return out
}

func loadStructureEntity(db *DB) map[int]int {
	out := map[int]int{}
	for _, s := range db.Table("structures") {
		out[s.Int("structure_id")] = s.Int("entity")
	}
	return out
}

func loadStructureType(db *DB) map[int]string {
	out := map[int]string{}
	for _, s := range db.Table("structures") {
		out[s.Int("structure_id")] = s.Str("structure_type")
	}
	return out
}

type FoodOption struct {
	Food  int
	Cost  int
	Time  int
	Xp    int
	Label string
}

func loadFoodOptions(db *DB) map[int][]FoodOption {
	out := map[int][]FoodOption{}
	for _, s := range db.Table("structures") {
		opts, ok := s.JSON("extra")["food_options"].([]any)
		if !ok {
			continue
		}
		var list []FoodOption
		for _, o := range opts {
			m, ok := o.(map[string]any)
			if !ok {
				continue
			}
			label, _ := m["label"].(string)
			list = append(list, FoodOption{
				Food:  numToInt(m["food"]),
				Cost:  numToInt(m["cost"]),
				Time:  numToInt(m["time"]),
				Xp:    numToInt(m["xp"]),
				Label: label,
			})
		}
		out[s.Int("structure_id")] = list
	}
	return out
}
