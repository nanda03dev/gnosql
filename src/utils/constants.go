package utils

import (
	"os/user"
	"path/filepath"
)

var usr, _ = user.Current()

const gnoSQLPath = "gnosql/db/"

var GNOSQLFULLPATH = filepath.Join(usr.HomeDir, gnoSQLPath)

const (
	GNOSQLPATH = gnoSQLPath
)
