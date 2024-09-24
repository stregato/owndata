
import 'dart:ffi';
import 'dart:io';

import 'package:ffi/ffi.dart';
import 'loader.dart';
import 'package:path/path.dart' as path;


class DB {
  int hnd = 0;

  DB(String dbPath) {
    var fun = stashLibrary!.lookupFunction<ArgsS, ArgsS>('stash_openDB');
    var res = fun(dbPath.toNativeUtf8());
    hnd = res.hnd;
  }

  static DB defaultDB() {
    var homeDir = Platform.environment['HOME'] ?? Platform.environment['USERPROFILE'];
    if (homeDir == null) {
      throw UnsupportedError("Default DB not supported on this platform. Use DB(String dbPath) instead.");
    }
    var dbPath = path.join(homeDir, '.config', 'stash.db');
    return DB(dbPath);
  }

  void close() {
    var fun = stashLibrary!.lookupFunction<ArgsI, Argsi>('stash_closeDB');
    fun(hnd);
  }
} 