#!/usr/bin/env python3

import os
import sys

src = sys.argv[1]
dst = sys.argv[2]

def mkdirs(path):
    if not os.path.exists(path):
        os.makedirs(path)

def convert_file(pathname):
    dstname = pathname.replace(f'{src}/', f'{dst}/')
    dstname = dstname.replace('.DAT', '.xml')
    cmd = f'{sys.executable} reader.py {pathname} 2>>debug.log 1>{dstname}'
    print(f'converting {pathname} --> {dstname}')
    os.system(cmd)
    return 1

def convert_dir(path):
    count = 0
    files = os.listdir(path)
    for name in files:
        relname = path + '/' + name
        if os.path.isdir(relname):
            count += convert_dir(relname)
        if relname.endswith('.DAT'):
            count += convert_file(relname)
    return count


mkdirs(dst)
if os.path.isdir(src):
    total = convert_dir(src)
else:
    total = convert_file(src)

print(f'converted {total} files')
