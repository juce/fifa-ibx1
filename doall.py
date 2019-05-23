#!/usr/bin/env python3

import os
import sys

if len(sys.argv)<3:
    print(f'Usage: {sys.argv[0]} <input_dir|input_file> <output_dir>')
    sys.exit(0)

src = sys.argv[1]
dst = sys.argv[2]
log = open("doall.log", "wt")

def mkdirs(path):
    if not os.path.exists(path):
        os.makedirs(path)

def convert_file(pathname):
    dstname = pathname.replace(f'{src}/', f'{dst}/')
    dstname, _ = os.path.splitext(dstname)
    dstname = f'{dstname}.xml'
    ddir, dfile = os.path.split(dstname)
    mkdirs(ddir)
    cmd = f'{sys.executable} reader.py {pathname} {dstname}'
    print(f'converting {pathname} --> {dstname} ... ', flush=True, end="")
    res = os.system(cmd)
    if res == 0:
        print('OK')
        print(f'converting {pathname} --> {dstname} : OK', file=log)
    else:
        print('FAILED')
        print(f'converting {pathname} --> {dstname} : FAILED', file=log)
    return 1

def convert_dir(path):
    count = 0
    files = os.listdir(path)
    for name in files:
        relname = path + '/' + name
        if os.path.isdir(relname):
            count += convert_dir(relname)
        if relname.lower().endswith('.dat'):
            count += convert_file(relname)
    return count


if os.path.isdir(src):
    total = convert_dir(src)
else:
    total = convert_file(src)

print(f'converted {total} files')
print(f'converted {total} files', file=log)
log.close()
