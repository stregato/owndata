import ctypes
import platform
import os
import json
from dataclasses import dataclass, asdict
from datetime import datetime
import pytz
import pkg_resources
import base64

def load_lib():
    lib = {
        "Windows": "windows/mio.dll",
        "Linux": "lib/libmio.so",
        "Darwin": "macos/libmio.dylib",
    }
    os_name = platform.system()
    package_dir = pkg_resources.get_distribution('mio').location
    path = os.path.join(package_dir, lib[os_name])
    return ctypes.CDLL(path)

def json_serial(obj):
    """JSON serializer for objects not serializable by default json code"""
    if isinstance(obj, datetime):
        if not obj.tzinfo:
            obj = pytz.utc.localize(obj)
        return obj.isoformat()
    raise TypeError("Type %s not serializable" % type(obj))

def e8(s):
    """encode utf-8 string"""
    return ctypes.c_char_p(s.encode("utf-8"))

def j8(s):
    """encode json object to utf-8 string"""
    return json.dumps(s).encode("utf-8")

def o8(o):
    """encode dataclass object to utf-8 string"""
    if hasattr(o, "toJson"):
        return e8(o.toJson())
    else:
        return json.dumps(asdict(o), default=json_serial).encode("utf-8")

class Result(ctypes.Structure):
    _fields_ = [("val", ctypes.c_char_p), ("hnd", ctypes.c_ulonglong), ("err", ctypes.c_char_p)]

    def __repr__(self):
        return f"Result(val={self.val}, hnd={self.hnd}, err={self.err})"

def r2d(r):
    if r.err:
        raise Exception(r.err.decode("utf-8"))
    if r.val is None:
        return None
    return json.loads(r.val)
    
