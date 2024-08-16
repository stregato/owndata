import json
import os
import appdirs
from options import CreateOptions, OpenOptions, ListOptions, Users, ListDirsOptions, PutOptions, GetOptions, SetUsersOptions
from dataclasses import dataclass
from miod import e8, j8, o8, Data
from miob import lib, consume
from typing import List, Union
from datetime import datetime
import base64
from urllib.parse import urlparse
from functools import lru_cache

global woland_started
woland_started = False

def set_mio_log_level(level: str):
    r = lib.mio_setLogLevel(e8(level))
    return consume(r)

class Identity:
    def __init__(self, nick: str):
        r = consume(lib.mio_newIdentity(e8(nick)))
        self.id = r["i"]
        self.private = r["p"]

    @staticmethod
    def fromPrivate(nick: str, privateId: str):
        r = lib.mio_newIdentityFromId(e8(nick), e8(privateId))
        return Identity.fromJson(consume(r))

    @staticmethod
    def get(id: str):
        r = lib.mio_getIdentity(e8(id))
        return Identity.fromJson(consume(r))

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
        for k, v in consume(r).items():
            setattr(self, k, v)
        self.hnd = r.hnd

    def __del__(self):
        r = lib.mio_closeDB(self.hnd)
        return consume(r)

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
        r = lib.mio_createStash(db.hnd, o8(creator), e8(url), o8(config))
        s = Safe()
        for k, v in consume(r).items():
            setattr(s, k, v)
        s.hnd = r.hnd
        s.db = db
        return s

    @staticmethod
    def open(db: DB, identity: Identity, url: str):
        r = lib.mio_openStash(db.hnd, o8(identity), e8(url))
        s = Safe()
        for k, v in consume(r).items():
            setattr(s, k, v)
        s.hnd = r.hnd
        return s

    def close(self):
        r = lib.mio_closeStash(self.hnd)
        return consume(r)
    
    def update_group(self, groupName: str, change: int, users: Users):
        r = lib.mio_updateGroup(self.hnd, e8(groupName), change, j8(users))
        return consume(r)

    def get_groups(self):
        r = lib.mio_getGroups(self.hnd)
        return consume(r)
    
    def get_keys(self, groupName: str, expectedMinimumLenght: int):
        r = lib.mio_getKeys(self.hnd, e8(groupName), expectedMinimumLenght)
        return consume(r)    

    def fs(self):
        rq = lib.mio_openFS(self.hnd)
        consume(rq)

        fs = FS()
        fs.hnd = rq.hnd
        fs.safe = self
        return fs
    
    def comm(self):
        rq = lib.mio_openComm(self.hnd)
        consume(rq)

        comm = Comm()
        comm.hnd = rq.hnd
        comm.safe = self
        return comm
    
    def database(self, groupName: str, ddls: dict[float, str] = {}):
        rq = lib.mio_openDatabase(self.hnd, e8(groupName), j8(ddls))
        consume(rq)

        database = Database()
        database.hnd = rq.hnd
        database.safe = self
        return database

    def __del__(self):
        r = lib.mio_closeFS(self.hnd)
        return consume(r)
    
    def __repr__(self) -> str:
        return self.URL

class Database:
    def exec(self, query: str, **args: dict):
        r = lib.mio_exec(self.hnd, e8(query), j8(args))
        return consume(r)
    
    def query(self, query: str, **args: dict):
        r = lib.mio_query(self.hnd, e8(query), j8(args))
        consume(r)
        rows = Rows()
        rows.hnd = r.hnd
        return rows
    
    def sync(self):
        r = lib.mio_sync(self.hnd)
        return consume(r)

class Rows:
    def __iter__(self):
        return self
    
    def __next__(self):
        if self.hnd == 0:
            raise StopIteration
        r = lib.mio_nextRow(self.hnd)
        val = consume(r)
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
        ls = consume(r)
        return ls if ls else []

    def stat(self, path: str):
        r = lib.mio_stat(self.hnd, e8(path))
        return consume(r)

    def put_file(self, dest: str, src: str, putOptions: PutOptions = PutOptions()):
        "put a file from src to dest"
        r = lib.mio_putFile(self.hnd, e8(dest), e8(src), o8(putOptions))
        return consume(r)

    def put_data(self, dest: str, data: bytes, putOptions: PutOptions = PutOptions()):
        "put data to dest"
        data = Data.from_byte_array(data)
        r = lib.mio_putData(self.hnd, e8(dest), data.ptr, data.len, o8(putOptions))
        return consume(r)
    
    def get_file(self, src: str, dest: str, getOptions: GetOptions = GetOptions()):
        "get a file from src to dest"
        r = lib.mio_getFile(self.hnd, e8(src), e8(dest), o8(getOptions))
        return consume(r)
    
    def get_data(self, src: str, getOptions: GetOptions = GetOptions()):
        "get data from src"
        r = lib.mio_getData(self.hnd, e8(src), o8(getOptions))
        return base64.b64decode(consume(r))

    def delete(self, path: str):
        "delete a file"
        r = lib.mio_delete(self.hnd, e8(path))
        return consume(r)
    
    def rename(self, old_path: str, new_path: str):
        "rename a file"
        r = lib.mio_rename(self.hnd, e8(old_path), e8(new_path))
        return consume(r)

    def __del__(self):
        r = lib.mio_closeFS(self.hnd)
        return consume(r)

    def __repr__(self) -> str:
        return self.safe.__repr__()



class Comm:
    def send(self, identityId: str, text: str = "", data: bytes = b"", file: str = ""):
        'send a message to identityId, message can be text, data or file where file is a path to local filesystem'
        m = dict(Text=text, Data=base64.b64encode(data).decode('utf-8'), File=file)
        r = lib.mio_send(self.hnd, e8(identityId), j8(m))
        return consume(r)
    
    def broadcast(self, groupName: str,text: str = "", data: bytes = b"", file: str = ""):
        'broadcast a message to groupName, message can be text, data or file where file is a path to local filesystem'
        m = dict(Text=text, Data=base64.b64encode(data).decode('utf-8'), File=file)
        print(m)
        r = lib.mio_broadcast(self.hnd, e8(groupName), j8(m))
        return consume(r)
    
    def receive(self, filter: str = ""):
        "receive messages, filter is optional and can be used to filter messages by identity id or group name"
        r = lib.mio_receive(self.hnd, e8(filter))
        return consume(r) or []
    
    def download(self, message: dict, dest: str):
        "download a file from message to a destination on local filesystem"
        r = lib.mio_download(self.hnd, j8(message), e8(dest))
        return consume(r)
    
    def rewind(self, dest:str,  messageId: int):
        r = lib.mio_rewind(self.hnd,  e8(dest), messageId)
        return consume(r)

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
    i = Identity("test")
    db = DB.default()
    url = "file:///tmp/poland/{}/test".format(i)
    safe = Safe.create(db, i, url)
    safe = Safe.open(db, i, url)
    fs = safe.fs()
    print(fs.list())
    fs.putData("test", b"test")
    print(fs.list())


if __name__ == "__main__":
    test()
