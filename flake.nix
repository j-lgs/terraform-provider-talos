{
  description =
    "A flake for setting up a build environment for terraform-provider-talos";

  inputs = {
    nixpkgs = { url = "github:NixOS/nixpkgs/nixos-22.05"; };
    flake-utils = { url = "github:numtide/flake-utils"; };
    devshell = { url = "github:numtide/devshell"; };
  };
  outputs = { self, nixpkgs, devshell, flake-utils }:
    flake-utils.lib.eachDefaultSystem (system:
      let
        pkgs = import nixpkgs {
          inherit system;
          overlays = [ devshell.overlay ];
        };
        tpt = pkgs.buildGo118Module {
          inherit system;
          pname = "terraform-provider-talos";
          version = "0.0.11";

          src = ./.;

          vendorSha256 = "sha256-/ttIyNegFRzWt8PvhNCgiw5tFOqUmHrJ02av8rofINY=";

          doCheck = false;

          meta = with pkgs.lib;
            {

            };
        };
      in {
        defaultPackage = tpt;
        devShells = {
          default = pkgs.devshell.mkShell {
            name = "terraform-provider-talos";

            packages = with pkgs; [
              go_1_18
              gopls
              gcc

              qemu_full
              inotify-tools

              terraform
            ];
            commands = [
              {
                name = "acctest";
                category = "testing";
                help = "Run acceptance tests.";
                command = "tools/pretest.sh";
              }
              {
                name = "build";
                category = "build";
                help = "Build the program for the current system";
                command = "nix build .#defaultPackage.${system}";
              }
              {
                name = "tests";
                category = "testing";
                help = "Run unit tests.";
                command = ''
                  go vet ./...
                  go run honnef.co/go/tools/cmd/staticcheck ./talos
                  go test -v -race -vet=off ./...
                '';
              }
              {
                name = "generate";
                category = "code";
                help = "Regenerate documentation.";
                command = "go generate ./...";
              }
            ];
          };
        };
      });
}
