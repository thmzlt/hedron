{ pkgs ? import ./pkgs.nix { } }:
let
  kubebuilder = pkgs.buildGo115Module rec {
    pname = "kubebuilder";
    version = "2.3.1";

    src = pkgs.fetchFromGitHub {
      owner = "kubernetes-sigs";
      repo = pname;
      rev = "v${version}";
      sha256 = "07frc9kl6rlrz2hjm72z9i12inn22jqykzzhlhf9mcr9fv21s3gk";
    };

    vendorSha256 = "079cnaflk2ap5fb357151fdqk7wr37dpghd3lmrmhcqwpfwp022m";

    subPackages = [ "./cmd" ];

    postInstall = ''
      mv $out/bin/cmd $out/bin/kubebuilder
    '';

    doCheck = false;
  };
in
(pkgs.callPackage ./default.nix { }).overrideAttrs (
  attrs: {
    src = null;

    nativeBuildInputs = [
      kubebuilder
      pkgs.go_1_15
      pkgs.k9s
      pkgs.kind
      pkgs.kubectl
      pkgs.kustomize
    ];

    shellHook = ''
      export GOPATH="$(pwd)/.go"
      export GOCACHE=""
      export GO111MODULE="on"
      export PATH="$PATH:$(pwd)/.go/bin"

      go mod init ${attrs.goPackagePath}

      set +v
    '';
  }
)
