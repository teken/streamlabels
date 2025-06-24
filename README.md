# Streamlabels

A simple Go program to generate text files containing Twitch stream information like newest followers, subscribers, and bits leaderboard. These files can then be used as sources in OBS Studio or other streaming software.

## Features

- **Newest Follower Tracking**: Get the most recent follower for your channel
- **Newest Subscriber Tracking**: Get the most recent subscriber for your channel  
- **Bits Leaderboard**: Get a leaderboard of top bits contributors (top 10)
- **Automatic Authentication**: Uses Twitch OAuth device flow for easy setup
- **Real-time Updates**: Configurable refresh intervals for live data

## Prerequisites

- Go 1.24.3 or later (for building from source)
- A Twitch account
- Proper Twitch API permissions (moderator:read:followers, channel:read:subscriptions)

## Installation

### Option 1: Download Pre-built Binary
1. Download the latest release from the releases page
2. Make the binary executable: `chmod +x streamlabels`

### Option 2: Build from Source
```bash
git clone <repository-url>
cd streamlabels
go build -o streamlabels main.go
```

## Configuration

The program uses a built-in Twitch client ID and doesn't require a config file. Authentication is handled automatically through Twitch's OAuth device flow.

## Usage

### Basic Usage
```bash
./streamlabels <channel_name> [flags]
```

### Examples

**Track newest follower:**
```bash
./streamlabels your_channel_name --newest-follower
```

**Track newest subscriber:**
```bash
./streamlabels your_channel_name --newest-subscriber
```

**Track bits leaderboard:**
```bash
./streamlabels your_channel_name --bits-leaderboard
```

**Track multiple data types:**
```bash
./streamlabels your_channel_name --newest-follower --newest-subscriber --bits-leaderboard
```

**Custom output directory:**
```bash
./streamlabels your_channel_name --newest-follower --output /path/to/output/directory
```

**Custom refresh interval:**
```bash
./streamlabels your_channel_name --newest-follower --refresh-interval 5s
```

### Command Line Flags

| Flag | Description | Default |
|------|-------------|---------|
| `--newest-follower` | Subscribe to newest follower updates | false |
| `--newest-subscriber` | Subscribe to newest subscriber updates | false |
| `--bits-leaderboard` | Subscribe to bits leaderboard updates | false |
| `--refresh-interval` | How often to refresh data | 1s |
| `--output` | Output directory for text files | Current directory |
| `--logout` | Clear stored authentication tokens | false |
| `--help` | Show help information | false |

### Output Files

The program generates the following text files in the specified output directory:

- `newest_followers.txt` - Contains the username of the most recent follower
- `newest_subscriber.txt` - Contains the username of the most recent subscriber  
- `bits_leaderboard.txt` - Contains a formatted list of top bits contributors

### Authentication

On first run, the program will display a device code and authorization URL. Follow these steps:

1. Visit the provided authorization URL
2. Enter the device code shown in the terminal
3. Authorize the application on Twitch
4. The program will automatically store your authentication tokens securely

To clear stored authentication tokens, use the `--logout` flag.

## Integration with OBS Studio

1. Run streamlabels with your desired options
2. In OBS Studio, add a "Text (GDI+)" or "Text (FreeType 2)" source
3. Check "Read from file" and browse to the generated text file
4. The text will update automatically as new data comes in

## Development

### Building
```bash
go build -o streamlabels main.go
```

### Dependencies
- `github.com/nicklaw5/helix/v2` - Twitch API client - currently using a forked one in my github, will go back to nicks once i put together a pr and get it merged
- `github.com/zalando/go-keyring` - Secure token storage

## License

This project is licensed under GNU GPL v3. See the [LICENSE](LICENSE) file for details.
