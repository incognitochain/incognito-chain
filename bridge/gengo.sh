#!/bin/bash

fullfile=$1
pkgname=$2
filename=$(basename -- "$fullfile")
filepath=${fullfile%/*}
extension="${filename##*.}"
filename="${filename%.*}"

vyper -f bytecode $fullfile > $filepath/$filename.bin
vyper -f abi $fullfile > $filepath/$filename.abi
vyper -f external_interface $fullfile > $filepath/${filename}_interface.vy

abigen -abi $filepath/$filename.abi -bin $filepath/$filename.bin -pkg $pkgname -out $filepath/$filename.go

