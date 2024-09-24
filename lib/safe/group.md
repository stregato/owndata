# Group Chains
Similar to common Unix operating systems, a group is a collection of users who share specific access rights within the system. For example, Linux has a group named _adm_ for users with administrative privileges.

In a trusted system, the current state of the group is maintained and managed by the system itself. 

In a distributed system, where peers may not fully trust each other, it is beneficial to keep a history of the changes. This allows the current state of the groups to be reconstructed at any time by replaying the changes from the history.

## The chain
A change in a group is defined by cryptographically signed block.

```python

class GroupChange:
    groupName       # the group that will be changed
    userId          # identifier of the user in the group
    change          # type of change
    signer          # the user that makes the change
    signature       # the ed25519 signature

```

For instance, the user _Alice_ can grant access to the user _Bob_ to a group _users_ with the following block

```python
GroupChange(121, 120, 'users', 'Bob', True, Alice, 0x39393...)
```

A block ID is a type of Snowflake ID, which ensures it is not only unique but also incrementally increasing.

### Why blocks are signed
Only members of the _adm_ group are granted the permissions to add or remove users from a group.

In this context, verifying the identity of the user who initiates a change is of utmost importance. This verification is typically achieved through a signature. The block incorporates both the user ID and the block's signature. The user ID includes a ed25519 public key, which is utilized to validate the signature.


### Why blocks are chained
Each block is linked to its predecessor by referencing the predecessor's id. Additionally, the signature of the block utilizes the hash of the previous block.

The chain enforces that counterfeit changes cannot be easily introduced.
As an example consider the following scenario:
1. Alice is an admin and it grants admin rights to Eve. That means that after the change Eve can grant rights to other users
2. Later Alice removes the rights to Eve and add the corresponding block to the chain.
3. Eve blames the change and create a new block to grant admin access to another user inserting it in the chain before her access was revoked.

The third action introduces a fork in the chain because Eve cannot give a valid successor to her block from the current chain. Eventually after the change there will be two blocks in the chain without a successor.
So enforcing a chain allows to detect counterfeit changes.

### Fork resolution
In the event of a fork, a peer that reads the chain must determine which branch is legitimate. There are multiple strategies to handle a fork in the chain, many of which are borrowed from blockchain literature. This design suggests two approaches:

1. The peer could opt for the branch where the majority of users have made changes. This strategy is effective in a network where peers frequently synchronize, enabling quick detection of a fork. It is particularly suitable for networks where synchronization is common and participants are likely to have a legitimate copy of the chain locally.

2. Alternatively, the peer could add a block that freezes the chain, with only the creator having the ability to unlock it and select the correct branch. This approach is more conservative and may require human intervention, but it introduces a potential single point of failure.

### Optional endorsment
Enhancing trust in the chain can be achieved by incorporating blocks that either endorse or challenge the current state. Whenever a peer synchronizes with the chain, it has the opportunity to add a block, thereby confirming the chain's validity if it aligns with their internal copy. This approach can aid in resolving forks, potentially by attributing greater significance to confirmations from administrators or long-standing users.

### First block
Initially, all groups, including the _adm_ group, are empty at the start of the chain. An exception is made for the first block, where the rules are relaxed for a special user known as _creator_. Typically, the _creator_ user adds themselves to the _adm_ group in the first block.


## The implementation
Description of the algorithms to maintain the change and derive the each group members.

## Synchorization
Synchronization is designed to align the public chain with the internal copy maintained by each peer. The underlying assumption is that interactions with the storage of the public chain incur a cost. Therefore, the algorithm strives to minimize this cost.

In the public storage, blocks are grouped into batches of 1024 items to limit the number of files. Simultaneously, the chain is distributed across multiple files, allowing peers to read only the files that have been modified. Files are named in sequential way.
The algorithm functions as follows:
1. A peer loads the chain from its local copy into memory, along with the current group states.
2. Assuming the local copy consists of 'n' blocks, these can be accommodated in 'm' = (1 + n / 1024) files. As changes typically occur in the most recent files, the peer fetches the 'm' file and any subsequent ones from the public storage.
3. If the 'm' file does not exist, it indicates that the public chain is compromised. In such a case, the synchronization halts and the process transitions to reconciliation.
4. The peer compares each block from the local copy with the corresponding one in the public storage. If there's a mismatch, the process fails and switches to reconciliation.
5. For each remaining block in the remote copies, the peer:
    - Verifies the validity of the signature.
    - If valid, applies the change to the group state.
    - If it's an enforce block, increments a related counter.
6. The peer calculates the optimal number of enforces based on the number of users. If the counter is less than this number, the peer adds an enforce block to the chain through the update process.

## Reconciliation
When a mismatch between the local copy and public chain exists, the chain is forked. 

## Update
