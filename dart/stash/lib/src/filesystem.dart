import 'dart:convert';
import 'dart:ffi';
import 'dart:typed_data';

import 'package:ffi/ffi.dart';
import 'loader.dart';

class File {
  int id = 0;
  String dir = "";
  String name = "";
  bool isDir = false;
  String groupName = "";
  String creator = "";
  int size = 0;
  DateTime? modTime;
  Set<String> tags = {};
  Map<String, dynamic> attributes = {};
  String localCopy = "";
  DateTime? copyTime;
  String encryptionKey = "";

  File.fromJson(Map<String, dynamic> json) {
    id = json['id'];
    dir = json['dir'];
    name = json['name'];
    isDir = json['isDir'];
    groupName = json['groupName'];
    creator = json['creator'];
    size = json['size'];
    modTime = DateTime.parse(json['modTime']);
    tags = json['tags'].keys.toSet();
    attributes = json['attributes'];
    localCopy = json['localCopy'];
    copyTime = DateTime.parse(json['copyTime']);
    encryptionKey = json['encryptionKey'];
  }
}

class ListOptions {
  DateTime? after;
  DateTime? before;
  String orderBy = "";
  bool reverse = false;
  int limit = 0;
  int offset = 0;
  String prefix = "";
  String suffix = "";
  String tag = "";

  ListOptions({
    this.after,
    this.before,
    this.orderBy = "",
    this.reverse = false,
    this.limit = 0,
    this.offset = 0,
    this.prefix = "",
    this.suffix = "",
    this.tag = "",
  });

  toJson() {
    return {
      'after': after?.toUtc().toIso8601String(),
      'before': before?.toUtc().toIso8601String(),
      'orderBy': orderBy,
      'reverse': reverse,
      'limit': limit,
      'offset': offset,
      'prefix': prefix,
      'suffix': suffix,
      'tag': tag,
    };
  }
}

/*
type PutOptions struct {
	ID         FileID           // the ID of the file, used to overwrite an existing file
	Async      bool             // put the file asynchronously
	DeleteSrc  bool             // delete the source file after a successful put
	GroupName  safe.GroupName   // the group name of the file. If empty, the group name is calculated from the directory
	Tags       core.Set[string] // the tags of the file
	Attributes map[string]any   // the attributes of the file
}
*/

class PutOptions {
  int id = 0;
  bool async = false;
  bool deleteSrc = false;
  String groupName = "";
  List<String> tags = [];
  Map<String, dynamic> attributes = {};

  toJson() {
    return {
      'id': id,
      'async': async,
      'deleteSrc': deleteSrc,
      'groupName': groupName,
      'tags': tags,
      'attributes': attributes,
    };
  }
}

class GetOptions {
  bool async = false;

  toJson() {
    return {
      'async': async,
    };
  }
}


class Filesystem {
  int hnd = 0;

  Filesystem(this.hnd);

  void close() {
    var fun = stashLibrary!.lookupFunction<ArgsI, Argsi>('stash_closeFS');
    fun(hnd);
  }

  List<File> list(String path, ListOptions options) {
    var fun = stashLibrary!.lookupFunction<ArgsISS, ArgsiSS>('stash_list');
    var s = jsonEncode(options.toJson());
    var res = fun(
        hnd, path.toNativeUtf8(), s.toNativeUtf8());
    return List<File>.from(
        res.list.map((e) => File.fromJson(e)));
  }

  File stat(String path) {
    var fun = stashLibrary!.lookupFunction<ArgsIS, ArgsiS>('stash_stat');
    var res = fun(hnd, path.toNativeUtf8());
    return File.fromJson(res.map);
  }

  File putFile(String dest, String src, PutOptions options) {
    var fun = stashLibrary!.lookupFunction<ArgsISSS, ArgsiSSS>('stash_putFile');
    var res = fun(hnd, dest.toNativeUtf8(), src.toNativeUtf8(),
        jsonEncode(options).toNativeUtf8());
    return File.fromJson(res.map);
  }

  File putData(String dest, Uint8List src, PutOptions options) {
    var fun = stashLibrary!.lookupFunction<ArgsISDS, ArgsiSDS>('stash_putData');
    var res = fun(hnd, dest.toNativeUtf8(), CData.fromUint8List(src),
        jsonEncode(options).toNativeUtf8());
    return File.fromJson(res.map);
  }

  File getFile(String src, String dest, GetOptions options) {
    var fun = stashLibrary!.lookupFunction<ArgsISSS, ArgsiSSS>('stash_getFile');
    var res = fun(hnd, src.toNativeUtf8(), dest.toNativeUtf8(), 
        jsonEncode(options).toNativeUtf8());
    return File.fromJson(res.map);
  }

  Uint8List getData(String src, GetOptions options) {
    var fun = stashLibrary!.lookupFunction<ArgsISS, ArgsiSS>('stash_getData');
    var res = fun(hnd, src.toNativeUtf8(), jsonEncode(options).toNativeUtf8());
    return res.data;
  }

  void delete(String path) {
    var fun = stashLibrary!.lookupFunction<ArgsIS, ArgsiS>('stash_delete');
    fun(hnd, path.toNativeUtf8()).check();
  }

  File rename(String src, String dest) {
    var fun = stashLibrary!.lookupFunction<ArgsISS, ArgsiSS>('stash_rename');
    return File.fromJson(fun(hnd, src.toNativeUtf8(), dest.toNativeUtf8()).map);
  }
}
