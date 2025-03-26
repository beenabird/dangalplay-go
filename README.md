# DangalPlay Downloader

Tool for downloading content from DangalPlay streaming platform

## Features

- Download movies and TV shows
- Season/episode selection
- Quality selection (360p, 480p, 720p, 1080p)
- Proxy support **(+)**
- Organized directory structure **(+)**
- Automatic metadata extraction (titles, years, seasons) **(+)**
- Episode range selection **(+)**

## Installation

1. Install dependencies:
   - [Go](https://golang.org/dl/) (1.16+)
   - [N_m3u8DL-RE](https://github.com/nilaoda/N_m3u8DL-RE) **(+ direct link)**
   - (Windows users) [VC Redist](https://aka.ms/vs/17/release/vc_redist.x64.exe) **(+)**

2. Build from source:
```bash
git clone https://github.com/beenabird/dangalplay-go  
cd dangalplay-go  
go build -o dangalplay

Configuration

First-Time Setup
Run the tool to initialize configuration:

./dangalplay
You'll be prompted for:

DangalPlay USER_ID (required)

Proxy configuration (optional)

Config File Location
config/dangalplay_config.json
Example structure:

json
Copy
{
  "user_id": "DP_1234567890",
  "proxy": "http://user:pass@proxy.example.com:8080"
}
Configuration Examples
Basic Configuration (USER_ID only):

json
Copy
{
  "user_id": "DP_ABCDE12345"
}
With Proxy:

json
Copy
{
  "user_id": "DP_ABCDE12345",
  "proxy": "socks5://localhost:9050"
}
Supported Proxy Formats:

HTTP: http://user:pass@host:port

SOCKS5: socks5://host:port

Authenticated: http://username:password@host:port

Managing Configuration

Edit config file:

nano config/dangalplay_config.json
Regenerate config:

bash
Copy
rm config/dangalplay_config.json && ./dangalplay
Configuration Precedence
Command-line arguments (highest priority)

Config file values

Default values

Example: Proxy in config is overridden by CLI flag:

bash
Copy
./dangalplay -url [URL] -p http://new.proxy:8080
Security Notes
🔒 Never share your config.json file

🔑 Proxy credentials are stored in plain text

🛑 Delete config to reset: rm config/dangalplay_config.json

Usage
bash
Copy
./dangalplay -url "CONTENT_URL" [OPTIONS]
Command Options
Flag	Description	Example
-url	Content URL (required)	-url "https://..."
-o	Output directory	-o ~/downloads
-q	Video quality	-q 1080
-s	Season number	-s 2
-e	Episode number (requires -s)	-s 2 -e 5
-w	Episode range	-w S02E05-E10
-p	Proxy URL	-p http://proxy:port
-debug	Enable debug logging	-debug
Examples
Download movie:

bash
Copy
./dangalplay -url "https://dangalplay.com/movies/movie-id" -q 1080
Download entire series:

bash
Copy
./dangalplay -url "https://dangalplay.com/shows/show-id"
Download specific season:

bash
Copy
./dangalplay -url [URL] -s 3
Download episode range:

bash
Copy
./dangalplay -url [URL] -w S02E05-E10
Directory Structure
Movies:

Copy
downloads/
└── Movies/
    └── Movie.Title.Year/
        └── Movie.Title.Year.720p.mp4
TV Shows:

Copy
downloads/
└── TV Shows/
    └── Series.Title/
        ├── Season 01/
        │   ├── Series.Title.S01E01.720p.mp4
        │   └── Series.Title.S01E02.720p.mp4
        └── Season 02/
            └── Series.Title.S02E01.720p.mp4
Legal Notice
Requires valid DangalPlay subscription

Only download content you have rights to access

Developer not responsible for misuse

Troubleshooting
Common Issues:

Enable debug mode: -debug

Verify N_m3u8DL-RE is in PATH

Check config file permissions

Ensure correct proxy configuration

Error Codes:

401: Invalid USER_ID (reconfigure with rm config/dangalplay_config.json)

403: Geo-restriction detected (use proxy)

500: Server error (retry later)

Support
Report issues or request features:
GitHub Issues

📄 License: MIT
⚠️ Disclaimer: Use at your own risk. May violate DangalPlay's Terms of Service.
