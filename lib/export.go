package main

/*
#include "cfunc.h"
#include <stdlib.h>
*/
import "C"
import (
	"encoding/json"
	"strconv"

	"github.com/sirupsen/logrus"
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
		return C.Result{nil, C.ulonglong(hnd), C.CString(err.Error())}
	}
	if v == nil {
		return C.Result{nil, C.ulonglong(hnd), nil}
	}

	val, err = json.Marshal(v)
	if err == nil {
		return C.Result{C.CString(string(val)), C.ulonglong(hnd), nil}
	}
	return C.Result{nil, C.ulonglong(hnd), C.CString(err.Error())}
}

func cInput(err error, i *C.char, v any) error {
	if err != nil {
		return err
	}
	data := C.GoString(i)
	return json.Unmarshal([]byte(data), v)
}

// func cUnmarshal(i *C.char, v any) error {
// 	data := C.GoString(i)
// 	err := json.Unmarshal([]byte(data), v)
// 	if core.IsErr(err, "cannot unmarshal %s: %v", data) {
// 		return err
// 	}
// 	return nil
// }

var (
	dbs       core.Registry[*sqlx.DB]
	safes     core.Registry[*safe.Safe]
	fss       core.Registry[*fs.FileSystem]
	databases core.Registry[*db.Database]
	rows      core.Registry[*sqlx.Rows]
)

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
	return C.Result{nil, 0, nil}
}

//export mio_newIdentity
func mio_newIdentity(nick *C.char) C.Result {
	identity, err := security.NewIdentity(C.GoString(nick))
	return cResult(identity, 0, err)
}

//export mio_nick
func mio_nick(identity *C.char) C.Result {
	var identityG security.Identity
	err := cInput(nil, identity, &identityG)
	if err != nil {
		return cResult(nil, 0, err)
	}
	return cResult(identityG.Id.Nick(), 0, nil)
}

//export mio_newUserId
func mio_newUserId(id *C.char) C.Result {
	id_, err := security.CastID(C.GoString(id))
	return cResult(id_, 0, err)
}

//export mio_decodeKeys
func mio_decodeKeys(id *C.char) C.Result {
	cryptKey, signKey, err := security.DecodeKeys(C.GoString(id))
	return cResult([]interface{}{cryptKey, signKey}, 0, err)
}

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

//export mio_closeDB
func mio_closeDB(dbH C.ulonglong) C.Result {
	d, err := dbs.Get(uint64(dbH))
	if err != nil {
		return cResult(nil, 0, err)
	}
	d.Close()
	dbs.Remove(uint64(dbH))
	return C.Result{nil, 0, nil}
}

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
	return C.Result{nil, 0, nil}
}

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

//export mio_getGroups
func mio_getGroups(safeH C.ulonglong) C.Result {
	s, err := safes.Get(uint64(safeH))
	if err != nil {
		return cResult(nil, 0, err)
	}
	groups, err := s.GetGroups()
	return cResult(groups, 0, err)
}

//export mio_getKeys
func mio_getKeys(safeH C.ulonglong, groupName *C.char, expectedMinimumLenght C.long) C.Result {
	s, err := safes.Get(uint64(safeH))
	if err != nil {
		return cResult(nil, 0, err)
	}
	keys, err := s.GetKeys(safe.GroupName(C.GoString(groupName)), int(expectedMinimumLenght))
	return cResult(keys, 0, err)
}

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

//export mio_stat
func mio_stat(fsH C.ulonglong, path *C.char) C.Result {
	f, err := fss.Get(uint64(fsH))
	if err != nil {
		return cResult(nil, 0, err)
	}

	file, err := f.Stat(C.GoString(path))
	return cResult(file, 0, err)
}

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

//export mio_putData
func mio_putData(fsH C.ulonglong, dest, data, options *C.char) C.Result {
	f, err := fss.Get(uint64(fsH))
	if err != nil {
		return cResult(nil, 0, err)
	}

	var optionsG fs.PutOptions
	err = cInput(err, options, &optionsG)
	if err != nil {
		return cResult(nil, 0, err)
	}

	file, err := f.PutData(C.GoString(dest), []byte(C.GoString(data)), optionsG)
	return cResult(file, 0, err)
}

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

//export mio_delete
func mio_delete(fsH C.ulonglong, path *C.char) C.Result {
	f, err := fss.Get(uint64(fsH))
	if err != nil {
		return cResult(nil, 0, err)
	}

	err = f.Delete(C.GoString(path))
	return cResult(nil, 0, err)
}

//export mio_rename
func mio_rename(fsH C.ulonglong, oldPath, newPath *C.char) C.Result {
	f, err := fss.Get(uint64(fsH))
	if err != nil {
		return cResult(nil, 0, err)
	}

	file, err := f.Rename(C.GoString(oldPath), C.GoString(newPath))
	return cResult(file, 0, err)
}

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

//export mio_exec
func mio_exec(dbH C.ulonglong, key *C.char, args *C.char) C.Result {
	d, err := databases.Get(uint64(dbH))
	if err != nil {
		return cResult(nil, 0, err)
	}

	var argsG sqlx.Args
	err = cInput(nil, args, &argsG)
	if err != nil {
		return cResult(nil, 0, err)
	}

	res, err := d.Exec(C.GoString(key), argsG)
	if err != nil {
		return cResult(nil, 0, err)
	}

	count, _ := res.RowsAffected()
	return cResult(count, 0, err)
}

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

//export mio_sync
func mio_sync(dbH C.ulonglong) C.Result {
	d, err := databases.Get(uint64(dbH))
	if err != nil {
		return cResult(nil, 0, err)
	}

	updates, err := d.Sync()
	return cResult(updates, 0, err)
}

//export mio_cancel
func mio_cancel(dbH C.ulonglong) C.Result {
	d, err := databases.Get(uint64(dbH))
	if err != nil {
		return cResult(nil, 0, err)
	}

	err = d.Cancel()
	return cResult(nil, 0, err)
}
