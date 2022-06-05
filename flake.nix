{
  description = "A flake for setting up a build environment for terraform-provider-talos";
  inputs.nixpkgs.url = github:NixOS/nixpkgs/nixos-22.05;
  outputs = { self, nixpkgs }:
  let
    version = builtins.substring 0 8 self.lastModifiedDate
    supportedSystems = [ "x86_64-linux" ];
    forAllSystems = nixpkgs.lib.genAttrs supportedSystems;
    nixpkgsFor = forAllSystems (system: import nixpkgs { inherit system; });
  in {
    packages =
      let
        pkgs = nixpkgsFor.x86_64-linux
      in {
        default = pkgs.buildGoModule {
	  pname = "terraform-provider-talos";
	  inherit version;
	  src = ./.;
	  vendorSha256 = pkgs.lib.fakeSha256;
	}
      };
  };
}