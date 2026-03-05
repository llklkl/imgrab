# imgrab

English | [中文](README.md)

imgrab is a Docker image pull CLI tool written in Go, providing image search, pull, save, and import functionality, supporting Docker Hub and private registries.

## Installation

```bash
git clone https://github.com/llklkl/imgrab.git
cd imgrab
go build -o imgrab .
```

## Command Reference

### pull - Pull an Image

Pull a Docker image from a registry and save it as a tar file.

```bash
./imgrab pull [image] [flags]
```

**Arguments:**
- `image`: Image name in format `[registry/]repository[:tag]`

**Flags:**
- `-o, --output string`: Output directory
- `-a, --arch string`: Architecture (default: current architecture)
- `-i, --import`: Import to Docker after pulling

**Examples:**
```bash
# Pull nginx latest
./imgrab pull nginx

# Pull specific version
./imgrab pull nginx:1.25.3

# Pull and save to specific directory
./imgrab pull nginx -o ./images

# Specify architecture
./imgrab pull nginx -a arm64

# Import to Docker after pulling
./imgrab pull nginx -i
```

### search - Search Images (TUI Interface)

Search Docker Hub images using an interactive TUI interface.

```bash
./imgrab search [query]
```

**Arguments:**
- `query`: Optional, initial search keyword

**Examples:**
```bash
# Open search interface
./imgrab search

# Search for nginx directly
./imgrab search nginx
```

**TUI Controls:**
- Enter keyword in search box, press `Enter` to start search
- Use arrow keys `↑`/`↓` to select image, press `Enter` to view tags
- Select tag in tag list, press `Enter` to confirm download
- Select architecture in confirmation dialog (←/→ to switch), press `y`/`Enter` to confirm download
- Press `Esc` to go back, press `q`/`Ctrl+C` to quit

### login - Login to Registry

Login to Docker Hub or a private registry.

```bash
./imgrab login [flags]
```

**Flags:**
- `-u, --username string`: Username
- `-p, --password string`: Password
- `-r, --registry string`: Registry address (default: Docker Hub)

**Examples:**
```bash
# Login to Docker Hub
./imgrab login -u your_username -p your_password

# Login to private registry
./imgrab login -u your_username -p your_password -r registry.example.com
```

## Configuration

imgrab saves credentials in the user directory at `~/.docker/config.json`.
