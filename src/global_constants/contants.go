package global_constants

import (
	"os/user"
	"path/filepath"
	"time"
)

var usr, _ = user.Current()

// Gnosql Path & Extensions
const GNOSQL_PATH = "gnosql/db/"

var GNOSQL_FULL_PATH = filepath.Join(usr.HomeDir, GNOSQL_PATH)

const DB_EXTENSION = "-db.gob"

const COLLECTION_EXTENSION = "-collection.gob"
const COLLECTION_BATCH_EXTENSION = "-data.gob"
const DOC_ID = "docId"
const DOC_INDEX = "docIndex"
const DOC_CREATED_AT = "created"
const COLLECTION_NAME = "CollectionName"
const INDEX_KEYS_NAME = "IndexKeys"
const FILTER_LIMIT = "limit"
const FILTER_KEY = "key"
const FILTER_VALUE = "value"

// Size % Limits
const INCOME_REQUEST_CHANNEL_SIZE = 100000
const BATCH_SIZE = 10000
const COLLECTION_CHANNEL_SIZE = 10000
const TIME_INTERVAL_TO_SYNC_DISK = 30 * time.Second
const FILTER_DEFAULT_LIMIT int = 1000
const FILTER_DEFAULT_WORKER_COUNT int = 4

// Events
const EVENT_CREATE = "EVENT_CREATE"
const EVENT_UPDATE = "EVENT_UPDATE"
const EVENT_DELETE = "EVENT_DELETE"
const EVENT_SAVE_TO_DISK = "EVENT_SAVE_TO_DISK"
const EVENT_STOP_GO_ROUTINE = "EVENT_STOP_GO_ROUTINE"

// Response Messages
const DATABASE_CREATE_SUCCESS_MSG = "Database created successfully"
const DATABASE_DELETE_SUCCESS_MSG = "Database deleted successfully"
const DATABASE_NOT_FOUND_MSG = "Database not found"
const DATABASE_ALREADY_EXISTS_MSG = "Database already exists"
const DATABASE_LOAD_TO_DISK_MSG = "Database to file disk process started"

const COLLECTION_CREATE_SUCCESS_MSG = "Collection created successfully"
const COLLECTION_DELETE_SUCCESS_MSG = "Collection deleted successfully"
const COLLECTION_NOT_FOUND_MSG = "Collection not found "

const DOCUMENT_DELETE_SUCCESS_MSG = "Document deleted successfully"
const DOCUMENT_NOT_FOUND_MSG = "Document not found"

// Error Response Messages
const ERROR_WHILE_BINDING_JSON = "Request JSON binding failed"
const ERROR_WHILE_UNMARSHAL_JSON = "Request JSON Unmarhsall failed"
const ERROR_WHILE_MARSHAL_JSON = "Request JSON Marhsall failed"

// Divider's
const COLLECTION_CHANNEL_NAME_DIVIDER = "-&-"
