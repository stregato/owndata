import 'dart:ffi';

import 'package:ffi/ffi.dart';
import 'loader.dart';
 

class Identity {
  String id = '';
  String private = '';

  Identity(String nick) {
    var fun = stashLibrary!.lookupFunction<ArgsS, ArgsS>('stash_newIdentity');
    var m = fun(nick.toNativeUtf8()).map;
    id = m['i'];
    private = m['p'];
  }

  String nick() {
    var idx = id.lastIndexOf('.');
    if (idx > 0) {
      return id.substring(0, idx);
    }
    return '';
  }

  @override
  String toString() {
    return id;
  }

  toJson() {
    return {
      'i': id,
      'p': private,
    };
  }
}
