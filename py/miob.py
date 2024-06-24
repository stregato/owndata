import ctypes
import platform
import os
import json
from dataclasses import dataclass, asdict
from datetime import datetime
from miod import *

lib = load_lib()

# Match the functions exported in export.go
lib.mio_setLogLevel.argtypes = [ctypes.c_char_p]
lib.mio_setLogLevel.restype = Result

lib.mio_openDB.argtypes = [ctypes.c_char_p]
lib.mio_openDB.restype = Result

lib.mio_closeDB.argtypes = [ctypes.c_ulonglong]
lib.mio_closeDB.restype = Result

lib.mio_newIdentity.argtypes = [ctypes.c_char_p]
lib.mio_newIdentity.restype = Result

lib.mio_createSafe.argtypes = [ctypes.c_ulonglong, ctypes.c_char_p, ctypes.c_char_p]
lib.mio_createSafe.restype = Result

lib.mio_openSafe.argtypes = [ctypes.c_ulonglong, ctypes.c_char_p, ctypes.c_char_p]
lib.mio_openSafe.restype = Result

lib.mio_closeSafe.argtypes = [ctypes.c_ulonglong]
lib.mio_closeSafe.restype = Result

lib.mio_updateGroup.argtypes = [ctypes.c_ulonglong, ctypes.c_char_p, ctypes.c_long, ctypes.c_char_p]
lib.mio_updateGroup.restype = Result

lib.mio_getGroups.argtypes = [ctypes.c_ulonglong]
lib.mio_getGroups.restype = Result

lib.mio_getKeys.argtypes = [ctypes.c_ulonglong, ctypes.c_char_p, ctypes.c_long]
lib.mio_getKeys.restype = Result

lib.mio_openFS.argtypes = [ctypes.c_ulonglong]
lib.mio_openFS.restype = Result

lib.mio_closeFS.argtypes = [ctypes.c_ulonglong]
lib.mio_closeFS.restype = Result

lib.mio_list.argtypes = [ctypes.c_ulonglong, ctypes.c_char_p, ctypes.c_char_p]
lib.mio_list.restype = Result

lib.mio_stat.argtypes = [ctypes.c_ulonglong, ctypes.c_char_p]
lib.mio_stat.restype = Result

lib.mio_putFile.argtypes = [ctypes.c_ulonglong, ctypes.c_char_p, ctypes.c_char_p, ctypes.c_char_p]
lib.mio_putFile.restype = Result

lib.mio_putData.argtypes = [ctypes.c_ulonglong, ctypes.c_char_p, ctypes.c_char_p, ctypes.c_char_p]
lib.mio_putData.restype = Result

lib.mio_getFile.argtypes = [ctypes.c_ulonglong, ctypes.c_char_p, ctypes.c_char_p, ctypes.c_char_p]
lib.mio_getFile.restype = Result

lib.mio_getData.argtypes = [ctypes.c_ulonglong, ctypes.c_char_p, ctypes.c_char_p]
lib.mio_getData.restype = Result

lib.mio_delete.argtypes = [ctypes.c_ulonglong, ctypes.c_char_p]
lib.mio_delete.restype = Result

lib.mio_rename.argtypes = [ctypes.c_ulonglong, ctypes.c_char_p, ctypes.c_char_p]
lib.mio_rename.restype = Result

lib.mio_openDatabase.argtypes = [ctypes.c_ulonglong, ctypes.c_char_p, ctypes.c_char_p]
lib.mio_openDatabase.restype = Result

lib.mio_closeDatabase.argtypes = [ctypes.c_ulonglong]
lib.mio_closeDatabase.restype = Result

lib.mio_exec.argtypes = [ctypes.c_ulonglong, ctypes.c_char_p, ctypes.c_char_p]
lib.mio_exec.restype = Result

lib.mio_query.argtypes = [ctypes.c_ulonglong, ctypes.c_char_p, ctypes.c_char_p]
lib.mio_query.restype = Result

lib.mio_nextRow.argtypes = [ctypes.c_ulonglong]
lib.mio_nextRow.restype = Result

lib.mio_closeRows.argtypes = [ctypes.c_ulonglong]
lib.mio_closeRows.restype = Result

lib.mio_sync.argtypes = [ctypes.c_ulonglong]
lib.mio_sync.restype = Result

lib.mio_cancel.argtypes = [ctypes.c_ulonglong]
lib.mio_cancel.restype = Result