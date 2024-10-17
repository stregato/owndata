# Welcome

Stash is a library for encrypted data storage and exchange. Similar to traditional storage, Stash offers controlled access with a security model inspired by Unix permissions.
Unlike traditional storage, the access is based on cryptographic features, enabling distributed control.
Stash follows the principle of Encrypt Remote where data stored on local devices is not necessarily protected while data on remote devices, especially cloud, must be protected by encryption.

The library is written in Go and is available on Windows, Linux, MacOS, Android and iOS.
Bindings are available for Python, Java and Dart.

## Hi Stack
The following Python code demonstrates how the library facilitates secure and encrypted communication.

```py
from pstack import *

# create public/private keys
alice = Identity('Alice')
bob = Identity('Bob')

# create a safe given a s3 storage. Creator is alice.
url = 's3:///..../'+alice.id+'/sample'
s = Safe.create(DB.default(), alice, url)

# grant access to Bob
s.update_group('usr', Safe.grant, [bob.id]) 

# send a message to all members of group 'usr', i.e. Bob
c = s.comm()
c.broadcast('usr', text='hello')
```

The library supports various data paradigms, including messaging, distributed SQL, and file storage.

Details about the design and implementation are available in the [manual](./MANUAL.md)

# Installation
The library is implemented in Go and is available as a binary distribution for Windows, Linux, macOS, iOS, and Android at https://github.com/stregato/stash/releases.

Java bindings are provided as a JAR library included in the release. Maven support will be added in the future. Additionally, the release includes a command-line tool, `stash`, which offers a basic client for the library.

Python bindings can be installed using:
```sh
pip install pstack
```

Dart binding can be installed using:
```sh
dart get dstack
```


# Build
In case the available binaries do no work on your system, you can build the library from the source code usind the available Makefile.
The Makefile contains different tasks for different architectures. In case you want to build for all architectures (default task) you need C cross compiling libraries. The Makefile has been tested on MacOS.


