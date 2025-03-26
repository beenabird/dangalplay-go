# DangalPlay Downloader

DangalPlay Downloader is a command-line tool for downloading movies and TV shows from DangalPlay. It fetches metadata, selects the best available streams, and downloads content using `N_m3u8DL-RE`.

## Features
- Download movies and TV shows from DangalPlay.
- Supports proxy configuration.
- Allows episode and season selection.
- Extracts metadata and organizes files properly.
- Uses `N_m3u8DL-RE` for HLS downloading and muxing.

## Prerequisites
- [Go](https://go.dev/) installed (for building the binary)
- [N_m3u8DL-RE](https://github.com/nilaoda/N_m3u8DL-RE) installed and accessible in `PATH`

## Installation

Clone the repository:
```sh
git clone https://github.com/beenabird/dangalplay-go.git
cd dangalplay-downloader
```

Build the binary:
```sh
go build -o dangalplay
```

## Usage

```sh
./dangalplay -url <content_url> [-o <output_dir>] [-p <proxy>] [-q <quality>] [-s <season>] [-e <episode>] [-w <episode_range>] [--debug]
```

### Parameters
| Parameter     | Description |
|--------------|-------------|
| `-url`       | Content URL (Required) |
| `-o`         | Output directory (Default: `downloads`) |
| `-p`         | Proxy server URL (Optional) |
| `-q`         | Video quality (Default: `720p`) |
| `-s`         | Season number (for TV Shows) |
| `-e`         | Episode number (for TV Shows) |
| `-w`         | Episode range (e.g., `S01E01-E05`) |
| `--debug`    | Enable debug mode |

## Example Commands

Download a movie:
```sh
./dangalplay -url https://www.dangalplay.com/movies/maa-ki-mamta
```

Download a TV show season:
```sh
./dangalplay -url https://www.dangalplay.com/shows/bandini -s 1
```

Download specific episodes:
```sh
./dangalplay -url https://www.dangalplay.com/shows/bandini -w S01E03-E05
```

## Configuration

The tool saves configuration in `config/dangalplay_config.json`. On first run, it asks for a `USER_ID` and optional proxy.

## License

This project is licensed under the MIT License.

## Disclaimer
This tool is for educational purposes only. Downloading copyrighted content without permission may violate terms of service and local laws. Use responsibly.

