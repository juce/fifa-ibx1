#!/usr/bin/env python3

import sys
import struct

with open(sys.argv[1],"rb") as f:
    data = f.read()

print(len(data), file=sys.stderr)

sig = data[:4]
if sig != b"IBX1":
    # not an IBX1 file : just mirror it to stdout
    sys.stdout.buffer.write(data)
    exit(0)

num_strings = data[4]
offs = 5
if num_strings == 0x40:
    # multi-part
    num_strings = data[offs]
    offs += 1

print(hex(num_strings), file=sys.stderr)

strings = []

# read strings
for i in range(num_strings):
    clen = data[offs]
    offs += 1
    if clen == 0x40:
        # 1 byte value follows
        clen = data[offs]
        offs += 1
    value = data[offs:offs+clen].decode("utf-8")
    zero = data[offs+clen]
    offs += clen+1
    print(hex(i), value, file=sys.stderr)
    strings.append(value)

# read typed values
num_typed_values = 0
if data[offs] == 0x40:
    # 1 byte vaue follows
    num_typed_values = data[offs+1]
    offs += 2
elif data[offs] == 0x80:
    # 2 byte value follows (big endian)
    num_typed_values = struct.unpack('>h',data[offs+1:offs+3])[0]
    offs += 3
else:
    num_typed_values = data[offs]
    offs += 1
print(hex(num_typed_values), file=sys.stderr)

values = []

def get_typed_value(data, offs):
    typ = data[offs]
    if typ < 0x10:
        # one-byte int
        v = (hex(typ), "int", typ)
        offs += 1
    elif typ == 0x10:
        # 1-byte integer
        value = data[offs+1]
        v = (hex(typ), "int", value)
        offs += 2
    elif typ == 0x20:
        # 2-byte integer
        value = struct.unpack("<h",data[offs+1:offs+3])[0]
        v = (hex(typ), "int", value)
        offs += 3
    elif typ == 0x30:
        # 4-byte integer
        value = struct.unpack("<i",data[offs+1:offs+5])[0]
        v = (hex(typ), "int", value)
        offs += 5
    elif typ == 0xb0:
        # 4-byte float
        value = struct.unpack("<f",data[offs+1:offs+5])[0]
        v = (hex(typ), "float", value)
        offs += 5
    elif typ >= 0xc0 and typ < 0xd0:
        # string
        value = typ - 0xc0
        v = (hex(typ), "string", strings[value])
        offs += 1
    elif typ == 0xd0:
        # string
        value = data[offs+1]
        v = (hex(typ), "string", strings[value])
        offs += 2
    elif typ == 0x40:
        # bool: false
        value = data[offs+1]
        v = (hex(typ), "bool", False)
        offs += 1
    elif typ == 0x41:
        # bool: true
        value = data[offs+1]
        v = (hex(typ), "bool", True)
        offs += 1
    else:
        v = (hex(typ), "__unknown__", None)
        offs += 1
    return v, offs


while len(values) < num_typed_values:
    value, offs = get_typed_value(data, offs)
    values.append(value)

for i,value in enumerate(values):
    print(hex(i), value, file=sys.stderr)

# encoding
encoding = data[offs]
offs += 1

print(f'encoding: {hex(encoding)}', file=sys.stderr)


def get_prop(data, offs):
    ptyp = data[offs]
    if ptyp == 0xa0 and data[offs+2] == 0x40:
        name = strings[data[offs+1]]
        value = values[data[offs+3]]
        offs += 4
    elif ptyp == 0xa0 and data[offs+2] == 0x80:
        name = strings[data[offs+1]]
        value = values[struct.unpack('>h',data[offs+3:offs+5])[0]]
        offs += 5
    elif ptyp == 0xa0:
        name = strings[data[offs+1]]
        value = values[data[offs+2]]
        offs += 3
    else:
        name = strings[data[offs] - 0x80]
        b = data[offs+1]
        if b == 0x40:
            value = values[data[offs+2]]
            offs += 3
        elif b == 0x80:
            value = values[struct.unpack('>h',data[offs+2:offs+4])[0]]
            offs += 4
        else:
            value = values[b]
            offs += 2
    print(('property',name,value), file=sys.stderr)
    return {'Elem': 'property', 'attrs': {'name':name, 'value':value}}, offs

def get_element(data, offs):
    print(f'offs: {hex(offs)}', file=sys.stderr)
    zero = data[offs]
    v = data[offs+1]
    offs += 2
    if v == 0x40:
        v = data[offs]
        offs += 1
    elif v == 0x80:
        v = struct.unpack('>h',data[offs:offs+2])[0]
        offs += 2
    name = strings[v]
    num_properties = data[offs]
    num_children = data[offs+1]
    print(f'Elem: {name}, {num_properties} props, {num_children} children', file=sys.stderr)
    props = []
    offs += 2
    for i in range(num_properties):
        prop, offs = get_prop(data, offs)
        props.append(prop)
    children = []
    for i in range(num_children):
        child, offs = get_element(data, offs)
        children.append(child)
    return {'Elem': name, 'props': props, 'children': children}, offs


# document structure
doc, offs = get_element(data, offs)
print(doc, file=sys.stderr)

def make_element(xdoc, e):
    elem = xdoc.createElement(e['Elem'])
    for prop in e.get('props',[]):
        child = make_element(xdoc, prop)
        elem.appendChild(child)
    for c in e.get('children',[]):
        child = make_element(xdoc, c)
        elem.appendChild(child)
    attrs = e.get('attrs')
    if attrs:
        for k,v in attrs.items():
            if type(v) == tuple:
                elem.setAttribute("type", v[1])
                elem.setAttribute("value", str(v[2]))
                if v[2] is None:
                    elem.setAttribute("__tbyte", str(v[0]))
            else:
                elem.setAttribute(k, str(v))
    return elem

# convert to xml
from xml.dom import minidom
impl = minidom.getDOMImplementation()
xdoc = impl.createDocument('',doc['Elem'],'')
root = xdoc.documentElement
for c in doc['children']:
    child = make_element(xdoc, c)
    root.appendChild(child)

s = xdoc.toprettyxml('  ')
print(s)

