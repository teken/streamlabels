{
  description = "Twitch stream labels application";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = { self, nixpkgs, flake-utils }:
    flake-utils.lib.eachDefaultSystem (system:
      let
        pkgs = nixpkgs.legacyPackages.${system};
        
        streamlabels = pkgs.buildGoModule {
          pname = "streamlabels";
          version = "1.0.0";
          src = ./.;
          
          vendorHash = "sha256-eZQe8gstzj9BxjuZiXV1UFnuzjICJBFjiTpJNGFiyFI=";
          
          env.CGO_ENABLED = "0";
          ldflags = [ "-s" "-w" ];
          
          meta = with pkgs.lib; {
            description = "Twitch stream labels application for OBS and other streaming software";
            homepage = "https://github.com/teken/streamlabels";
            license = licenses.mit;
            platforms = platforms.all;
            mainProgram = "streamlabels";
          };
        };
      in
      {
        # Default package - the application for the current system
        packages.default = streamlabels;

        # App - allows running with 'nix run'
        apps.default = {
          type = "app";
          program = "${streamlabels}/bin/streamlabels";
          meta = with pkgs.lib; {
            mainProgram = "streamlabels";
            description = "A small go app to get basic stream labels for twitch";
            longDescription = ''
              A simple Go program to generate text files containing Twitch stream information like newest followers, subscribers, and bits leaderboard.
              These files can then be used as sources in OBS Studio or other streaming software.
            '';
            homepage = "https://github.com/teken/streamlabels";
            license = lib.licenses.gpl3Plus;
            maintainers = with lib.maintainers; [
              teken
            ];
          };
        };
      }
    );
}