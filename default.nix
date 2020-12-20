{ pkgs ? import ./pkgs.nix {} }:

pkgs.buildGo115Module {
  pname = "hedron";
  version = "0.0.1";
  src = builtins.path { path = ./.; name = "hedron"; };
  goPackagePath = "github.com/thmzlt/hedron";
  vendorSha256 = null;
}
