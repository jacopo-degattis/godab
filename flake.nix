{
  description = "Go flake for ARM64 Darwin";

  inputs = {
    nixpkgs.url = "github:nixos/nixpkgs/nixos-unstable"; # Use the desired nixpkgs version
  };

  outputs = { self, nixpkgs, ... }@inputs:
    let
      system = "aarch64-darwin"; # System for ARM64 macOS
      pkgs = nixpkgs.legacyPackages.${system};    
    in
    {
      devShells.${system}.default = pkgs.mkShell {
        packages = with pkgs; [
            go
        ];
      };
    };
}