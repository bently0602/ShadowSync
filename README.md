# ShadowSync

A WebDAV server with support for mirroring files and directories across multiple local directories. The server supports read operations from a primary directory and mirrors write operations (create, update, delete) across all specified directories.

## Features
- **WebDAV Support**: Fully functional WebDAV server for file management over HTTP.
- **Multi-Filesystem Mirroring**: Ensures file changes are mirrored across multiple directories.
- **Rollback on Write Failures**: If an operation fails on any directory, all previous changes are rolled back.
- **Basic Authentication**: Protect the server with a username and password.
- **Configurable via Command-Line Flags**: Easily specify directories, authentication credentials, and the server port.

---

## Installation

### Prerequisites
- Go 1.18+ installed on your machine.

### Clone the Repository
```bash
git clone <repository_url>
cd <repository_directory>
```

### Build
```bash
go build -o webdav-server .
```

---

## Usage

Run the server using the following command:

```bash
./shadowsync --port <PORT> --dirs <DIRS> --username <USERNAME> --password <PASSWORD>
```

### Flags
- `--port`: Port number on which the server will run (default: `8080`).
- `--dirs`: Comma-separated list of directories to use for mirroring (e.g., `./primary,./mirror1,./mirror2`). The first directory is used as the **primary directory** for read operations.
- `--username`: Username for Basic Authentication (default: `admin`).
- `--password`: Password for Basic Authentication (default: `password`).

### Example
```bash
./shadowsync --port 9090 --dirs ./primary,./mirror1,./mirror2 --username myuser --password mypass
```

- The server will:
  - Serve files from `./primary` for read operations.
  - Mirror write operations to `./primary`, `./mirror1`, and `./mirror2`.
  - Protect access with Basic Authentication using `myuser` and `mypass`.

---

## How It Works

### Primary Directory
- The first directory specified in the `--dirs` flag is the **primary directory**.
- All **read operations** (e.g., file listing, opening files) are performed on the primary directory.

### Mirroring
- All **write operations** (e.g., creating, updating, deleting files) are mirrored across all specified directories.

### Rollback
- If a write operation fails on any directory, changes made on previously successful directories are rolled back to maintain consistency.

---

## Authentication

The server uses Basic Authentication to restrict access. Credentials are provided via the `--username` and `--password` flags. Clients must supply these credentials to access the server.

---

## Example Use Cases

### File Redundancy
Keep multiple directories in sync, ensuring that any changes to files are mirrored for redundancy.

### File Distribution
Easily set up multiple locations where the same files are stored, useful for systems that require synchronized copies.

---

## Development

### File Structure
- `main.go`: Entry point of the application. Handles server setup and configuration.
- `operations.go`: Defines file operations (e.g., create, delete, rename) and their rollback logic.
- `multifs.go`: Implements the `MultiFS` structure for managing multiple filesystems.
- `multifile.go`: Implements the `multiFile` structure for managing file-level operations across filesystems.

### Running the Project
```bash
go run main.go multifs.go operations.go multifile.go --port 8080 --dirs ./primary,./mirror1 --username admin --password secret
```

---

## Security Considerations

- Ensure that the directories specified in `--dirs` are secure and accessible only by trusted users.
- Use strong credentials for Basic Authentication.
- If exposing the server over the internet, consider using HTTPS to protect authentication credentials.

---

## Contributing

Contributions are welcome! Feel free to submit issues or pull requests to improve this project.
