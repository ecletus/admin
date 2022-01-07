#!/usr/bin/env python3
import os
import shutil
import subprocess
import sys

import whatthepatch

vars = {}
prefix = "/* override */"
root = os.getcwd()
src_path = "assets/static/admin/stylesheets/scss/mdl/_variables.scss"
mdl_path = "_git_submodules/material-design-lite"
dst_path = "src/_variables.scss"
src_path = os.path.relpath(src_path, mdl_path)

mdl_dst_path = os.path.relpath("assets/static/admin/stylesheets/vendors/material.min.css", mdl_path)
mdl_src_path = "dist/material.min.css"

os.chdir(mdl_path)

patchs = {
    'src/data-table/_data-table.scss': [
        [54, '    padding: 0 $data-table-column-padding 12px $data-table-column-padding;',
         '    padding: 0 $data-table-column-padding $data-table-column-padding-bottom $data-table-column-padding;'],
        [58, '      padding-left: 24px;', '      padding-left: 5px;'],
        [68, '    vertical-align: middle;', '    vertical-align: $data-table-cell-valign;'],
        [76, '      vertical-align: middle;', '      vertical-align: $data-table-cell-valign;'],
        [88, '    padding-bottom: 8px;', '    padding-bottom: 0;']
    ],
    'src/dialog/_dialog.scss': [
        [23, '    @include dialog-width;', '    width: max-content;max-width:80%;']
    ]
}


def cmd(cmd):
    subprocess.run(cmd, shell=True, check=True, stdout=sys.stdout, stderr=sys.stderr)


cmd("git checkout src/_variables.scss " + " ".join([k for k in patchs]))

for (f, items) in patchs.items():
    text = []

    for path in items:
        text.append(f"@@ -{path[0]},1 +{path[0]},1 @@\n-{path[1]}\n+{path[2]}")

    diff = [x for x in whatthepatch.parse_patch('\n'.join(text))]
    diff = diff[0]

    with open(f, "r") as fh:
        lao = fh.read()

    tzu = whatthepatch.apply_diff(diff, lao)

    with open(f, "w") as fg:
        fg.write("\n".join(tzu))


class Val:
    def __init__(self, value):
        self.value = value
        self.count = 0


class State:
    key = None
    value = None
    started = False
    line = 0
    set = None
    write = None

    def __init__(self, write):
        self.write = write

    def check_done(self, lf):
        if lf[len(lf) - 1] == ';':
            self.set()
            self.started = False
            self.key, self.value = "", ""
            return True
        return False

    def do(self, l):
        start = l.find('$')
        if start != 0:
            if self.started:
                self.value += " " + l.strip()
                self.check_done(l.strip())
            else:
                self.write(l.rstrip('\r\n'))
            return

        lf = l.strip()
        pos = lf.find(':')

        if pos < 0:
            if self.started:
                self.value += " " + lf
                self.check_done(l.strip())
            else:
                self.write(l.rstrip('\r\n'))
            return

        if lf[len(lf) - 1] != ';':
            if self.started:
                self.value += " " + lf
            else:
                self.key, self.value = lf[:pos], lf[pos + 1:].strip()
                self.started = True
            return
        elif self.started:
            self.started = False

        if not self.started:
            self.key, self.value = lf[:pos], lf[pos + 1:].strip()

        self.set()

    def set_to_out(self):
        if self.key in vars:
            val = vars[self.key]
            val.count += 1
            new_value = val.value
            if new_value != self.value:
                out.append(f"{prefix} {self.key}: {new_value}")
                return
        out.append(f"{self.key}: {self.value}")

    def set_to_vars(self, vars):
        vars[self.key] = Val(self.value)


with open(src_path, "r") as a:
    s = State(lambda v: None)
    s.set = lambda: s.set_to_vars(vars)

    for l in a:
        s.line += 1
        s.do(l)

out = []

with open(dst_path, "r") as b:
    s = State(out.append)
    s.set = s.set_to_out

    for l in b:
        s.line += 1
        s.do(l)

with open(dst_path, "w") as b:
    b.write("\n".join(out))
    b.write("\n\n\n/* NEWS VARIABLES */\n")
    for (key, val) in vars.items():
        if val.count == 0:
            b.write(f"\n{key}: {val.value}")

gulp_local = "node_modules/gulp/bin/gulp.js"
gulp = "gulp"

if os.path.exists(gulp_local):
    gulp = gulp_local

cmd(gulp)

shutil.copyfile(mdl_src_path, mdl_dst_path)

os.chdir(os.path.join(root, "../core"))

gulp_local = "node_modules/gulp/bin/gulp.js"
gulp = "gulp"

if os.path.exists(gulp_local):
    gulp = gulp_local

cmd(gulp + " vendors.css")
cmd(gulp + " release.css+")
