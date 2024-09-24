
import 'package:stash/stash.dart';

void main() async {
  var alice = Identity('alice');
  var db = await DB.defaultDB();
  
  var safe = Safe.create(db, alice, 'file:///tmp/${alice.id}/sample');
  safe.close();
}
