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

// TODO: if I'm only storing player data in saves shouldn't this just be a utils/savemanager.go?
// TODO: double check theres no sql injection
// NOTE: this was not tested on actual servers, concurency may be an issue
package save

import (
	"database/sql"

	_ "modernc.org/sqlite"
)

type Record struct {
	BBBID int64
	Data  []byte
}

type Store struct {
	db *sql.DB
}

func Open(path string) (*Store, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1)
	if _, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS players (
			bbb_id     INTEGER PRIMARY KEY,
			data       TEXT NOT NULL,
			updated_at INTEGER NOT NULL DEFAULT 0
		)`); err != nil {
		db.Close()
		return nil, err
	}
	return &Store{db: db}, nil
}

func (s *Store) Load() ([]Record, error) {
	rows, err := s.db.Query(`SELECT bbb_id, data FROM players`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []Record
	for rows.Next() {
		var r Record
		var data string
		if err := rows.Scan(&r.BBBID, &data); err != nil {
			return nil, err
		}
		r.Data = []byte(data)
		out = append(out, r)
	}
	return out, rows.Err()
}

func (s *Store) Save(bbbID int64, data []byte, updatedAt int64) error {
	_, err := s.db.Exec(`
		INSERT INTO players (bbb_id, data, updated_at) VALUES (?, ?, ?)
		ON CONFLICT(bbb_id) DO UPDATE SET data = excluded.data, updated_at = excluded.updated_at`,
		bbbID, string(data), updatedAt)
	return err
}

func (s *Store) Close() error {
	return s.db.Close()
}
