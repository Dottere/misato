# Misato 📚

A lightweight, self-hosted manga server built for the modern reader. 

Misato allows you to host your personal manga library and read it beautifully in your browser. It is designed to be fast, minimal, and fully controllable from your terminal.

## ✨ Features

* **Native CBZ Support:** Directly reads and serves your `.cbz` comic archives without extracting them to the disk.
* **Blazing Fast I/O:** Built with Go, featuring a thread-safe, in-memory ZIP cache to eliminate file reading bottlenecks.
* **Interactive CLI:** A built-in terminal interface for real-time server management, uptime tracking, and memory stats.
* **Zero Dependencies:** Compiles into a single, standalone binary with an embedded SQLite database.
* **Highly Configurable:** Easily manage ports, timeouts, and guest access through a simple `config.json` or CLI flags.

---

## 🚀 Getting Started

To easily compile and manage the project across different operating systems, we use [Task](https://taskfile.dev/) as our build tool. It acts as a modern, cross-platform alternative to traditional Makefiles.

### Prerequisites

1. **Go 1.22+** (Required for compiling the server and the advanced HTTP routing).
2. **Task** runner. Install it based on your OS:

**Windows (Winget)**
```powershell
winget install Task.Task
```

**Debian/Ubuntu**
```bash
sh -c "$(curl --location [https://taskfile.dev/install.sh](https://taskfile.dev/install.sh))" -- -d
```

## 🛠️ Compilation & Development

Clone the repository and navigate to the project folder. You can use the following commands to build or run the server:

### Run the application directly (Local Development)
```bash
task run
```

### Build an executable for your current OS
```bash
task build
```

### Cross-compile for a Linux server (amd64)
```bash
task build-linux
```

## ⚙️ Configuration

On its first run, Misato will automatically generate a config.json file in its root directory. You can edit this file to change the server port, library directory, timeouts, and authentication settings.

You can also override the configuration file using command-line flags:

```bash
  ./misato --port 8080 --verbose
```
