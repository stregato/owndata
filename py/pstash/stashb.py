import ctypes
import platform
import os
import json
from dataclasses import dataclass, asdict
from datetime import datetime
from .stashd import *

lib = load_lib()

# Match the functions exported in export.go
lib.stash_setLogLevel.argtypes = [ctypes.c_char_p]
lib.stash_setLogLevel.restype = Result

lib.stash_openDB.argtypes = [ctypes.c_char_p]
lib.stash_openDB.restype = Result

lib.stash_closeDB.argtypes = [ctypes.c_ulonglong]
lib.stash_closeDB.restype = Result

lib.stash_newIdentity.argtypes = [ctypes.c_char_p]
lib.stash_newIdentity.restype = Result

lib.stash_createSafe.argtypes = [ctypes.c_ulonglong, ctypes.c_char_p, ctypes.c_char_p, ctypes.c_char_p]
lib.stash_createSafe.restype = Result

lib.stash_openSafe.argtypes = [ctypes.c_ulonglong, ctypes.c_char_p, ctypes.c_char_p]
lib.stash_openSafe.restype = Result

lib.stash_closeSafe.argtypes = [ctypes.c_ulonglong]
lib.stash_closeSafe.restype = Result

lib.stash_updateGroup.argtypes = [ctypes.c_ulonglong, ctypes.c_char_p, ctypes.c_long, ctypes.c_char_p]
lib.stash_updateGroup.restype = Result

lib.stash_getGroups.argtypes = [ctypes.c_ulonglong]
lib.stash_getGroups.restype = Result

lib.stash_getKeys.argtypes = [ctypes.c_ulonglong, ctypes.c_char_p, ctypes.c_long]
lib.stash_getKeys.restype = Result

lib.stash_openFS.argtypes = [ctypes.c_ulonglong]
lib.stash_openFS.restype = Result

lib.stash_closeFS.argtypes = [ctypes.c_ulonglong]
lib.stash_closeFS.restype = Result

lib.stash_list.argtypes = [ctypes.c_ulonglong, ctypes.c_char_p, ctypes.c_char_p]
lib.stash_list.restype = Result

lib.stash_stat.argtypes = [ctypes.c_ulonglong, ctypes.c_char_p]
lib.stash_stat.restype = Result

lib.stash_putFile.argtypes = [ctypes.c_ulonglong, ctypes.c_char_p, ctypes.c_char_p, ctypes.c_char_p]
lib.stash_putFile.restype = Result

lib.stash_putData.argtypes = [ctypes.c_ulonglong, ctypes.c_char_p, ctypes.c_void_p, ctypes.c_size_t, ctypes.c_char_p]
lib.stash_putData.restype = Result

lib.stash_getFile.argtypes = [ctypes.c_ulonglong, ctypes.c_char_p, ctypes.c_char_p, ctypes.c_char_p]
lib.stash_getFile.restype = Result

lib.stash_getData.argtypes = [ctypes.c_ulonglong, ctypes.c_char_p, ctypes.c_char_p]
lib.stash_getData.restype = Result

lib.stash_delete.argtypes = [ctypes.c_ulonglong, ctypes.c_char_p]
lib.stash_delete.restype = Result

lib.stash_rename.argtypes = [ctypes.c_ulonglong, ctypes.c_char_p, ctypes.c_char_p]
lib.stash_rename.restype = Result

lib.stash_openDatabase.argtypes = [ctypes.c_ulonglong, ctypes.c_char_p, ctypes.c_char_p]
lib.stash_openDatabase.restype = Result

lib.stash_closeDatabase.argtypes = [ctypes.c_ulonglong]
lib.stash_closeDatabase.restype = Result

lib.stash_exec.argtypes = [ctypes.c_ulonglong, ctypes.c_char_p, ctypes.c_char_p]
lib.stash_exec.restype = Result

lib.stash_query.argtypes = [ctypes.c_ulonglong, ctypes.c_char_p, ctypes.c_char_p]
lib.stash_query.restype = Result

lib.stash_nextRow.argtypes = [ctypes.c_ulonglong]
lib.stash_nextRow.restype = Result

lib.stash_closeRows.argtypes = [ctypes.c_ulonglong]
lib.stash_closeRows.restype = Result

lib.stash_sync.argtypes = [ctypes.c_ulonglong]
lib.stash_sync.restype = Result

lib.stash_rewind.argtypes = [ctypes.c_ulonglong, ctypes.c_char_p, ctypes.c_ulonglong]
lib.stash_rewind.restype = Result

lib.stash_cancel.argtypes = [ctypes.c_ulonglong]
lib.stash_cancel.restype = Result

lib.stash_openComm.argtypes = [ctypes.c_ulonglong]
lib.stash_openComm.restype = Result

lib.stash_send.argtypes = [ctypes.c_ulonglong, ctypes.c_char_p, ctypes.c_char_p]
lib.stash_send.restype = Result

lib.stash_broadcast.argtypes = [ctypes.c_ulonglong, ctypes.c_char_p]
lib.stash_broadcast.restype = Result

lib.stash_receive.argtypes = [ctypes.c_ulonglong, ctypes.c_char_p]
lib.stash_receive.restype = Result

lib.stash_download.argtypes = [ctypes.c_ulonglong, ctypes.c_char_p, ctypes.c_char_p]
lib.stash_download.restype = Result

def consume(r):
    try:
        if r.err:
            raise Exception(r.err.decode("utf-8"))
        
        if not r.ptr:
            return None
        
        # Interpret ptr as a byte array of length len
        byte_array = (ctypes.c_ubyte * r.len).from_address(r.ptr)
        byte_data = bytes(byte_array)
        
        # Convert byte array to JSON object
        return json.loads(byte_data)
    
    finally:
        # Free the allocated memory for ptr and err
        if r.ptr:
            lib.free(r.ptr)
            r.ptr = None