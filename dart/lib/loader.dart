import 'dart:convert';
import 'dart:ffi';
import 'dart:io';
import 'dart:typed_data';

import 'package:ffi/ffi.dart';


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

typedef ArgsISDS = CResult Function(
    Int64, Pointer<Utf8>, CData, Pointer<Utf8>);
typedef ArgsiSDS = CResult Function(
    int, Pointer<Utf8>, CData, Pointer<Utf8>);


DynamicLibrary? mioLibrary;

void loadMioLibrary() {
  if (mioLibrary != null) {
    return;
  }

  List<String> libraryPaths = [];

  switch (Platform.operatingSystem) {
    case 'linux':
      libraryPaths.add('libmio.so');
      libraryPaths.add('lib/libmio.so');
      break;
    case 'android':
      libraryPaths.add('libmio.so');
      break;
    case 'macos':
      libraryPaths.add('libmiod.dylib');
      break;
    case 'ios':
      mioLibrary = DynamicLibrary.process();
      freeC = mioLibrary!.lookupFunction<FreeC, FreeCDart>('free');
      return;
    case 'windows':
      libraryPaths.add('miod.dll');
      break;
    default:
      throw Exception('Unsupported platform');
  }
  for (String libraryPath in libraryPaths) {
    try {
      mioLibrary = DynamicLibrary.open(libraryPath);
      freeC = mioLibrary!.lookupFunction<FreeC, FreeCDart>('free');
      return;
    } catch (e) {
      // ignore
    }
  }
  throw Exception('Failed to load Mio library');
}

void setLogLevel(String level) {
  var fun = mioLibrary!.lookupFunction<ArgsS, ArgsS>('mio_setLogLevel');
  var r = fun(level.toNativeUtf8());
  r.check();
}