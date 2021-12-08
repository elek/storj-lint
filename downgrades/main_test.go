// Copyright (C) 2021 Storj Labs, Inc.
// See LICENSE for copying information.

package downgrades

import (
	"testing"
)

func TestGetModules(t *testing.T) {
	_, err := getModules(".", "nonexistent.mod")
	if err != nil {
		t.FailNow()
	}
}
