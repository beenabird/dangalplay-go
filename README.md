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
