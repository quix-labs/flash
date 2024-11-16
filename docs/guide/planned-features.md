# Planned Features

The following features are planned for future implementation:

- ⏳ Support for conditional listens.

| Operator |      trigger      |    wal_logical    |
|:--------:|:-----------------:|:-----------------:|
|  equals  |         ✅         |         ✅         |
|   neq    |         ❌         |         ❌         |
|    lt    |         ❌         |         ❌         |
|   lte    |         ❌         |         ❌         |
|   gte    |         ❌         |         ❌         |
| not null |         ❌         |         ❌         |
| is null  | ⚠️ using eq + nil | ⚠️ using eq + nil |

- ⏳ Handling custom primary for fake insert/delete when change appears
- ⏳ Tests implementation
- ⬜ Remove client in favor of direct listener start
- ⬜ Support attaching/detaching new listener during runtime.
- ... any suggestions is welcome.
