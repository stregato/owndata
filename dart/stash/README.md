dstash is a Dart binding for the Stash library. It enables encrypted data storage and exchange. Similar to traditional systems, Stash offers controlled access with a security model inspired by Unix permissions.
Unlike traditional storage, the access is based on cryptographic features, enabling distributed control.

## Features
Stash offers convenient  

It supports Android, iOS, macOS, Linux and Windows on the common architectures for each OS.

## Getting started
Run _flutter pub get dstash_ or _dart pub get dstash_ from your terminal.
Alternatevely add _dstash_ dependency in your pubspec.yaml file.

## Usage


```dart
    loadstashLibrary();
    
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
```

## Additional information

More information available on github [page](http://github.com/stregato/pstash)
