#!/bin/bash

fullfile=$1
filename=$(basename -- "$fullfile")
extension="${filename##*.}"
filename="${filename%.*}"

vyper -f bytecode $fullfile > $filename.bin
vyper -f abi $fullfile > $filename.abi
abigen -abi $filename.abi -bin $filename.bin -pkg bridge -out $filename.go
