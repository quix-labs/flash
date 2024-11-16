# Installation

To get started with Flash, follow the steps below to install the library and set it up in your Go project.

## Install in your own project

### Step 1: Install Flash

You can install Flash directly from GitHub using the following `go get` command. The `main` branch is currently used for
development, but the library is stable enough for most use cases:

```bash
go get -u github.com/quix-labs/flash@main
```

### Step 2: Install Additional Drivers (Optional)

Flash supports a variety of drivers for different configurations. For example, you can install
the [Trigger](./drivers/trigger/) driver with this command:

```bash
go get -u github.com/quix-labs/flash/drivers/trigger@main
```

You can explore other drivers in [Drivers Overview](./drivers/) page.

If you need a specific driver, just install it the same way.

### Step 3: Set up Your Main Package

After installing Flash and the necessary drivers, you can now start using it in your Go project.

Begin by creating your `main.go` file and importing the Flash library.

### Step 4: Run `go mod tidy` (Optional)

In case there are nested dependencies or missing packages, run the following command to tidy up your Go modules:

```bash
go mod tidy
```

This step ensures that all required dependencies are downloaded, and it also removes any unused dependencies.

## Troubleshooting

If you encounter any issues during installation, here are a few things to check:

- Make sure Go is installed correctly on your system. You can verify this by running:
    ```bash
    go version
    ```
- Ensure that your `$GOPATH` and `$GOROOT` are set correctly, especially if you're using a custom Go workspace.
- If the installation fails due to permissions, try running the `go get` command with elevated permissions (e.g., `sudo` on Linux or macOS):
    ```bash
    sudo go get -u github.com/quix-labs/flash@main
    ```

If you're still having trouble, feel free to open an issue on our [GitHub repository](https://github.com/quix-labs/flash/issues).


## Next Steps
Once you've successfully installed Flash, you're ready to start listening to PostgreSQL database events.

Check out the [usage guide](./start-listening) to dive deeper into setting up your listeners and configuring your events.