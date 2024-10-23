import 'dart:convert';
import 'dart:ffi';
import 'dart:io';

import 'package:ffi/ffi.dart';
import 'package:flutter/services.dart';

typedef FreeC = Void Function(Pointer<Uint8>);
typedef FreeCDart = void Function(Pointer<Uint8>);
late FreeCDart freeC;

sealed class CResult extends Struct {
  external Pointer<Uint8> ptr;
  @Uint64()
  external int len;
  @Int64()
  external int hnd;
  external Pointer<Utf8> err;

  int get handle {
    check();
    return hnd;
  }

  void check() {
    if (err.address != 0) {
      var e = CException(err.toDartString());
      freeC(err.cast());
      throw e;
    }
  }

  String ptrAsString() {
    final List<int> bytes = ptr.asTypedList(len);
    var res = String.fromCharCodes(bytes);
    freeC(ptr);
    return res;
  }

  String get string {
    check();
    if (ptr.address == 0) {
      return "";
    }

    return jsonDecode(ptrAsString()) as String;
  }

  int get integer {
    check();
    if (ptr.address == 0) {
      return 0;
    }

    return jsonDecode(ptrAsString()) as int;
  }

  Map<String, dynamic> get map {
    check();
    if (ptr.address == 0) {
      return {};
    }

    return jsonDecode(ptrAsString()) as Map<String, dynamic>;
  }

  List<dynamic> get list {
    check();
    if (ptr.address == 0) {
      return [];
    }

    var ls = jsonDecode(ptrAsString());

    return ls == null ? [] : ls as List<dynamic>;
  }

  Uint8List get data {
    final List<int> bytes = ptr.asTypedList(len);
    var res = Uint8List.fromList(bytes);
    freeC(ptr);
    return res;
  }
}

sealed class CData extends Struct {
  external Pointer<Uint8> ptr;
  @Uint64()
  external int len;

  static CData fromUint8List(Uint8List data) {
    final cDataPtr = calloc<CData>();

    // Allocate memory for the byte array
    final byteArray = calloc<Uint8>(data.length);

    // Copy the data from the Uint8List to the allocated memory
    for (int i = 0; i < data.length; i++) {
      byteArray[i] = data[i];
    }

    cDataPtr.ref.ptr = byteArray;
    cDataPtr.ref.len = data.length;

    // Return the CData instance
    return cDataPtr.ref;
  }
}

class CException implements Exception {
  String msg;
  CException(this.msg);

  @override
  String toString() {
    return msg;
  }
}

typedef Args = CResult Function();
typedef ArgsS = CResult Function(Pointer<Utf8>);

typedef ArgsI = CResult Function(Int64);
typedef Argsi = CResult Function(int);

typedef ArgsIS = CResult Function(Int64, Pointer<Utf8>);
typedef ArgsiS = CResult Function(int, Pointer<Utf8>);

typedef ArgsISS = CResult Function(Int64, Pointer<Utf8>, Pointer<Utf8>);
typedef ArgsiSS = CResult Function(int, Pointer<Utf8>, Pointer<Utf8>);

typedef ArgsISI = CResult Function(Int64, Pointer<Utf8>, Int64);
typedef ArgsiSi = CResult Function(int, Pointer<Utf8>, int);

typedef ArgsISIS = CResult Function(Int64, Pointer<Utf8>, Int64, Pointer<Utf8>);
typedef ArgsiSiS = CResult Function(int, Pointer<Utf8>, int, Pointer<Utf8>);

typedef ArgsISSS = CResult Function(
    Int64, Pointer<Utf8>, Pointer<Utf8>, Pointer<Utf8>);
typedef ArgsiSSS = CResult Function(
    int, Pointer<Utf8>, Pointer<Utf8>, Pointer<Utf8>);

typedef ArgsISDS = CResult Function(Int64, Pointer<Utf8>, CData, Pointer<Utf8>);
typedef ArgsiSDS = CResult Function(int, Pointer<Utf8>, CData, Pointer<Utf8>);

DynamicLibrary? stashLibrary;

Future<void> loadStashLibrary() async {
  if (stashLibrary != null) {
    return;
  }


  // Special handling for iOS since it doesn't use DynamicLibrary.open
  if (Platform.isIOS) {
    stashLibrary = DynamicLibrary.process();
    freeC = stashLibrary!.lookupFunction<FreeC, FreeCDart>('free');
    return;
  }

  var libraryPath = await _getLibraryPath();
  if (libraryPath != null) {
    // Load the dynamic library from the library path
    stashLibrary = DynamicLibrary.open(libraryPath.path);
    freeC = stashLibrary!.lookupFunction<FreeC, FreeCDart>('free');
    return;
  }
  
  throw Exception('Failed to find Stash library at $libraryPath');
}

Future<File?> _getLibraryPath() async {
  String assetPath = _getAssetPath();

  String libraryPath = 'lib/assets/$assetPath';
  File libraryFile = File(libraryPath).absolute;
  if (libraryFile.existsSync()) {
    return libraryFile;  
  }

  final byteData = await rootBundle.load('packages/dstash/assets/$assetPath');
  final tempDir = Directory.systemTemp;
  final tempFile = File('${tempDir.path}/$assetPath');
  tempFile.createSync(recursive: true);
  tempFile.writeAsBytesSync(byteData.buffer.asUint8List());

  return tempFile;
}

String _getAssetPath() {

  // Get the operating system and architecture to construct the folder name
  var os = Platform.operatingSystem; // linux, macos, windows, android, ios
  final arch = Platform.version.toLowerCase();

  String archFolder;

  if (os == 'macos') {
    os = 'darwin';
  }

  if (arch.contains('amd64') || arch.contains('x64')) {
    archFolder = 'amd64';
  } else if (arch.contains('arm64')) {
    archFolder = 'arm64';
  } else {
    throw Exception('Unsupported architecture: $arch');
  }

  // Compose the folder name using the operating system and architecture
  String osArchFolder = '${os}_$archFolder';

  // Compose the library name based on the operating system
  String libraryFileName;
  switch (os) {
    case 'linux':
    case 'android':
      libraryFileName = 'libstash.so';
      break;
    case 'darwin':
      libraryFileName = 'libstash.dylib';
      break;
    case 'windows':
      libraryFileName = 'stashd.dll';
      break;
    default:
      throw Exception('Unsupported platform: $os');
  }
  
  return '$osArchFolder/$libraryFileName';
}

void setLogLevel(String level) {
  var fun = stashLibrary!.lookupFunction<ArgsS, ArgsS>('stash_setLogLevel');
  var r = fun(level.toNativeUtf8());
  r.check();
}
