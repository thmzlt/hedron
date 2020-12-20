{ pkgsRef ? "a25913605b4bdc6dc224b60e9c9297d6af2b796b" }:

import (fetchTarball "https://github.com/nixos/nixpkgs/archive/${pkgsRef}.tar.gz") {}
