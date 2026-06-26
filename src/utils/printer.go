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

package utils

import (
	"fmt"
	"log"
	"strings"

	"github.com/Paficent/GoFox2X/data"
	"github.com/Paficent/GoFox2X/protocol"
)

const sizeLimit = 1000

func Dump(direction string, msg *protocol.Message) {
	body, err := msg.MarshalBinary()
	if err != nil || len(body) >= sizeLimit {
		psize := ""
		if len(body) > 0xFFFF {
			psize = "  >64KB]"
		}

		log.Printf("  send %-30s %6d bytes%s", direction, len(body), psize)
	}

	var b strings.Builder
	fmt.Fprintf(&b, "\n---- %s (%d bytes) ", strings.ToUpper(direction), len(body))
	if cmd, params, ok := protocol.ParseExtensionRequest(msg); ok {
		fmt.Fprintf(&b, "ext %s ----\n", cmd)
		if params != nil {
			b.WriteString(params.Dump(0))
		}
	} else {
		fmt.Fprintf(&b, "sys c=%s a=%d ----\n", msg.Controller, msg.Action)
		if msg.Payload != nil {
			b.WriteString(msg.Payload.Dump(0))
		}
	}
	b.WriteString("\n----")
	log.Print(b.String())
}

func DumpObject(label string, obj *data.GFSObject) {
	if obj == nil {
		return
	}
	log.Printf("\n---- %s ----\n%s\n----", label, obj.Dump(0))
}
