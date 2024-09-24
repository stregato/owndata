import 'dart:convert';
import 'dart:ffi';

import 'package:ffi/ffi.dart';
import 'comm.dart';
import 'database.dart';
import 'db.dart';
import 'filesystem.dart';
import 'identity.dart';
import 'loader.dart';

class Config {
  int quota = 0;
  String description = '';

  fromJson(Map<String, dynamic> json) {
    quota = json['quota'];
    description = json['description'];
  }

  toJson() {
    return {
      'quota': quota,
      'description': description,
    };
  }
}

final emptyConfig = Config();

typedef Groups = Map<String, Set<String>>;

class Safe {
  int hnd = 0;

  static int grant = 0;
  static int revoke = 1;
  static int curse = 2;
  static int endorse = 3;

  Safe.create(DB db, Identity identity, String url, {Config? config}) {
    config ??= emptyConfig;
    var fun = stashLibrary!.lookupFunction<ArgsISSS, ArgsiSSS>('stash_createSafe');
    var res = fun(db.hnd, jsonEncode(identity.toJson()).toNativeUtf8(),
        url.toNativeUtf8(), jsonEncode(config.toJson()).toNativeUtf8());
    hnd = res.handle;
  }

  Safe.open(DB db, Identity identity, String url) {
    var fun = stashLibrary!.lookupFunction<ArgsISS, ArgsiSS>('stash_openSafe');
    var res = fun(db.hnd, jsonEncode(identity.toJson()).toNativeUtf8(),
        url.toNativeUtf8());
    hnd = res.handle;
  }

  void close() {
    var fun = stashLibrary!.lookupFunction<ArgsI, Argsi>('stash_closeSafe');
    fun(hnd);
  }

  Groups updateGroup(String groupName, int change, List<String> users) {
    var fun =
        stashLibrary!.lookupFunction<ArgsISIS, ArgsiSiS>('stash_updateGroup');
    var res = fun(hnd, groupName.toNativeUtf8(), change,
        jsonEncode(users).toNativeUtf8());
    return res.map
        .map((k, v) => MapEntry(k, (v as Map<String, dynamic>).keys.toSet()));
  }

  Groups getGroups() {
    var fun = stashLibrary!.lookupFunction<ArgsI, Argsi>('stash_getGroups');
    var res = fun(hnd);
    return res
        .map
        .map((k, v) => MapEntry(k, (v as Map<String, dynamic>).keys.toSet()));
  }

  List<String> getKeys(String groupName) {
    var fun = stashLibrary!.lookupFunction<ArgsIS, ArgsiS>('stash_getKeys');
    var res = fun(hnd, groupName.toNativeUtf8());
    return List<String>.from(res.list);
  }

  Filesystem openFS() {
    var fun = stashLibrary!.lookupFunction<ArgsI, Argsi>('stash_openFS');
    var res = fun(hnd);
    return Filesystem(res.handle);
  }

  Database openDatabase(String groupName, Map<String, String> ddls) {
    var fun = stashLibrary!.lookupFunction<ArgsISS, ArgsiSS>('stash_openDatabase');
    var s = jsonEncode(ddls);
    var res = fun(hnd, groupName.toNativeUtf8(), s.toNativeUtf8());
    return Database(res.handle);
  }

  Comm openComm() {
    var fun = stashLibrary!.lookupFunction<ArgsI, Argsi>('stash_openComm');
    var res = fun(hnd);
    return Comm(res.handle);
  }
}
