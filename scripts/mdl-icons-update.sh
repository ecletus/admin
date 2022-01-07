#!/usr/bin/env bash

mod="_git_submodules/material-design-icons"
fonts_dst="assets/static/admin/fonts"
icons_dst="assets/static/admin/stylesheets/scss/mdl-icons/_icons.scss"
codepoints="$mod/font/MaterialIcons-Regular.codepoints"

[ ! -e "$codepoints" ] && {
    echo '`'"$codepoints"'` does not exists' >&2
    exit 1
}

cp -a $mod/font/*.{ttf,otf,codepoints} "$fonts_dst" || exit $?

echo "// Auto generated file. Do not edit it.
" > "assets//_icons.scss"

cat /dev/null > "$icons_dst"

cat $codepoints | while read l
do
    name=$(echo "$l" | cut -d" " -f 1)
    code=$(echo "$l" | cut -d" " -f 2)
    echo "[mdl-icon-name*=\"$name\"] > a::before {
  content: \"\\$code\";
}" >> "$icons_dst"
done