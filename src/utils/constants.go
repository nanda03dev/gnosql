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
var CollectionBatchExtension = "-data.gob"
var MaximumLengthNoOfDocuments = 10000

var EVENT_CREATE = "EVENT_CREATE"
var EVENT_UPDATE = "EVENT_UPDATE"
var EVENT_DELETE = "EVENT_DELETE"
var EVENT_SAVE_TO_DISK = "EVENT_SAVE_TO_DISK"
var EVENT_STOP_GO_ROUTINE = "EVENT_STOP_GO_ROUTINE"

var DATABASE_CREATE_SUCCESS_MSG = "Database created successfully"
var DATABASE_DELETE_SUCCESS_MSG = "Database deleted successfully"
var DATABASE_NOT_FOUND_MSG = "Database not found"
var DATABASE_ALREADY_EXISTS_MSG = "Database already exists"
var DATABASE_LOAD_TO_DISK_MSG = "Database to file disk process started"

var COLLECTION_CREATE_SUCCESS_MSG = "Collection created successfully"
var COLLECTION_DELETE_SUCCESS_MSG = "Collection deleted successfully"
var COLLECTION_NOT_FOUND_MSG = "Collection not found "

var DOCUMENT_DELETE_SUCCESS_MSG = "Document deleted successfully"
var DOCUMENT_NOT_FOUND_MSG = "Document not found"

var ERROR_WHILE_BINDING_JSON = "Request JSON binding failed"
var ERROR_WHILE_UNMARSHAL_JSON = "Request JSON Unmarhsall failed"
var ERROR_WHILE_MARSHAL_JSON = "Request JSON Marhsall failed"
