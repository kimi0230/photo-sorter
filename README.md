# Photo Sorter

A photo and video sorting tool developed in Go that automatically organizes media files based on capture time, device model, and geographic location.

**Currently only supports macOS**

## Features

- Automatic sorting by capture date (Create Date)
- Support for multiple media formats (JPG, JPEG, HEIC, PNG, MP4, MOV, GIF, WEBP, HEIF, HEVC, MKV, AVI, WMV, FLV, MPEG, MPG, M4V, BMP, and RAW formats: CR2, CR3, NEF, NRW, ARW, RAF, RW2, ORF, PEF, DNG, RWL, RAW)
- Automatic handling of filename conflicts
- Multi-threaded processing support
- Detailed processing logs
- Graceful shutdown support
- Geographic location tagging (Geo Tagging)
- Detailed processing statistics
- Command-line interface with flexible options
- Performance profiling support (CPU and memory)

## Requirements

- Go 1.23 or higher
- [exiftool](#exiftool)
- [spatialite](#spatialite-tools)
- [tag](#tag)

## Installation

### Using Makefile

```bash
make build
```

### Using Docker

```bash
docker build -t photo-sorter .
```

### spatialite-tools

```sh
brew install spatialite-tools

which spatialite
```

### exiftool
```sh
brew install exiftool
```

### tag
```sh
brew install tag
```

### gpmf-parser

https://github.com/gopro/gpmf-parser/tree/main

```sh
git submodule add https://github.com/gopro/gpmf-parser.git third_party/gpmf-parser
git submodule update --init --recursive

cd third_party/gpmf-parser
mkdir build
cd build
cmake ..
make
chmod +x gpmf-parser
```

## Usage

### Basic Usage

**It is recommended to disable Spotlight before running the program!!!**

```sh
# Disable Spotlight:
sudo mdutil -i off /

# Run with source directory
./photo-sorter -src /path/to/source/folder

# Run with source and destination directories
./photo-sorter -src /path/to/source/folder -dst /path/to/destination/folder

# Run with custom worker count
./photo-sorter -src /path/to/source/folder -workers 8

# Run with custom config file
./photo-sorter -src /path/to/source/folder -c /path/to/config.yaml

# Show version
./photo-sorter -version

# After completion, rebuild Spotlight index:
sudo mdutil -E /
sudo mdutil -i on /

```

### Command Line Options

| Option        | Type   | Default            | Description                                                         |
|---------------|--------|--------------------|---------------------------------------------------------------------|
| `-src`        | string | `.`                | Source photo folder path                                            |
| `-dst`        | string | `.`                | Destination folder for sorted files (if `.`, uses `{src_dir}_sort`) |
| `-workers`    | int    | `runtime.NumCPU()` | Maximum number of concurrent workers                                |
| `-c`          | string | `config.yaml`      | Configuration file path                                             |
| `-version`    | bool   | `false`            | Show version information                                            |
| `-cpuprofile` | string | ``                 | CPU profile file path (for performance profiling)                   |
| `-memprofile` | string | ``                 | Memory profile file path (for performance profiling)                |

### Configuration File

```yaml
version: "0.1.23"  # Version number

# Photo sorting tool configuration file

# Source photo folder path
src_dir: "source_media"

# Destination folder for sorted files (if empty, defaults to {src_dir}_sort)
dst_dir: ""

# Number of concurrent workers (default: 4)
workers: 4

# Dry run mode (only shows files that will be moved, does not actually execute)
dry_run: false

# Date format: YYYY-MM-DD (2006-01-02) or YYYY-MM (2006-01)
date_format: "2006-01"

# Enable geographic location tagging
enable_geo_tag: true

# Geo database file path (supports .geojson or .sqlite)
# Options: ./geodata/states.geojson or ./geodata/states.sqlite
geo_db_path: "./geodata/states.sqlite"

# Geocoder type: "geo_json" or "geo_spatialite"
geocoder_type: "geo_spatialite"

# Log level (debug, info, warn, error)
log_level: "info"

# Enable verification
enable_verify: true

# Supported file formats
formats:
  - ".jpg"
  - ".jpeg"
  - ".heic"
  - ".png"
  - ".mp4"
  - ".mov"
  - ".gif"
  - ".webp"
  - ".heif"
  - ".hevc"
  - ".mkv"
  - ".avi"
  - ".wmv"
  - ".flv"
  - ".mpeg"
  - ".mpg"
  - ".m4v"
  - ".bmp"
  - ".cr2"
  - ".cr3"
  - ".nef"
  - ".nrw"
  - ".arw"
  - ".raf"
  - ".rw2"
  - ".orf"
  - ".pef"
  - ".dng"
  - ".rwl"
  - ".raw"
  
# File types to ignore
ignore:
  - ".git"
  - ".gitignore"
  - ".go"
  - ".mod"
  - ".sum"
  - ".md"
  - ".log"
  - ".yaml"
  - ".sample"
  - ".DS_Store"
  - "Thumbs.db"

```

### Using Docker

```bash
docker run -v /path/to/photos:/app/input -v /path/to/output:/app/output photo-sorter -config config.yaml
```

## Output Structure

```
sorted_media/
├── 2024-08-02-Japan/
│   └── GoPro_HERO8_Black/
│       └── GH011629.MP4
├── 2024-06-01/
│   └── iPhone11/
│       └── IMG_1234.JPG
├── unknown_date/
│   └── unknown_device/
│       └── IMG_5678.JPG
└── unknown_format/
    └── document.pdf
```

## Error Handling

- Log files are recorded in logs/app.log
- Files without date information are categorized into the unknown_date folder
- Files without device information are categorized into the unknown_device folder
- Unsupported file formats are categorized into the unknown_format folder

## Processing Statistics

The program provides detailed processing statistics:
- Total number of files
- Number of successfully processed files
- Number of failed files
- Processing time
- Statistics of unsupported file formats
- Directory structure and file count statistics
