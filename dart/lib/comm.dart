import 'dart:convert';
import 'dart:ffi';
import 'dart:typed_data';

import 'package:ffi/ffi.dart';
import 'package:mio/loader.dart';

class Message {
  String sender = '';
  int encryptionId = 0;
  String recipient = '';
  int id = 0;
  String text = '';
  Uint8List? data;
  String file = '';
  Message(
      {this.sender = '',
      this.encryptionId = 0,
      this.recipient = '',
      this.id = 0,
      this.text = '',
      this.data,
      this.file = ''});

  Message.fromJson(Map<String, dynamic> json) {
    sender = json['sender'];
    encryptionId = json['encryptionId'];
    recipient = json['recipient'];
    id = json['id'];
    text = json['text'];
    data = json['data'] != null ? base64Decode(json['data']) : null;
    file = json['file'];
  }

  toJson() {
    return {
      'sender': sender,
      'encryptionId': encryptionId,
      'recipient': recipient,
      'id': id,
      'text': text,
      'data': data != null ? base64Encode(data!) : null,
      'file': file,
    };
  }
}

class Comm {
  int hnd;
  Comm(this.hnd);

  void send(String userId, Message message) {
    var fun = mioLibrary!.lookupFunction<ArgsISS, ArgsiSS>('mio_send');
    fun(hnd, userId.toNativeUtf8(), jsonEncode(message.toJson()).toNativeUtf8())
        .check();
  }

  void broadcast(String groupName, Message message) {
    var fun = mioLibrary!.lookupFunction<ArgsISS, ArgsiSS>('mio_broadcast');
    fun(hnd, groupName.toNativeUtf8(), jsonEncode(message.toJson()).toNativeUtf8())
        .check();
  }

  List<Message> receive({String filter = ''}) {
    var fun = mioLibrary!.lookupFunction<ArgsIS, ArgsiS>('mio_receive');
    var res = fun(hnd, filter.toNativeUtf8());
    return res.list.map((x) => Message.fromJson(x)).toList();
  }

  void download(String message, String dest) {
    var fun = mioLibrary!.lookupFunction<ArgsISS, ArgsiSS>('mio_download');
    fun(hnd, message.toNativeUtf8(), dest.toNativeUtf8()).check();
  }

  void rewind(String dest, int messageId) {
    var fun = mioLibrary!.lookupFunction<ArgsISI, ArgsiSi>('mio_rewind');
    fun(hnd, dest.toNativeUtf8(), messageId);
  }
}
