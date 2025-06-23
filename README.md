# Streamlabels

A simple program to generate text files containing Twitch stream information like newest followers, subscribers, and bits leaderboard. These files can then be used as sources in OBS Studio or other streaming software.

## Installation

1. Download the latest release from the releases page
2. Create a config file in one of these locations:
   - `/etc/streamlabels/config.toml`
   - `$HOME/.config/streamlabels/config.toml` 
   - `config.toml` in the same directory as the executable

## Configuration

Create a config.toml file with the following contents:

```toml
client_id = "your_client_id"
client_secret = "your_client_secret"
login = "your_login"
```

## Usage

```bash
./streamlabels
```

## License

This project is licensed under GNU GPL v3.
