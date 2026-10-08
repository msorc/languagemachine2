// Package version holds the version, date and licence strings.
package version

import (
	"runtime/debug"
	"strings"
)

// Summary is the copyright line that -version prints.
const Summary = `Language Machine 2 (C) 2024, 2025 Mikhail Sorochan (msorc@users.sourceforge.net). Distribution permitted subject to GNU GPLv3.
The Language Machine 2 is free software as defined by the GNU GPL and comes with ABSOLUTELY NO WARRANTY.
Language Machine (C) 2005 Peri Hankey (mpah@users.sourceforge.net), GNU GPLv2.`

// Copyright is the licence text that -license prints.
const Copyright = `
 ***************************************************************************
 *        The Language Machine 2 - a toolkit for language and grammar      *
 *                   Copyright (C) 2024 by Mikhail Sorochan                *
 *                      msorc@users.sourceforge.net                        *
 *   Based on original work:                                               *
 *        The Language Machine - a toolkit for language and grammar        *
 *                   Copyright (C) 2005 by Peri Hankey                     *
 *                      mpah@users.sourceforge.net                         *
 *                                                                         *
 *   This program is free software; you can redistribute it and/or modify  *
 *   it under the terms of the GNU General Public License as published by  *
 *   the Free Software Foundation; either version 3 of the License, or     *
 *   (at your option) any later version.                                   *
 *                                                                         *
 *   This program is distributed in the hope that it will be useful,       *
 *   but WITHOUT ANY WARRANTY; without even the implied warranty of        *
 *   MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the         *
 *   GNU General Public License for more details.                          *
 *                                                                         *
 *   You should have received a copy of the GNU General Public License     *
 *   along with this program; if not, visit www.gnu.org or write to the    *
 *   Free Software Foundation, Inc.,                                       *
 *   51 Franklin Street, Fifth Floor, Boston, MA  02110-1301, USA.         *
 ***************************************************************************
`

// The release and its date, used when the build does not say otherwise.
const (
	release     = "0.1.3"
	releaseDate = "20261009"
)

// version and date can be set when building:
//
//	go build -ldflags "-X github.com/msorc/languagemachine2/internal/version.version=0.1.0 -X github.com/msorc/languagemachine2/internal/version.date=20261008" ./cmd/lm
//
// make build does this when VERSION (and optionally DATE) is set.
var (
	version string
	date    string
)

// Version is the version of this build: as set when building, else the
// module version recorded in the binary (go install module@vX.Y.Z), else
// the release.
func Version() string {
	if version != "" {
		return version
	}
	if info, ok := debug.ReadBuildInfo(); ok {
		// a local build records "(devel)" or, from an untagged checkout, a
		// v0.0.0- pseudo-version; neither names a release
		if v := info.Main.Version; v != "" && v != "(devel)" && !strings.HasPrefix(v, "v0.0.0-") {
			return strings.TrimPrefix(v, "v")
		}
	}
	return release
}

// Date is the date of this build as YYYYMMDD, the form lmDate gives the
// rules: as set when building, else the release date.
func Date() string {
	if date != "" {
		return date
	}
	return releaseDate
}
