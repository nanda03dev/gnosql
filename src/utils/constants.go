package utils

import (
	"os/user"
	"path/filepath"
)

var usr, _ = user.Current()

var GNOSQPATH = "gnosql/db/"

var GNOSQLFULLPATH = filepath.Join(usr.HomeDir, GNOSQPATH)

var DBExtension = "-db.gob"
var CollectionExtension = "-collection.gob"
