
import 'dart:ffi';
import 'dart:io';

import 'package:ffi/ffi.dart';
import 'package:stash/loader.dart';
import 'package:path_provider/path_provider.dart';
import 'package:path/path.dart' as path;


class DB {
  int hnd = 0;

  static Future<String> getConfigDir() async {
    Directory configDir;

 if (Platform.isWindows) {
    configDir = await getApplicationSupportDirectory();
  } else if (Platform.isMacOS) {
    configDir = await getApplicationDocumentsDirectory();
  } else if (Platform.isLinux) {
    configDir = Directory('${Platform.environment['HOME']}/.config');
  } else {
    throw UnsupportedError("This OS is not supported.");
  }
    return configDir.path;
  }

  DB(String dbPath) {
    var fun = stashLibrary!.lookupFunction<ArgsS, ArgsS>('stash_openDB');
    var res = fun(dbPath.toNativeUtf8());
    hnd = res.hnd;
  }

  static Future<DB> defaultDB() async {
    var configDir = await getConfigDir();
    var dbPath = path.join(configDir, 'stash.db');
    return DB(dbPath);
  }

  void close() {
    var fun = stashLibrary!.lookupFunction<ArgsI, Argsi>('stash_closeDB');
    fun(hnd);
  }
} 