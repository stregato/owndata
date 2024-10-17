package config

import (
	"github.com/stregato/stash/lib/core"
	"github.com/stregato/stash/lib/sqlx"
	"github.com/vmihailenco/msgpack/v5"
)

func GetConfigValue(db *sqlx.DB, domain string, key string) (s string, i int64, v []byte, ok bool) {
	err := db.QueryRow("STASH_GET_CONFIG", sqlx.Args{"node": domain, "key": key}, &s, &i, &v)
	switch err {
	case sqlx.ErrNoRows:
		ok = false
	case nil:
		ok = true
	default:
		core.IsErr(err, "cannot get config for %s/%s: %v", domain, key)
		ok = false
	}
	core.Trace("SQL: STASH_GET_CONFIG: %s/%s - ok=%t, %s, %d, %v", domain, key, ok, s, i, v)
	return s, i, v, ok
}

func SetConfigValue(db *sqlx.DB, domain string, key string, s string, i int64, v []byte) error {
	_, err := db.Exec("STASH_SET_CONFIG", sqlx.Args{"node": domain, "key": key, "s": s, "i": i, "b": v})
	if err != nil {
		return core.Errorw(err, "cannot set config %s/%s with values %s, %d, %v: %v", domain, key, s, i, v)
	}
	core.Trace("SQL: STASH_SET_CONFIG: %s/%s - %s, %d, %v", domain, key, s, i, v)
	return nil
}

func ListConfigKeys(db *sqlx.DB, domain string) ([]string, error) {
	rows, err := db.Query("STASH_LIST_CONFIG", sqlx.Args{"node": domain})
	if err != nil && err != sqlx.ErrNoRows {
		return nil, core.Errorw(err, "cannot list configs for %s: %v", domain, err)
	}
	defer rows.Close()
	var keys []string
	for rows.Next() {
		var key string
		err = rows.Scan(&key)
		if err != nil {
			return nil, core.Errorw(err, "cannot scan config key for %s: %v", domain)
		}
		keys = append(keys, key)
	}
	return keys, nil
}

func GetConfigStruct(db *sqlx.DB, domain string, key string, v interface{}) error {
	_, _, b, ok := GetConfigValue(db, domain, key)
	if ok {
		return msgpack.Unmarshal(b, v)
	}
	return sqlx.ErrNoRows
}

func SetConfigStruct(db *sqlx.DB, domain string, key string, v interface{}) error {
	data, err := msgpack.Marshal(v)
	if err != nil {
		return core.Errorw(err, "cannot marshal config %s/%s: %v", domain, key)
	}
	return SetConfigValue(db, domain, key, "", 0, data)
}

func DelConfigNode(db *sqlx.DB, domain string) error {
	_, err := db.Exec("STASH_DEL_CONFIG", sqlx.Args{"node": domain})
	if err != nil {
		return core.Errorw(err, "cannot del configs %s", domain)
	}
	return err
}
