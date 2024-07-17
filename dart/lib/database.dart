

import 'dart:convert';
import 'dart:ffi';

import 'package:ffi/ffi.dart';
import 'package:mio/loader.dart';


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
    var fun = mioLibrary!.lookupFunction<ArgsI, Argsi>('mio_closeDB');
    fun(hnd);
  }

  int exec(String query, Map<String, dynamic> args) {
    var fun = mioLibrary!.lookupFunction<ArgsISS, ArgsiSS>('mio_exec');
    var s = jsonEncode(args);
    return fun(hnd, query.toNativeUtf8(), s.toNativeUtf8()).integer;
  }

  Rows query(String query, Map<String, dynamic> args) {
    var fun = mioLibrary!.lookupFunction<ArgsISS, ArgsiSS>('mio_query');
    var s = jsonEncode(args);
    return Rows(fun(hnd, query.toNativeUtf8(), s.toNativeUtf8()).handle);
  }

  List<Update> sync() {
    var fun = mioLibrary!.lookupFunction<ArgsI, Argsi>('mio_sync');
    return fun(hnd).list.map((x) => Update.fromJson(x)).toList();
  }

  void cancel() {
    var fun = mioLibrary!.lookupFunction<ArgsI, Argsi>('mio_cancel');
    fun(hnd).check();
  }
}

class Rows {
  int hnd;
  Rows(this.hnd);

  List<dynamic> next() {
    var fun = mioLibrary!.lookupFunction<ArgsI, Argsi>('mio_nextRow');
    return fun(hnd).list;
  }

  void close() {
    var fun = mioLibrary!.lookupFunction<ArgsI, Argsi>('mio_closeRows');
    fun(hnd);
  }
}