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

      in {
        devShells = {
          default = pkgs.devshell.mkShell {
            name = "terraform-provider-talos";

            packages = with pkgs; [
              go_1_18
              gopls
              qemu_full
              inotify-tools
              gcc
              terraform
            ];
            commands = [
              {
                name = "acctest";
                category = "testing";
                help = "Run acceptance tests.";
                command ="tools/pretest.sh";
              }
              {
                name ="tests";
                category = "testing";
                help = "Run unit tests.";
                command=''
                  go vet ./...
                  go run honnef.co/go/tools/cmd/staticcheck ./talos
                  go test -v -race -vet=off ./...
'';
              }
              {
                name = "generate";
                category = "code";
                help = "Regenerate documentation.";
                command ="go generate ./...";
              }
            ];
          };
        };
      });
}