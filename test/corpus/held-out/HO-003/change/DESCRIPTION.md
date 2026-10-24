Rewrite the snapshot copy.

The copy adds the read bits for everyone and leaves the rest as it was.

The change's test copies a small tree and checks that every file and directory in the copy has the mode the snapshot
promises: files 0644, directories 0755.
