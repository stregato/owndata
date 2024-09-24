import 'dart:convert';
import 'dart:typed_data';

import 'package:flutter_test/flutter_test.dart';
import 'package:stash/comm.dart';

import 'package:stash/db.dart';
import 'package:stash/filesystem.dart';
import 'package:stash/identity.dart';
import 'package:stash/loader.dart';
import 'package:stash/safe.dart';

void main() {
  setUpAll(() async {
    TestWidgetsFlutterBinding.ensureInitialized();
    loadstashLibrary();
    setLogLevel('debug');
  });
  test('create safe', () async {
    var i = Identity('Admin');

    var db = await DB.defaultDB();

    var url = 'file:///tmp/${i.id}/sample';
    var s = Safe.create(db, i, url);

    var groups = s.getGroups();
    expect(groups, isNotNull);

    var alice = Identity('Alice');
    groups = s.updateGroup('usr', Safe.grant, [alice.id]);
    expect(groups['usr']?.contains(alice.id), true);

    groups = s.getGroups();
    expect(groups['usr']?.contains(alice.id), true);

    var keys = s.getKeys('usr');
    expect(keys, isNotNull);

    s.close();
    db.close();
  });

  test('filesystem', () async {
    var i = Identity('Admin');

    var db = await DB.defaultDB();

    var url = 'file:///tmp/${i.id}/sample';
    var s = Safe.create(db, i, url);

    var fs = s.openFS();
    expect(fs, isNotNull);

    var files = fs.list('', ListOptions());
    expect(files, isNotNull);

    var f = fs.putData("data.sample", Uint8List.fromList(utf8.encode("Hello, World!")), PutOptions());
    expect(f, isNotNull);
    expect(f.name, "data.sample");
    expect(f.creator, i.id);

    f = fs.stat("data.sample");
    expect(f, isNotNull);
    expect(f.name, "data.sample");
    expect(f.creator, i.id);

    var data = fs.getData("data.sample", GetOptions());
    expect(data, isNotNull);
    expect(utf8.decode(data), "Hello, World!");

    f = fs.rename("data.sample", "data2.sample");
    expect(f, isNotNull);
    expect(f.name, "data2.sample");

    fs.delete("data2.sample");

    fs.close();
    s.close();
    db.close();
  });

  test('database', () async {
    var i = Identity('Admin');

    var db = await DB.defaultDB();

    var url = 'file:///tmp/${i.id}/sample';
    var s = Safe.create(db, i, url);

    var database = s.openDatabase('usr', {
      '1.0': 
''' -- INIT
CREATE TABLE IF NOT EXISTS names (id INTEGER PRIMARY KEY, name TEXT);'''   });

    var cnt = database.exec('INSERT INTO names (name) VALUES (:name);', {'name': 'Alice'});
    expect(cnt, 1);

    var rows = database.query('SELECT * FROM names;', {});
    expect(rows, isNotNull);

    var row = rows.next();
    expect(row, isNotNull);
    expect(row[0], 1);
    expect(row[1], 'Alice');

    database.close();
    s.close();
    db.close();

  });

  test('comm', () async {
    var i = Identity('Admin');

    var db = await DB.defaultDB();

    var url = 'file:///tmp/${i.id}/sample';
    var s = Safe.create(db, i, url);

    var alice = Identity('Alice');
    s.updateGroup('usr', Safe.grant, [alice.id]);

    var comm = s.openComm();

    var m = Message(recipient: alice.id, text: 'Hello, Alice!');
    comm.send(alice.id, m);

    comm.broadcast('usr', m);

    var messages = comm.receive();
    expect(messages, isNotNull);
    expect(messages[0].text, 'Hello, Alice!');

    s.close();
    db.close();
  });
}
