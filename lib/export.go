package main

/*
#include "cfunc.h"
#include <stdlib.h>
#include <stdint.h>
#include <string.h>
#include <stdio.h>

*/
import "C"
import (
	"encoding/json"
	"strconv"
	"unsafe"

	"github.com/sirupsen/logrus"
	"github.com/stregato/stash/lib/core"
	"github.com/stregato/stash/lib/db"
	"github.com/stregato/stash/lib/fs"
	"github.com/stregato/stash/lib/messanger"
	"github.com/stregato/stash/lib/safe"
	"github.com/stregato/stash/lib/security"
	"github.com/stregato/stash/lib/sqlx"
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
		logrus.Tracef("cResult: %v", v)

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
	dbs          core.Registry[*sqlx.DB]
	safes        core.Registry[*safe.Safe]
	fss          core.Registry[*fs.FileSystem]
	dbs_         core.Registry[*db.DB]
	transactions core.Registry[*db.Transaction]
	rows         core.Registry[*sqlx.Rows]
	messangers   core.Registry[*messanger.Messenger]
)

// stash_setLogLevel sets the log level for the stash library. Possible values are: trace, debug, info, warn, error, fatal, panic.
//
//export stash_setLogLevel
func stash_setLogLevel(level *C.char) C.Result {
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

//export stash_test
func stash_test(nick *C.char) C.Result {
	print(C.GoString(nick))
	return cResult(nil, 0, nil)
}

// stash_newIdentity creates a new identity with the specified nick. An identity is a key pair used for encryption and signing, and a nick name for human readable identification.
// An identity is made of two fields ID and Private. ID is a concatenation of the nick name and the public key of the key pair. Private is the private key of the key pair.
//
//export stash_newIdentity
func stash_newIdentity(nick *C.char) C.Result {
	identity, err := security.NewIdentity(C.GoString(nick))
	return cResult(identity, 0, err)
}

// stash_nick returns the nick name of the specified identity.
//
//export stash_nick
func stash_nick(identity *C.char) C.Result {
	var identityG security.Identity
	err := cInput(nil, identity, &identityG)
	if err != nil {
		return cResult(nil, 0, err)
	}
	return cResult(identityG.Id.Nick(), 0, nil)
}

// stash_castID casts the specified string to an Identity ID.
//
//export stash_castID
func stash_castID(id *C.char) C.Result {
	id_, err := security.CastID(C.GoString(id))
	return cResult(id_, 0, err)
}

//export stash_decodeKeys
func stash_decodeKeys(id *C.char) C.Result {
	cryptKey, signKey, err := security.DecodeKeys(C.GoString(id))
	return cResult([]interface{}{cryptKey, signKey}, 0, err)
}

// stash_openDB opens a new database connection to the specified URL.Stash library requires a database connection to store safe and file system data. The function returns a handle to the database connection.
//
//export stash_openDB
func stash_openDB(url *C.char) C.Result {
	var db *sqlx.DB
	var err error

	db, err = sqlx.Open(C.GoString(url))
	if err != nil {
		return cResult(nil, 0, err)
	}

	return cResult(db, dbs.Add(db), err)
}

// stash_closeDB closes the specified database connection.
//
//export stash_closeDB
func stash_closeDB(dbH C.ulonglong) C.Result {
	d, err := dbs.Get(uint64(dbH))
	if err != nil {
		return cResult(nil, 0, err)
	}
	d.Close()
	dbs.Remove(uint64(dbH))
	return cResult(nil, 0, nil)
}

// stash_createSafe creates a new safe with the specified identity, URL and configuration. A safe is a secure storage for keys and files. The function returns a handle to the safe.
//
//export stash_createSafe
func stash_createSafe(dbH C.ulonglong, identity, url, config *C.char) C.Result {
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

// stash_openSafe opens an existing safe with the specified identity and URL. The function returns a handle to the safe.
//
//export stash_openSafe
func stash_openSafe(dbH C.ulonglong, identity, url *C.char) C.Result {
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

// stash_closeSafe closes the specified safe.
//
//export stash_closeSafe
func stash_closeSafe(safeH C.ulonglong) C.Result {
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

// stash_createGroup applies the specified change to the specified group. The change can be add, remove or update users identified by their IDs.
// The function returns all the groups in the safe after the change.
//
//export stash_updateGroup
func stash_updateGroup(safeH C.ulonglong, groupName *C.char, change C.long, users *C.char) C.Result {
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

// stash_getGroups returns all the groups in the specified safe. It is a map of group names to a list of identity IDs.
//
//export stash_getGroups
func stash_getGroups(safeH C.ulonglong) C.Result {
	s, err := safes.Get(uint64(safeH))
	if err != nil {
		return cResult(nil, 0, err)
	}
	groups, err := s.GetGroups()
	return cResult(groups, 0, err)
}

// stash_getKeys returns all the keys in the specified group. The function returns a list of keys sorted by their creation time.
//
//export stash_getKeys
func stash_getKeys(safeH C.ulonglong, groupName *C.char, expectedMinimumLenght C.long) C.Result {
	s, err := safes.Get(uint64(safeH))
	if err != nil {
		return cResult(nil, 0, err)
	}
	keys, err := s.GetKeys(safe.GroupName(C.GoString(groupName)), int(expectedMinimumLenght))
	return cResult(keys, 0, err)
}

// stash_openFS opens a file system in the specified safe. The function returns a handle to the file system.
//
//export stash_openFS
func stash_openFS(safeH C.ulonglong) C.Result {
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

// stash_closeFS closes the specified file system.
//
//export stash_closeFS
func stash_closeFS(fsH C.ulonglong) C.Result {
	f, err := fss.Get(uint64(fsH))
	if err != nil {
		return cResult(nil, 0, err)
	}
	f.Close()

	fss.Remove(uint64(fsH))
	return cResult(nil, 0, nil)
}

// stash_list returns a list of files in the specified path in the file system. The function returns a list of file information.
//
//export stash_list
func stash_list(fsH C.ulonglong, path, options *C.char) C.Result {
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

// stash_stat returns the information of the specified file in the file system. The function returns the file information.
//
//export stash_stat
func stash_stat(fsH C.ulonglong, path *C.char) C.Result {
	f, err := fss.Get(uint64(fsH))
	if err != nil {
		return cResult(nil, 0, err)
	}

	file, err := f.Stat(C.GoString(path))
	return cResult(file, 0, err)
}

// stash_putFile puts the specified file in the local filesystem to the specified path in the file system. The function returns the file information.
//
//export stash_putFile
func stash_putFile(fsH C.ulonglong, dest, src, options *C.char) C.Result {
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

// stash_putData puts the specified data to the specified path in the file system. The function returns the file information.
//
//export stash_putData
func stash_putData(fsH C.ulonglong, dest *C.char, data C.Data, options *C.char) C.Result {
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

// stash_getFile gets the specified file in the file system and saves it to the specified destination in the local filesystem. The function returns the file information.
//
//export stash_getFile
func stash_getFile(fsH C.ulonglong, src, dest, options *C.char) C.Result {
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

// stash_getData gets the specified data in the file system. The function returns the data.
//
//export stash_getData
func stash_getData(fsH C.ulonglong, src, options *C.char) C.Result {
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

// stash_delete deletes the specified file in the file system.
//
//export stash_delete
func stash_delete(fsH C.ulonglong, path *C.char) C.Result {
	f, err := fss.Get(uint64(fsH))
	if err != nil {
		return cResult(nil, 0, err)
	}

	err = f.Delete(C.GoString(path))
	return cResult(nil, 0, err)
}

// stash_rename renames the specified file in the file system.
//
//export stash_rename
func stash_rename(fsH C.ulonglong, oldPath, newPath *C.char) C.Result {
	f, err := fss.Get(uint64(fsH))
	if err != nil {
		return cResult(nil, 0, err)
	}

	file, err := f.Rename(C.GoString(oldPath), C.GoString(newPath))
	return cResult(file, 0, err)
}

// stash_openDatabase opens a new database connection to the specified safe using the specified group name and DDLs. The group name defines the users that can access the database.
// The DDLs is a map of version to DDL. The DDL is a string that defines the database schema and should use conditional statements to create or update tables.
// The function returns a handle to the database connection.
//
//export stash_openDatabase
func stash_openDatabase(safeH C.ulonglong, groupName *C.char, ddls *C.char) C.Result {
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
	return cResult(db, dbs_.Add(&db), err)
}

// stash_closeDatabase closes the specified database connection.
//
//export stash_closeDatabase
func stash_closeDatabase(dbH C.ulonglong) C.Result {
	d, err := dbs_.Get(uint64(dbH))
	if err != nil {
		return cResult(nil, 0, err)
	}
	err = d.Close()
	if err != nil {
		return cResult(nil, 0, err)
	}
	dbs_.Remove(uint64(dbH))
	return cResult(nil, 0, nil)
}

// stash_transaction starts a new transaction in the specified database. The function returns a handle to the transaction.
//
//export stash_transaction
func stash_transaction(dbH C.ulonglong) C.Result {
	d, err := dbs_.Get(uint64(dbH))
	if err != nil {
		return cResult(nil, 0, err)
	}

	tx, err := d.Transaction()
	if err != nil {
		return cResult(nil, 0, err)
	}
	id := transactions.Add(tx)
	return cResult(nil, id, nil)
}

// stash_exec executes the specified SQL statement with the specified arguments in the database. The function returns the number of rows affected.
//
//export stash_exec
func stash_exec(txH C.ulonglong, query *C.char, args *C.char) C.Result {
	t, err := transactions.Get(uint64(txH))
	if err != nil {
		return cResult(nil, 0, err)
	}

	var argsG sqlx.Args
	err = cInput(nil, args, &argsG)
	if err != nil {
		return cResult(nil, 0, err)
	}

	res, err := t.Exec(C.GoString(query), argsG)
	if err != nil {
		return cResult(nil, 0, err)
	}

	count, _ := res.RowsAffected()
	return cResult(count, 0, err)
}

// stash_query executes the specified SQL query with the specified arguments in the database. The function returns a handle to the result set.
//
//export stash_query
func stash_query(sqH C.ulonglong, key *C.char, args *C.char) C.Result {
	sq, err := dbs_.Get(uint64(sqH))
	if err != nil {
		return cResult(nil, 0, err)
	}

	var argsG sqlx.Args
	err = cInput(nil, args, &argsG)
	if err != nil {
		return cResult(nil, 0, err)
	}

	rows_, err := sq.Query(C.GoString(key), argsG)
	if err != nil {
		return cResult(nil, 0, err)
	}

	return cResult(nil, rows.Add(&rows_), err)
}

// stash_nextRow returns the next row in the result set. The function returns the values of the row as a list.
//
//export stash_nextRow
func stash_nextRow(rowsH C.ulonglong) C.Result {
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

// stash_closeRows closes the specified result set.
//
//export stash_closeRows
func stash_closeRows(rowsH C.ulonglong) C.Result {
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

// stash_sync synchronizes the database with the safe. The function returns the number of updates.
//
//export stash_sync
func stash_sync(dbH C.ulonglong) C.Result {
	d, err := dbs_.Get(uint64(dbH))
	if err != nil {
		return cResult(nil, 0, err)
	}

	updates, err := d.Sync()
	return cResult(updates, 0, err)
}

// stash_commit commits the current database operation.
//
//export stash_commit
func stash_commit(txH C.ulonglong) C.Result {
	t, err := transactions.Get(uint64(txH))
	if err != nil {
		return cResult(nil, 0, err)
	}

	err = t.Commit()
	transactions.Remove(uint64(txH))
	return cResult(nil, 0, err)
}

// stash_rollback cancels the current database operation
//
//export stash_rollback
func stash_rollback(txH C.ulonglong) C.Result {
	t, err := transactions.Get(uint64(txH))
	if err != nil {
		return cResult(nil, 0, err)
	}

	err = t.Rollback()
	transactions.Remove(uint64(txH))
	return cResult(nil, 0, err)
}

// stash_getCounter returns the value of the specified counter in the database.
//
//export stash_getCounter
func stash_getCounter(dbH C.ulonglong, table, key *C.char) C.Result {
	d, err := dbs_.Get(uint64(dbH))
	if err != nil {
		return cResult(nil, 0, err)
	}

	counter, err := d.GetCounter(C.GoString(table), C.GoString(key))
	return cResult(counter, 0, err)
}

// stash_incCounter increments the specified counter in the database. If the counter does not exist, it is created. The function returns the new value of the counter.
//
//export stash_incCounter
func stash_incCounter(txH C.ulonglong, table *C.char, key *C.char, value C.int) C.Result {
	t, err := transactions.Get(uint64(txH))
	if err != nil {
		return cResult(nil, 0, err)
	}

	err = t.IncCounter(C.GoString(table), C.GoString(key), int(value))
	return cResult(nil, 0, err)
}

// stash_openMessanger opens a point to point communication channel for the specified safe.
//
//export stash_openMessanger
func stash_openMessanger(safeH C.ulonglong) C.Result {
	s, err := safes.Get(uint64(safeH))
	if err != nil {
		return cResult(nil, 0, err)
	}

	c := messanger.Open(s)
	return cResult(c, messangers.Add(c), nil)
}

// stash_rewind rewinds the communication channel to the specified message ID. When calling receive, only messages with a higher ID will be received.
//
//export stash_rewind
func stash_rewind(commH C.ulonglong, dest *C.char, messageID C.ulonglong) C.Result {
	c, err := messangers.Get(uint64(commH))
	if err != nil {
		return cResult(nil, 0, err)
	}

	err = c.Rewind(C.GoString(dest), messanger.MessageID(messageID))
	return cResult(nil, 0, err)
}

// stash_send sends a message to the specified user.
//
//export stash_send
func stash_send(messangerH C.ulonglong, userId *C.char, message *C.char) C.Result {
	c, err := messangers.Get(uint64(messangerH))
	if err != nil {
		return cResult(nil, 0, err)
	}

	var messageG messanger.Message
	err = cInput(nil, message, &messageG)
	if err != nil {
		return cResult(nil, 0, err)
	}

	err = c.Send(security.ID(C.GoString(userId)), messageG)
	return cResult(nil, 0, err)
}

// stash_broadcast broadcasts a message to the specified group.
//
//export stash_broadcast
func stash_broadcast(messangerH C.ulonglong, groupName *C.char, message *C.char) C.Result {
	c, err := messangers.Get(uint64(messangerH))
	if err != nil {
		return cResult(nil, 0, err)
	}

	var messageG messanger.Message
	err = cInput(nil, message, &messageG)
	if err != nil {
		return cResult(nil, 0, err)
	}

	err = c.Broadcast(safe.GroupName(C.GoString(groupName)), messageG)
	return cResult(nil, 0, err)
}

// stash_receive receives messages from the communication channel that match the specified filter. Filter is either a user ID or a group name.
// When filter is empty, all messages are received.
//
//export stash_receive
func stash_receive(messangerH C.ulonglong, filter *C.char) C.Result {
	c, err := messangers.Get(uint64(messangerH))
	if err != nil {
		return cResult(nil, 0, err)
	}

	messages, err := c.Receive(C.GoString(filter))
	return cResult(messages, 0, err)
}

// stash_download downloads a file attached to a message to the specified destination in the local filesystem.
//
//export stash_download
func stash_download(messangerH C.ulonglong, message *C.char, dest *C.char) C.Result {
	c, err := messangers.Get(uint64(messangerH))
	if err != nil {
		return cResult(nil, 0, err)
	}

	var messageG messanger.Message
	err = cInput(nil, message, &messageG)
	if err != nil {
		return cResult(nil, 0, err)
	}

	err = c.DownloadFile(messageG, C.GoString(dest))
	return cResult(nil, 0, err)
}
