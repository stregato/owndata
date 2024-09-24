import 'package:dstash/dstash.dart';

void main() async {
  var alice = Identity('alice');
  var db = DB.defaultDB();

  var s = Safe.create(db, alice, 'file:///tmp/${alice.id}/sample');

  var groups = s.getGroups();
  print(groups);
  var bob = Identity('Bob');
  groups = s.updateGroup('usr', Safe.grant, [bob.id]);
  groups = s.getGroups();

  var keys = s.getKeys('usr');
  print(keys);

  s.close();
}
