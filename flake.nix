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
          
          CGO_ENABLED = "0";
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
        };

        # Overlay for easy installation
        overlays.default = final: prev: {
          streamlabels = streamlabels;
        };
      }
    );
}