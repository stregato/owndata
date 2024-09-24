import 'package:dstash/dstash.dart';
import 'package:test/test.dart';

void main() {
  group('A group of tests', () {
    setUp(() {
    loadstashLibrary();
    setLogLevel('debug');
    });

    test('First Test', () {
    var i = Identity('Admin');

    var db = DB.defaultDB();

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
  });
}

