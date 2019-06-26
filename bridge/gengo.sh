#!/bin/bash

fullfile=$1
filename=$(basename -- "$fullfile")
extension="${filename##*.}"
filename="${filename%.*}"

vyper -f bytecode $fullfile > $filename.bin
vyper -f abi $fullfile > $filename.abi
vyper -f external_interface $fullfile > ${filename}_interface.vy
abigen -abi $filename.abi -bin $filename.bin -pkg bridge -out $filename.go

