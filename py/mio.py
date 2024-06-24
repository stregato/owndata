import json
import os
import appdirs
from options import CreateOptions, OpenOptions, ListOptions, Users, ListDirsOptions, PutOptions, GetOptions, SetUsersOptions
from dataclasses import dataclass
from miod import e8, j8, o8, r2d
from miob import lib
from typing import List, Union
from datetime import datetime
import base64
from urllib.parse import urlparse
from functools import lru_cache

global woland_started
woland_started = False

def set_mio_log_level(level: str):
    r = lib.mio_setLogLevel(e8(level))
    return r2d(r)

class Identity:
    def __init__(self, nick: str):
        r = lib.mio_newIdentity(e8(nick))
        self.id = r2d(r)["i"]
        self.private = r2d(r).get("p")

    @staticmethod
    def fromPrivate(nick: str, privateId: str):
        r = lib.mio_newIdentityFromId(e8(nick), e8(privateId))
        return Identity.fromJson(r2d(r))

    @staticmethod
    def get(id: str):
        r = lib.mio_getIdentity(e8(id))
        return Identity.fromJson(r2d(r))

    def toJson(self):
        return json.dumps({
            "i": self.id,
            "p": self.private
        })

    def __repr__(self) -> str:
        sp = self.id.split(".")
        return "%s [%s]" % (sp[0], sp[1]) if len(sp) > 1 else self.id 

    def __str__(self) -> str:
        return self.id

    @property
    def nick(self):
        return self.id.split(".")[0]


class DB:
    def __init__(self, path: str):
        r = lib.mio_openDB(e8(path))
        for k, v in r2d(r).items():
            setattr(self, k, v)
        self.hnd = r.hnd

    def __del__(self):
        r = lib.mio_closeDB(self.hnd)
        return r2d(r)

    @classmethod
    def default(cls):
        if hasattr(cls, "__default_db"):
            return cls.__default_db
        cls.__default_db = DB(appdirs.user_config_dir("mio.db"))
        return cls.__default_db
    
    def __repr__(self) -> str:
        return self.DbPath
    
@dataclass
class Config:
    "store configuration"
    description: str = ""
    quota: int = 0
    signature: str = ""
    
class Safe():
    grant   = 0
    revoke  = 1
    curse   = 2
    endorse = 3

    @staticmethod
    def create(db: DB, creator: Identity, url: str, config: Config = Config()):
        r = lib.mio_createSafe(db.hnd, o8(creator), e8(url), o8(config))
        s = Safe()
        for k, v in r2d(r).items():
            setattr(s, k, v)
        s.hnd = r.hnd
        s.db = db
        return s

    @staticmethod
    def open(identity: Identity, name: str, storeUrl: str, creatorId: str):
        r = lib.mio_openSafe(o8(identity), e8(name),
                              e8(storeUrl), e8(creatorId))
        s = Safe()
        for k, v in r2d(r).items():
            setattr(s, k, v)
        s.hnd = r.hnd
        return s

    def close(self):
        r = lib.mio_closeSafe(self.hnd)
        return r2d(r)
    
    def update_group(self, groupName: str, change: int, users: Users):
        r = lib.mio_updateGroup(self.hnd, e8(groupName), change, j8(users))
        return r2d(r)

    def get_groups(self):
        r = lib.mio_getGroups(self.hnd)
        return r2d(r)
    
    def get_keys(self, groupName: str, expectedMinimumLenght: int):
        r = lib.mio_getKeys(self.hnd, e8(groupName), expectedMinimumLenght)
        return r2d(r)    

    def fs(self):
        rq = lib.mio_openFS(self.hnd)
        r2d(rq)

        fs = FS()
        fs.hnd = rq.hnd
        fs.safe = self
        return fs
    
    def database(self, groupName: str, ddls: dict[float, str] = {}):
        rq = lib.mio_openDatabase(self.hnd, e8(groupName), j8(ddls))
        r2d(rq)

        database = Database()
        database.hnd = rq.hnd
        database.safe = self
        return database

    def __del__(self):
        r = lib.mio_closeFS(self.hnd)
        return r2d(r)
    
    def __repr__(self) -> str:
        return self.URL

class Database:
    def exec(self, key: str, **args: dict):
        r = lib.mio_exec(self.hnd, e8(key), j8(args))
        return r2d(r)
    
    def query(self, key: str, **args: dict):
        r = lib.mio_query(self.hnd, e8(key), j8(args))
        r2d(r)
        rows = Rows()
        rows.hnd = r.hnd
        return rows
    
    def sync(self):
        r = lib.mio_sync(self.hnd)
        return r2d(r)

class Rows:
    def __iter__(self):
        return self
    
    def __next__(self):
        if self.hnd == 0:
            raise StopIteration
        r = lib.mio_nextRow(self.hnd)
        val = r2d(r)
        if val is None:
            lib.mio_closeRows(self.hnd)
            self.hnd = 0
            raise StopIteration
        else:
            return val
    
    def __del__(self):
        if self.hnd:
            lib.mio_closeRows(self.hnd)

class FS:
    def list(self, path: str = "", listOptions: ListOptions = ListOptions()):
        r = lib.mio_list(self.hnd, e8(path), o8(listOptions))
        ls = r2d(r)
        return ls if ls else []

    def stat(self, path: str):
        r = lib.mio_stat(self.hnd, e8(path))
        return r2d(r)

    def put_file(self, dest: str, src: str, putOptions: PutOptions = PutOptions()):
        "put a file from src to dest"
        r = lib.mio_putFile(self.hnd, e8(dest), e8(src), o8(putOptions))
        return r2d(r)

    def put_data(self, dest: str, data: bytes, putOptions: PutOptions = PutOptions()):
        "put data to dest"
        r = lib.mio_putData(self.hnd, e8(dest), data, o8(putOptions))
        return r2d(r)
    
    def get_file(self, src: str, dest: str, getOptions: GetOptions = GetOptions()):
        "get a file from src to dest"
        r = lib.mio_getFile(self.hnd, e8(src), e8(dest), o8(getOptions))
        return r2d(r)
    
    def get_data(self, src: str, getOptions: GetOptions = GetOptions()):
        "get data from src"
        r = lib.mio_getData(self.hnd, e8(src), o8(getOptions))
        return base64.b64decode(r2d(r))

    def delete(self, path: str):
        "delete a file"
        r = lib.mio_delete(self.hnd, e8(path))
        return r2d(r)
    
    def rename(self, old_path: str, new_path: str):
        "rename a file"
        r = lib.mio_rename(self.hnd, e8(old_path), e8(new_path))
        return r2d(r)

    def __del__(self):
        r = lib.mio_closeFS(self.hnd)
        return r2d(r)

    def __repr__(self) -> str:
        return self.safe.__repr__()

def parseInvite(link: str):
    "parse an invite url"

    try:
        url = urlparse(link)
    except Exception:
        return None

    if not url.path:
        return None

    parts = url.path.split('/')
    if len(parts) != 5 or parts[1] != "a":
        return None

    try:
        # In Python, base64.urlsafe_b64decode automatically handles '-' and '_'
        decoded_url = base64.urlsafe_b64decode(parts[4] + '==').decode('utf-8')
        return [parts[2], parts[3], decoded_url]
    except Exception:
        return None    

def test():
    i = Identity.new("test")
    url = "file:///tmp/poland/{}/test".format(i)
    safe = Safe.create(DB.default(), i, url)
    safe = Safe.open(DB.default(), i, url)
    fs = safe.fs()
    print(fs.list())
    fs.putData("test", b"test")
    print(fs.list())


if __name__ == "__main__":
    test()
