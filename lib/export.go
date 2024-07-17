package main

/*
#include "cfunc.h"
#include <stdlib.h>
#include <stdint.h>
#include <string.h>

*/
import "C"
import (
	"encoding/json"
	"strconv"
	"unsafe"

	"github.com/sirupsen/logrus"
	"github.com/stregato/mio/lib/comm"
	"github.com/stregato/mio/lib/core"
	"github.com/stregato/mio/lib/db"
	"github.com/stregato/mio/lib/fs"
	"github.com/stregato/mio/lib/safe"
	"github.com/stregato/mio/lib/security"
	"github.com/stregato/mio/lib/sqlx"
)

func cResult(v any, hnd uint64, err error) C.Result {
	var val []byte

	if err != nil {
		return C.Result{nil, 0, C.ulonglong(hnd), C.CString(err.Error())}
	}
	if v == nil {
		return C.Result{nil, 0, C.ulonglong(hnd), nil}
	}

	val, ok := v.([]byte)
	if !ok {
		val, err = json.Marshal(v)
	}
	if err == nil {
		// Allocate memory in the C heap
		len := C.size_t(len(val))
		ptr := C.malloc(len)
		if ptr == nil {
			return C.Result{nil, 0, C.ulonglong(hnd), C.CString("memory allocation failed")}
		}
		// Copy data from Go slice to C heap
		C.memcpy(ptr, unsafe.Pointer(&val[0]), len)
		return C.Result{ptr, len, C.ulonglong(hnd), nil}
	}
	return C.Result{nil, 0, C.ulonglong(hnd), C.CString(err.Error())}
}

func cInput(err error, i *C.char, v any) error {
	if err != nil {
		return err
	}
	data := C.GoString(i)
	err = json.Unmarshal([]byte(data), v)
	if err != nil {
		err = core.Errorf("failed to unmarshal input - %v: %s", err, data)
	}
	return err
}

var (
	dbs       core.Registry[*sqlx.DB]
	safes     core.Registry[*safe.Safe]
	fss       core.Registry[*fs.FileSystem]
	databases core.Registry[*db.Database]
	rows      core.Registry[*sqlx.Rows]
	comms     core.Registry[*comm.Comm]
)

// mio_setLogLevel sets the log level for the mio library. Possible values are: trace, debug, info, warn, error, fatal, panic.
//
//export mio_setLogLevel
func mio_setLogLevel(level *C.char) C.Result {
	switch C.GoString(level) {
	case "trace":
		logrus.SetLevel(logrus.TraceLevel)
	case "debug":
		logrus.SetLevel(logrus.DebugLevel)
	case "info":
		logrus.SetLevel(logrus.InfoLevel)
	case "warn":
		logrus.SetLevel(logrus.WarnLevel)
	case "error":
		logrus.SetLevel(logrus.ErrorLevel)
	case "fatal":
		logrus.SetLevel(logrus.FatalLevel)
	case "panic":
		logrus.SetLevel(logrus.PanicLevel)
	}
	return cResult(nil, 0, nil)
}

// mio_newIdentity creates a new identity with the specified nick. An identity is a key pair used for encryption and signing, and a nick name for human readable identification.
// An identity is made of two fields ID and Private. ID is a concatenation of the nick name and the public key of the key pair. Private is the private key of the key pair.
//
//export mio_newIdentity
func mio_newIdentity(nick *C.char) C.Result {
	identity, err := security.NewIdentity(C.GoString(nick))
	return cResult(identity, 0, err)
}

// mio_nick returns the nick name of the specified identity.
//
//export mio_nick
func mio_nick(identity *C.char) C.Result {
	var identityG security.Identity
	err := cInput(nil, identity, &identityG)
	if err != nil {
		return cResult(nil, 0, err)
	}
	return cResult(identityG.Id.Nick(), 0, nil)
}

// mio_castID casts the specified string to an Identity ID.
//
//export mio_castID
func mio_castID(id *C.char) C.Result {
	id_, err := security.CastID(C.GoString(id))
	return cResult(id_, 0, err)
}

//export mio_decodeKeys
func mio_decodeKeys(id *C.char) C.Result {
	cryptKey, signKey, err := security.DecodeKeys(C.GoString(id))
	return cResult([]interface{}{cryptKey, signKey}, 0, err)
}

// mio_openDB opens a new database connection to the specified URL.Mio library requires a database connection to store safe and file system data. The function returns a handle to the database connection.
//
//export mio_openDB
func mio_openDB(url *C.char) C.Result {
	var db *sqlx.DB
	var err error

	db, err = sqlx.Open(C.GoString(url))
	if err != nil {
		return cResult(nil, 0, err)
	}

	return cResult(db, dbs.Add(db), err)
}

// mio_closeDB closes the specified database connection.
//
//export mio_closeDB
func mio_closeDB(dbH C.ulonglong) C.Result {
	d, err := dbs.Get(uint64(dbH))
	if err != nil {
		return cResult(nil, 0, err)
	}
	d.Close()
	dbs.Remove(uint64(dbH))
	return cResult(nil, 0, nil)
}

// mio_createSafe creates a new safe with the specified identity, URL and configuration. A safe is a secure storage for keys and files. The function returns a handle to the safe.
//
//export mio_createSafe
func mio_createSafe(dbH C.ulonglong, identity, url, config *C.char) C.Result {
	var identityG security.Identity
	var configG safe.Config

	err := cInput(nil, identity, &identityG)
	if err != nil {
		return cResult(nil, 0, err)
	}

	err = cInput(nil, config, &configG)
	if err != nil {
		return cResult(nil, 0, err)
	}

	d, err := dbs.Get(uint64(dbH))
	if err != nil {
		return cResult(nil, 0, err)
	}

	s, err := safe.Create(d, &identityG, C.GoString(url), configG)
	if err != nil {
		return cResult(nil, 0, err)
	}

	return cResult(s, safes.Add(s), err)
}

// mio_openSafe opens an existing safe with the specified identity and URL. The function returns a handle to the safe.
//
//export mio_openSafe
func mio_openSafe(dbH C.ulonglong, identity, url *C.char) C.Result {
	var identityG security.Identity

	err := cInput(nil, identity, &identityG)
	if err != nil {
		return cResult(nil, 0, err)
	}

	d, err := dbs.Get(uint64(dbH))
	if err != nil {
		return cResult(nil, 0, err)
	}

	s, err := safe.Open(d, &identityG, C.GoString(url))
	if err != nil {
		return cResult(nil, 0, err)
	}

	return cResult(s, safes.Add(s), err)
}

// mio_closeSafe closes the specified safe.
//
//export mio_closeSafe
func mio_closeSafe(safeH C.ulonglong) C.Result {
	s, err := safes.Get(uint64(safeH))
	if err != nil {
		return cResult(nil, 0, err)
	}
	err = s.Close()
	if err != nil {
		return cResult(nil, 0, err)
	}
	safes.Remove(uint64(safeH))
	return cResult(nil, 0, nil)
}

// mio_createGroup applies the specified change to the specified group. The change can be add, remove or update users identified by their IDs.
// The function returns all the groups in the safe after the change.
//
//export mio_updateGroup
func mio_updateGroup(safeH C.ulonglong, groupName *C.char, change C.long, users *C.char) C.Result {
	s, err := safes.Get(uint64(safeH))
	if err != nil {
		return cResult(nil, 0, err)
	}
	var usersG []security.ID
	err = cInput(nil, users, &usersG)
	if err != nil {
		return cResult(nil, 0, err)
	}

	groups, err := s.UpdateGroup(safe.GroupName(C.GoString(groupName)), safe.Change(change), usersG...)
	return cResult(groups, 0, err)
}

// mio_getGroups returns all the groups in the specified safe. It is a map of group names to a list of identity IDs.
//
//export mio_getGroups
func mio_getGroups(safeH C.ulonglong) C.Result {
	s, err := safes.Get(uint64(safeH))
	if err != nil {
		return cResult(nil, 0, err)
	}
	groups, err := s.GetGroups()
	return cResult(groups, 0, err)
}

// mio_getKeys returns all the keys in the specified group. The function returns a list of keys sorted by their creation time.
//
//export mio_getKeys
func mio_getKeys(safeH C.ulonglong, groupName *C.char, expectedMinimumLenght C.long) C.Result {
	s, err := safes.Get(uint64(safeH))
	if err != nil {
		return cResult(nil, 0, err)
	}
	keys, err := s.GetKeys(safe.GroupName(C.GoString(groupName)), int(expectedMinimumLenght))
	return cResult(keys, 0, err)
}

// mio_openFS opens a file system in the specified safe. The function returns a handle to the file system.
//
//export mio_openFS
func mio_openFS(safeH C.ulonglong) C.Result {
	s, err := safes.Get(uint64(safeH))
	if err != nil {
		return cResult(nil, 0, err)
	}
	fs, err := fs.Open(s)
	if err != nil {
		return cResult(nil, 0, err)
	}
	return cResult(fs, fss.Add(fs), err)
}

// mio_closeFS closes the specified file system.
//
//export mio_closeFS
func mio_closeFS(fsH C.ulonglong) C.Result {
	f, err := fss.Get(uint64(fsH))
	if err != nil {
		return cResult(nil, 0, err)
	}
	f.Close()

	fss.Remove(uint64(fsH))
	return cResult(nil, 0, nil)
}

// mio_list returns a list of files in the specified path in the file system. The function returns a list of file information.
//
//export mio_list
func mio_list(fsH C.ulonglong, path, options *C.char) C.Result {
	f, err := fss.Get(uint64(fsH))
	if err != nil {
		return cResult(nil, 0, err)
	}

	var optionsG fs.ListOptions
	err = cInput(err, options, &optionsG)
	if err != nil {
		return cResult(nil, 0, err)
	}

	files, err := f.List(C.GoString(path), optionsG)
	return cResult(files, 0, err)
}

// mio_stat returns the information of the specified file in the file system. The function returns the file information.
//
//export mio_stat
func mio_stat(fsH C.ulonglong, path *C.char) C.Result {
	f, err := fss.Get(uint64(fsH))
	if err != nil {
		return cResult(nil, 0, err)
	}

	file, err := f.Stat(C.GoString(path))
	return cResult(file, 0, err)
}

// mio_putFile puts the specified file in the local filesystem to the specified path in the file system. The function returns the file information.
//
//export mio_putFile
func mio_putFile(fsH C.ulonglong, dest, src, options *C.char) C.Result {
	f, err := fss.Get(uint64(fsH))
	if err != nil {
		return cResult(nil, 0, err)
	}

	var optionsG fs.PutOptions
	err = cInput(err, options, &optionsG)
	if err != nil {
		return cResult(nil, 0, err)
	}

	file, err := f.PutFile(C.GoString(dest), C.GoString(src), optionsG)
	return cResult(file, 0, err)
}

// mio_putData puts the specified data to the specified path in the file system. The function returns the file information.
//
//export mio_putData
func mio_putData(fsH C.ulonglong, dest *C.char, data C.Data, options *C.char) C.Result {
	f, err := fss.Get(uint64(fsH))
	if err != nil {
		return cResult(nil, 0, err)
	}

	var optionsG fs.PutOptions
	err = cInput(err, options, &optionsG)
	if err != nil {
		return cResult(nil, 0, err)
	}

	dataG := C.GoBytes(unsafe.Pointer(data.ptr), C.int(data.len))
	file, err := f.PutData(C.GoString(dest), dataG, optionsG)
	return cResult(file, 0, err)
}

// mio_getFile gets the specified file in the file system and saves it to the specified destination in the local filesystem. The function returns the file information.
//
//export mio_getFile
func mio_getFile(fsH C.ulonglong, src, dest, options *C.char) C.Result {
	f, err := fss.Get(uint64(fsH))
	if err != nil {
		return cResult(nil, 0, err)
	}

	var optionsG fs.GetOptions
	err = cInput(err, options, &optionsG)
	if err != nil {
		return cResult(nil, 0, err)
	}

	file, err := f.GetFile(C.GoString(src), C.GoString(dest), optionsG)
	return cResult(file, 0, err)
}

// mio_getData gets the specified data in the file system. The function returns the data.
//
//export mio_getData
func mio_getData(fsH C.ulonglong, src, options *C.char) C.Result {
	f, err := fss.Get(uint64(fsH))
	if err != nil {
		return cResult(nil, 0, err)
	}

	var optionsG fs.GetOptions
	err = cInput(err, options, &optionsG)
	if err != nil {
		return cResult(nil, 0, err)
	}

	data, err := f.GetData(C.GoString(src), optionsG)
	return cResult(data, 0, err)
}

// mio_delete deletes the specified file in the file system.
//
//export mio_delete
func mio_delete(fsH C.ulonglong, path *C.char) C.Result {
	f, err := fss.Get(uint64(fsH))
	if err != nil {
		return cResult(nil, 0, err)
	}

	err = f.Delete(C.GoString(path))
	return cResult(nil, 0, err)
}

// mio_rename renames the specified file in the file system.
//
//export mio_rename
func mio_rename(fsH C.ulonglong, oldPath, newPath *C.char) C.Result {
	f, err := fss.Get(uint64(fsH))
	if err != nil {
		return cResult(nil, 0, err)
	}

	file, err := f.Rename(C.GoString(oldPath), C.GoString(newPath))
	return cResult(file, 0, err)
}

// mio_openDatabase opens a new database connection to the specified safe using the specified group name and DDLs. The group name defines the users that can access the database.
// The DDLs is a map of version to DDL. The DDL is a string that defines the database schema and should use conditional statements to create or update tables.
// The function returns a handle to the database connection.
//
//export mio_openDatabase
func mio_openDatabase(safeH C.ulonglong, groupName *C.char, ddls *C.char) C.Result {
	s, err := safes.Get(uint64(safeH))
	if err != nil {
		return cResult(nil, 0, err)
	}

	var ddlsG map[string]string
	err = cInput(nil, ddls, &ddlsG)
	if err != nil {
		return cResult(nil, 0, err)
	}

	ddls2 := db.DDLs{}
	for k, v := range ddlsG {
		f, err := strconv.ParseFloat(k, 32)
		if err != nil {
			return cResult(nil, 0, err)
		}

		ddls2[float32(f)] = v
	}

	db, err := db.Open(s, safe.GroupName(C.GoString(groupName)), ddls2)
	return cResult(db, databases.Add(&db), err)
}

// mio_closeDatabase closes the specified database connection.
//
//export mio_closeDatabase
func mio_closeDatabase(dbH C.ulonglong) C.Result {
	d, err := databases.Get(uint64(dbH))
	if err != nil {
		return cResult(nil, 0, err)
	}
	err = d.Close()
	if err != nil {
		return cResult(nil, 0, err)
	}
	databases.Remove(uint64(dbH))
	return cResult(nil, 0, nil)
}

// mio_exec executes the specified SQL statement with the specified arguments in the database. The function returns the number of rows affected.
//
//export mio_exec
func mio_exec(dbH C.ulonglong, query *C.char, args *C.char) C.Result {
	d, err := databases.Get(uint64(dbH))
	if err != nil {
		return cResult(nil, 0, err)
	}

	var argsG sqlx.Args
	err = cInput(nil, args, &argsG)
	if err != nil {
		return cResult(nil, 0, err)
	}

	res, err := d.Exec(C.GoString(query), argsG)
	if err != nil {
		return cResult(nil, 0, err)
	}

	count, _ := res.RowsAffected()
	return cResult(count, 0, err)
}

// mio_query executes the specified SQL query with the specified arguments in the database. The function returns a handle to the result set.
//
//export mio_query
func mio_query(dbH C.ulonglong, key *C.char, args *C.char) C.Result {
	d, err := databases.Get(uint64(dbH))
	if err != nil {
		return cResult(nil, 0, err)
	}

	var argsG sqlx.Args
	err = cInput(nil, args, &argsG)
	if err != nil {
		return cResult(nil, 0, err)
	}

	rows_, err := d.Query(C.GoString(key), argsG)
	if err != nil {
		return cResult(nil, 0, err)
	}

	return cResult(nil, rows.Add(&rows_), err)
}

// mio_nextRow returns the next row in the result set. The function returns the values of the row as a list.
//
//export mio_nextRow
func mio_nextRow(rowsH C.ulonglong) C.Result {
	rows_, err := rows.Get(uint64(rowsH))
	if err != nil {
		return cResult(nil, 0, err)
	}

	values, err := rows_.NextRow()
	if err != nil {
		return cResult(nil, 0, err)
	}

	return cResult(values, 0, nil)
}

// mio_closeRows closes the specified result set.
//
//export mio_closeRows
func mio_closeRows(rowsH C.ulonglong) C.Result {
	rows_, err := rows.Get(uint64(rowsH))
	if err != nil {
		return cResult(nil, 0, err)
	}

	err = rows_.Close()
	if err != nil {
		return cResult(nil, 0, err)
	}

	rows.Remove(uint64(rowsH))
	return cResult(nil, 0, nil)
}

// mio_sync synchronizes the database with the safe. The function returns the number of updates.
//
//export mio_sync
func mio_sync(dbH C.ulonglong) C.Result {
	d, err := databases.Get(uint64(dbH))
	if err != nil {
		return cResult(nil, 0, err)
	}

	updates, err := d.Sync()
	return cResult(updates, 0, err)
}

// mio_cancel cancels the current database operation
//
//export mio_cancel
func mio_cancel(dbH C.ulonglong) C.Result {
	d, err := databases.Get(uint64(dbH))
	if err != nil {
		return cResult(nil, 0, err)
	}

	err = d.Cancel()
	return cResult(nil, 0, err)
}

// mio_openComm opens a point to point communication channel for the specified safe.
//
//export mio_openComm
func mio_openComm(safeH C.ulonglong) C.Result {
	s, err := safes.Get(uint64(safeH))
	if err != nil {
		return cResult(nil, 0, err)
	}

	c := comm.Open(s)
	return cResult(c, comms.Add(c), nil)
}

// mio_rewind rewinds the communication channel to the specified message ID. When calling receive, only messages with a higher ID will be received.
//
//export mio_rewind
func mio_rewind(commH C.ulonglong, dest *C.char, messageID C.ulonglong) C.Result {
	c, err := comms.Get(uint64(commH))
	if err != nil {
		return cResult(nil, 0, err)
	}

	err = c.Rewind(C.GoString(dest), comm.MessageID(messageID))
	return cResult(nil, 0, err)
}

// mio_send sends a message to the specified user.
//
//export mio_send
func mio_send(commH C.ulonglong, userId *C.char, message *C.char) C.Result {
	c, err := comms.Get(uint64(commH))
	if err != nil {
		return cResult(nil, 0, err)
	}

	var messageG comm.Message
	err = cInput(nil, message, &messageG)
	if err != nil {
		return cResult(nil, 0, err)
	}

	err = c.Send(security.ID(C.GoString(userId)), messageG)
	return cResult(nil, 0, err)
}

// mio_broadcast broadcasts a message to the specified group.
//
//export mio_broadcast
func mio_broadcast(commH C.ulonglong, groupName *C.char, message *C.char) C.Result {
	c, err := comms.Get(uint64(commH))
	if err != nil {
		return cResult(nil, 0, err)
	}

	var messageG comm.Message
	err = cInput(nil, message, &messageG)
	if err != nil {
		return cResult(nil, 0, err)
	}

	err = c.Broadcast(safe.GroupName(C.GoString(groupName)), messageG)
	return cResult(nil, 0, err)
}

// mio_receive receives messages from the communication channel that match the specified filter. Filter is either a user ID or a group name.
// When filter is empty, all messages are received.
//
//export mio_receive
func mio_receive(commH C.ulonglong, filter *C.char) C.Result {
	c, err := comms.Get(uint64(commH))
	if err != nil {
		return cResult(nil, 0, err)
	}

	messages, err := c.Receive(C.GoString(filter))
	return cResult(messages, 0, err)
}

// mio_download downloads a file attached to a message to the specified destination in the local filesystem.
//
//export mio_download
func mio_download(commH C.ulonglong, message *C.char, dest *C.char) C.Result {
	c, err := comms.Get(uint64(commH))
	if err != nil {
		return cResult(nil, 0, err)
	}

	var messageG comm.Message
	err = cInput(nil, message, &messageG)
	if err != nil {
		return cResult(nil, 0, err)
	}

	err = c.DownloadFile(messageG, C.GoString(dest))
	return cResult(nil, 0, err)
}
