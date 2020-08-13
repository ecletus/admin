#!/usr/bin/env bash

which dirname >/dev/null || {
    echo '`dirname` not found' >&2
    exit 1
}

d=$(dirname $0)
codepoints="$d/../../../fonts/codepoints"

[ ! -e "$codepoints" ] && {
    echo '`'"$codepoints"'` does not exists' >&2
    exit 1
}

echo "// Auto generated file. Do not edit it.
" > "$d/_icons.scss"

cat $codepoints | while read l
do
    name=$(echo "$l" | cut -d" " -f 1)
    code=$(echo "$l" | cut -d" " -f 2)
    echo "[mdl-icon-name*=\"$name\"] > a::before {
  content: \"\\$code\";
}" >> "$d/_icons.scss"
done