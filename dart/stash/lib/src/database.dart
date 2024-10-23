

import 'dart:convert';
import 'dart:ffi';

import 'package:ffi/ffi.dart';
import 'loader.dart';


class Update {
  String key = '';
  Map<String, dynamic> args = {};
  double version = 0.0;
  Update(this.key, this.args, this.version);

  Update.fromJson(Map<String, dynamic> json) {
    key = json['key'];
    args = json['args'];
    version = json['version'];
  }
}

class Database {
  int hnd;
  Database(this.hnd);

  void close() {
    var fun = stashLibrary!.lookupFunction<ArgsI, Argsi>('stash_closeDB');
    fun(hnd);
  }

  Transaction transaction() {
    var fun = stashLibrary!.lookupFunction<ArgsI, Argsi>('stash_transaction');
    return Transaction(fun(hnd).handle);
  }
  
  Rows query(String query, Map<String, dynamic> args) {
    var fun = stashLibrary!.lookupFunction<ArgsISS, ArgsiSS>('stash_query');
    var s = jsonEncode(args);
    return Rows(fun(hnd, query.toNativeUtf8(), s.toNativeUtf8()).handle);
  }

  List<Update> sync() {
    var fun = stashLibrary!.lookupFunction<ArgsI, Argsi>('stash_sync');
    return fun(hnd).list.map((x) => Update.fromJson(x)).toList();
  }

  void cancel() {
    var fun = stashLibrary!.lookupFunction<ArgsI, Argsi>('stash_cancel');
    fun(hnd).check();
  }
}

class Transaction {
  int hnd;
  Transaction(this.hnd);

  int exec(String query, Map<String, dynamic> args) {
    var fun = stashLibrary!.lookupFunction<ArgsISS, ArgsiSS>('stash_exec');
    var s = jsonEncode(args);
    return fun(hnd, query.toNativeUtf8(), s.toNativeUtf8()).integer;
  }

  void commit() {
    var fun = stashLibrary!.lookupFunction<ArgsI, Argsi>('stash_commit');
    fun(hnd).check();
  }

  void rollback() {
    var fun = stashLibrary!.lookupFunction<ArgsI, Argsi>('stash_rollback');
    fun(hnd).check();
  }

}

class Rows {
  int hnd;
  Rows(this.hnd);

  List<dynamic> next() {
    var fun = stashLibrary!.lookupFunction<ArgsI, Argsi>('stash_nextRow');
    return fun(hnd).list;
  }

  void close() {
    var fun = stashLibrary!.lookupFunction<ArgsI, Argsi>('stash_closeRows');
    fun(hnd);
  }
}