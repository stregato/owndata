package main

/*
#include "cfunc.h"
#include <stdlib.h>
*/
import "C"
import (
	"encoding/json"

	"github.com/stregato/mio/lib/core"
	"github.com/stregato/mio/lib/safe"
	"github.com/stregato/mio/lib/security"
	"github.com/stregato/mio/lib/sqlx"
)

func cResult(v any, err error) C.Result {
	var res []byte

	if err != nil {
		return C.Result{nil, C.CString(err.Error())}
	}
	if v == nil {
		return C.Result{nil, nil}
	}

	res, err = json.Marshal(v)
	if err == nil {
		return C.Result{C.CString(string(res)), nil}
	}
	return C.Result{nil, C.CString(err.Error())}
}

func cInput(err error, i *C.char, v any) error {
	if err != nil {
		return err
	}
	data := C.GoString(i)
	return json.Unmarshal([]byte(data), v)
}

func cUnmarshal(i *C.char, v any) error {
	data := C.GoString(i)
	err := json.Unmarshal([]byte(data), v)
	if core.IsErr(err, "cannot unmarshal %s: %v", data) {
		return err
	}
	return nil
}

var dbs core.Registry[*sqlx.DB]
var safes core.Registry[*safe.Safe]

//export mio_openDB
func mio_openDB(url *C.char) C.Result {
	var db *sqlx.DB
	var err error

	db, err = sqlx.Open(C.GoString(url))
	if err != nil {
		return cResult(nil, err)
	}

	return cResult(dbs.Add(db), err)
}

//export mio_closeDB
func mio_closeDB(dbH int) error {
	d, err := dbs.Get(int(dbH))
	if err != nil {
		return err
	}
	d.Close()
	dbs.Remove(int(dbH))
	return nil
}

//export mio_openSafe
func mio_openSafe(dbH C.long, identity, url *C.char) C.Result {
	var identityG security.Identity

	err := cUnmarshal(identity, &identityG)
	if err != nil {
		return cResult(nil, err)
	}
	d, err := dbs.Get(int(dbH))
	if err != nil {
		return cResult(nil, err)
	}

	s, err := safe.Open(d, &identityG, C.GoString(url))
	if err != nil {
		return cResult(nil, err)
	}

	return cResult(safes.Add(s), err)
}

//export mio_closeSafe
func mio_closeSafe(safeH C.long) C.Result {
	s, err := safes.Get(int(safeH))
	if err != nil {
		return cResult(nil, err)
	}
	err = s.Close()
	if err != nil {
		return cResult(nil, err)
	}
	safes.Remove(int(safeH))
	return C.Result{nil, nil}
}

//export mio_updateGroup
func mio_updateGroup(safeH C.long, groupName *C.char, change C.long, users *C.char) C.Result {
	s, err := safes.Get(int(safeH))
	if err != nil {
		return cResult(nil, err)
	}
	var usersG []security.ID
	err = cUnmarshal(users, &usersG)
	if err != nil {
		return cResult(nil, err)
	}

	groups, err := s.UpdateGroup(safe.GroupName(C.GoString(groupName)), safe.Change(change), usersG...)
	return cResult(groups, err)
}

//export mio_getGroups
func mio_getGroups(safeH C.long) C.Result {
	s, err := safes.Get(int(safeH))
	if err != nil {
		return cResult(nil, err)
	}
	groups, err := s.GetGroups()
	return cResult(groups, err)
}

//export mio_getKeys
func mio_getKeys(safeH C.long, groupName *C.char, expectedMinimumLenght C.long) C.Result {
	s, err := safes.Get(int(safeH))
	if err != nil {
		return cResult(nil, err)
	}
	keys, err := s.GetKeys(safe.GroupName(C.GoString(groupName)), int(expectedMinimumLenght))
	return cResult(keys, err)
}
