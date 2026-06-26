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
	"io/fs"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"paficent/bfg/utils"
)

// returns dlc/<ver> if it exists
func (s *authServer) resolveContentRoot(ver string) (string, bool) {
	versionRoot := filepath.Join(s.contentRoot, ver)
	if info, err := os.Stat(versionRoot); err == nil && info.IsDir() {
		return versionRoot, true
	}
	versions := s.contentVersions()
	if len(versions) == 0 {
		return "", false
	}
	log.Printf("content version %q not found; falling back to %q", ver, versions[0])
	return filepath.Join(s.contentRoot, versions[0]), true
}

// lists the version folders under the dlc root, newest first.
func (s *authServer) contentVersions() []string {
	entries, err := os.ReadDir(s.contentRoot)
	if err != nil {
		return nil
	}
	var versions []string
	for _, e := range entries {
		if e.IsDir() {
			versions = append(versions, e.Name())
		}
	}
	sort.Sort(sort.Reverse(sort.StringSlice(versions)))
	return versions
}

func (s *authServer) handleFilesJSON(w http.ResponseWriter, r *http.Request) {
	root, ok := s.resolveContentRoot(r.PathValue("ver"))
	if !ok {
		errorResponse(w, "Unknown content version and no fallback available.", http.StatusNotFound)
		return
	}

	files := []map[string]any{}
	_ = filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return nil
		}
		rel = filepath.ToSlash(rel)
		sum, err := utils.MD5File(path)
		if err != nil {
			return nil
		}
		files = append(files, map[string]any{"localName": rel, "serverName": rel, "checksum": sum})
		return nil
	})
	writeJSON(w, http.StatusOK, files)
}

func (s *authServer) handleServeFile(w http.ResponseWriter, r *http.Request) {
	root, ok := s.resolveContentRoot(r.PathValue("ver"))
	if !ok {
		http.NotFound(w, r)
		return
	}
	name := filepath.FromSlash(strings.ReplaceAll(r.PathValue("file"), "\\", "/"))
	full := filepath.Join(root, name)

	absRoot, _ := filepath.Abs(root)
	absFull, _ := filepath.Abs(full)
	if !strings.HasPrefix(absFull, absRoot+string(os.PathSeparator)) {
		http.NotFound(w, r)
		return
	}
	if info, err := os.Stat(full); err != nil || info.IsDir() {
		http.NotFound(w, r)
		return
	}
	http.ServeFile(w, r, full)
}

func (s *authServer) handleLoadingMP4(w http.ResponseWriter, r *http.Request) {
	for _, ver := range s.contentVersions() {
		mp4 := filepath.Join(s.contentRoot, ver, "gfx", "BigFishSplashScreen.mp4")
		if info, err := os.Stat(mp4); err == nil && !info.IsDir() {
			http.ServeFile(w, r, mp4)
			return
		}
	}
	http.NotFound(w, r)
}
