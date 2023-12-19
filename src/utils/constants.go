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

var EVENT_CREATE = "EVENT_CREATE"
var EVENT_UPDATE = "EVENT_UPDATE"
var EVENT_DELETE = "EVENT_DELETE"
var EVENT_SAVE_TO_DISK = "EVENT_SAVE_TO_DISK"
