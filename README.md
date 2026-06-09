# Misato
Local manga hosting server that let's you browse your beloved manga in a modern way.

## Getting Started

To easily compile and manage the project across different operating systems, we use [Task](https://taskfile.dev/) as our build tool. It acts as a modern, cross-platform alternative to traditional Makefiles.

### Prerequisites

You will need to install the Task runner before compiling. 

**Windows**
```powershell
winget install Task.Task
```

**Debian/Ubuntu**
```bash
sh -c "$(curl --location [https://taskfile.dev/install.sh](https://taskfile.dev/install.sh))" -- -d
```

#### For other:
https://taskfile.dev/installation/

## Compilation

### Run the application directly for local development
```sh
task run
```

### Build an executable for your current operating system
```sh
task build
```

### Cross-compile for a Linux server (amd64)
```sh
task build-linux
```
