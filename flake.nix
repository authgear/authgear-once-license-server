{
  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixpkgs-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs =
    {
      nixpkgs,
      flake-utils,
      ...
    }:
    flake-utils.lib.eachDefaultSystem (
      system:
      let
        pkgs = import nixpkgs {
          inherit system;
          # Build 1.24.4 ourselves.
          overlays = [
            (final: prev: {
              go = (
                prev.go.overrideAttrs {
                  version = "1.24.4";
                  src = prev.fetchurl {
                    url = "https://go.dev/dl/go1.24.4.src.tar.gz";
                    hash = "sha256-WoaoOjH5+oFJC4xUIKw4T9PZWj5x+6Zlx7P5XR3+8rQ=";
                  };
                }
              );
            })
          ];
        };
      in
      {
        devShells.default = pkgs.mkShell {
          packages = [
            pkgs.go
          ];
        };
      }
    );
}
