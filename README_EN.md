# imgrab

English | [中文](README.md)

imgrab is a Docker image pull CLI tool written in Go, providing image search, pull, save, and import functionality, supporting Docker Hub and private registries.

## Features

- Search Docker Hub images (TUI interface)
- Auto-import to Docker by default
- Download-only mode (save tar file)
- Multi-architecture support

## Installation

```bash
git clone https://github.com/llklkl/imgrab.git
cd imgrab
go build -o imgrab .
```

## Command Reference

### pull - Pull an Image

Pull a Docker image from a registry. **Automatically imports to Docker by default**, no manual action required.

```bash
./imgrab pull [image] [flags]
```

**Arguments:**
- `image`: Image name in format `[registry/]repository[:tag]`

**Flags:**
- `-d, --download-only`: Download only, do not import to Docker
- `-o, --output string`: Output directory (only with `--download-only`)
- `-a, --arch string`: Architecture (default: current architecture)

**Examples:**
```bash
# Pull and auto-import to Docker (default behavior)
./imgrab pull nginx

# Pull specific version and import
./imgrab pull nginx:1.25.3

# Download only, do not import
./imgrab pull nginx --download-only

# Download only to specific directory
./imgrab pull nginx -d -o ./images

# Specify architecture
./imgrab pull nginx -a arm64
```

**Notes:**
- In default mode, the image is downloaded to a temporary directory and automatically imported to Docker, then the temp files are cleaned up
- Use `--download-only` to keep the tar file in the current or specified directory

### search - Search Images (TUI Interface)

Search Docker Hub images using an interactive TUI interface. Supports visual selection of images, tags, architectures, and choice between download or import operations.

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

**TUI Workflow:**

1. **Search Images**
   - Enter keyword in search box, press `Enter` to start search
   - Use `↑`/`↓` to select an image, press `Enter` to view tags

2. **Select Version**
   - Use `↑`/`↓` to select a version tag
   - Press `Enter` to view available architectures

3. **Select Architecture**
   - Use `↑`/`↓` to select an architecture
   - Press `Enter` to enter confirmation screen

4. **Confirm Operation**
   - Use `←`/`→` to switch operation mode:
     - `Download Only` - Download image file only
     - `Import to Docker` - Download and import to Docker (default)
   - Press `y` or `Enter` to confirm and start download
   - Auto-exits after download/import completes

5. **Progress Display**
   - Real-time progress bar with speed and ETA
   - Auto-exits after completion

**Keyboard Shortcuts:**
- `Esc` - Go back to previous screen
- `q` / `Ctrl+C` - Exit application
