{
  description = "A Finite State Machine Go library, kept simple.";

  inputs.nixpkgs.url = "github:NixOS/nixpkgs/nixos-21.11";

  outputs = { self, nixpkgs }: {

    devShell.x86_64-linux =
      with import nixpkgs { system = "x86_64-linux"; };
      mkShell {
        packages = [ go_1_17 ];
      };

  };
}
