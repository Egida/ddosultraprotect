go env -w GOPRIVATE=github.com/meldyer1/ddosultraprotect
export file=~/.gitconfig
export sub="insteadOf"
if [ -z "$(grep -e "$sub" $file)" ]; then echo -e "[url \"git@github.com:\"]\n insteadOf = https://github.com/" >> $file; fi
