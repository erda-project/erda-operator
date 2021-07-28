// Copyright (c) 2021 Terminus, Inc.
//
// This program is free software: you can use, redistribute, and/or modify
// it under the terms of the GNU Affero General Public License, version 3
// or later ("AGPL"), as published by the Free Software Foundation.
//
// This program is distributed in the hope that it will be useful, but WITHOUT
// ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or
// FITNESS FOR A PARTICULAR PURPOSE.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <http://www.gnu.org/licenses/>.

package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConvertToHostAlias(t *testing.T) {
	hosts := []string{
		" 127.0.0.1  www.terminus.com",
		" 127.0.0.2       dice.terminus.com     hello.terminus.com",
		" error.terminus.com    ",
	}
	res := ConvertToHostAlias(hosts)
	assert.Equal(t, res[0].IP, "127.0.0.1")
	assert.Equal(t, res[0].Hostnames[0], "www.terminus.com")
	assert.Equal(t, res[1].Hostnames[0], "dice.terminus.com")
	assert.Equal(t, res[1].Hostnames[1], "hello.terminus.com")
	assert.Equal(t, len(res), 2)
}
