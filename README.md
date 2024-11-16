# Flash

[![Documentation](https://img.shields.io/github/actions/workflow/status/quix-labs/flash/deploy_docs.yml?label=Documentation)](https://flash.quix-labs.com/guide)
[![License](https://img.shields.io/github/license/quix-labs/flash?color=blue)](https://github.com/quix-labs/flash/blob/main/LICENSE.md)

**Flash** is a lightweight Go library for managing real-time PostgreSQL changes using event management.

## Notes

**This library is currently under active development.**

Features and APIs may change.

Contributions and feedback are welcome!

## Features

- ✅ Start/Stop listening during runtime.
- ✅ Supports common PostgreSQL events: Insert, Update, Delete, Truncate.
- ✅ Driver interfaces for creating new drivers.
- ✅ Parallel Callback execution using goroutine
- ✅ Listen for changes in specific columns, not the entire row.
- ✅ Listen changes using WAL replication

## 🌐 Visit Our Website

For more information, updates, and resources, check out the official website:

- [Flash Official Website](https://flash.quix-labs.com)

## 📚 Documentation

Our detailed documentation is available to help you get started, learn how to configure and use Flash, and explore
advanced features:

- [Full Documentation](https://flash.quix-labs.com/guide)

## Contributing

1. Fork the repository.
2. Create a new branch for your feature or bugfix.
3. Commit your changes.
4. Push your branch.
5. Create a pull request.

## Credits

- [COLANT Alan](https://github.com/alancolant)
- [All Contributors](../../contributors)

## License

MIT. See the [License File](LICENSE.md) for more information.
