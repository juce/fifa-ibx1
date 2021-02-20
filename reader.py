#!/usr/bin/env python3

"""
IBX1 Decoder for FIFA
"""

import sys
import struct

if len(sys.argv)<3:
    print(f"Usage {sys.argv[0]} <input.DAT> <output.xml> [--hex8] [--hex16] [--hex32] [--debug]")
    sys.exit(0)

infile = sys.argv[1]
outfile = sys.argv[2]

with open(infile, 'rb') as f:
    data = f.read()

_debug = False
_hex8, _hex16, _hex32 = False, False, False
for x in sys.argv[3:]:
    if x == '--hex8':
        _hex8 = True
    elif x == '--hex16':
        _hex16 = True
    elif x == '--hex32':
        _hex32 = True
    elif x == '--debug':
        _debug = True


def debug(*args, **kwargs):
    if _debug:
        print(*args, **kwargs)

#debug(len(data))

sig = data[:4]
if sig != b"IBX1":
    # not an IBX1 file : just mirror it to stdout
    with open(outfile, 'wb') as f:
        f.write(data)
    exit(0)


def get_value(data, offs):
    v = data[offs]
    if v == 0x40:
        # 1-byte num follows
        v = data[offs+1]
        return v, offs+2
    elif v == 0x80:
        # 2-byte big endian unsigned num follows
        v = struct.unpack('>H',data[offs+1:offs+3])[0]
        return v, offs+3
    elif v == 0xc0:
        # 4-byte big endian unsigned num follows
        v = struct.unpack('>I',data[offs+1:offs+5])[0]
        return v, offs+5
    return v, offs+1


offs = 4
num_strings, offs = get_value(data, offs)

debug(hex(num_strings))

strings = []

# read strings
for i in range(num_strings):
    clen, offs = get_value(data, offs)
    value = data[offs:offs+clen].decode("utf-8")
    zero = data[offs+clen]
    offs += clen+1
    debug(hex(i), value)
    strings.append(value)

# read typed values
num_typed_values, offs = get_value(data, offs)
debug(hex(num_typed_values))

values = []


def get_typed_value(data, offs):
    typ = data[offs]
    if typ < 0x10:
        # one-byte int
        v = (hex(typ), "int8", typ)
        offs += 1
    elif typ == 0x10:
        # 1-byte integer
        value = struct.unpack("<b",data[offs+1:offs+2])[0]
        v = (hex(typ), "int8", value)
        offs += 2
    elif typ == 0x20:
        # 2-byte integer
        value = struct.unpack("<h",data[offs+1:offs+3])[0]
        v = (hex(typ), "int16", value)
        offs += 3
    elif typ == 0x30:
        # 4-byte integer
        value = struct.unpack("<i",data[offs+1:offs+5])[0]
        v = (hex(typ), "int32", value)
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
    elif typ == 0xe0:
        # 2 byte unsigned int: big-endian
        value = struct.unpack(">H",data[offs+1:offs+3])[0]
        v = (hex(typ), "string", strings[value])
        offs += 3
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
    debug(hex(i), value)

# encoding
encoding = data[offs]
offs += 1

debug(f'encoding: {hex(encoding)}')


def get_prop(data, offs):
    ptyp = data[offs]
    if ptyp == 0xa0:
        name = strings[data[offs+1]]
        idx, offs = get_value(data, offs+2)
        value = values[idx]
    elif ptyp == 0xc0:
        nidx = struct.unpack('>H',data[offs+1:offs+3])[0]
        name = strings[nidx]
        idx, offs = get_value(data, offs+3)
        value = values[idx]
    else:
        name = strings[data[offs] - 0x80]
        idx, offs = get_value(data, offs+1)
        value = values[idx]
    debug(('property',name,value))
    return {'Elem': 'property', 'attrs': {'name':name, 'value':value}}, offs


def get_element(data, offs):
    debug(f'offs: {hex(offs)}')
    zero = data[offs]
    if zero != 0:
        raise Exception("elem byte not zero")
    v, offs = get_value(data, offs+1)
    name = strings[v]
    num_properties, offs = get_value(data, offs)
    num_children, offs = get_value(data, offs)
    debug(f'Elem: {name}, {num_properties} props, {num_children} children')
    props = []
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
debug(doc)


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
                if v[1] == 'int8' and _hex8:
                    elem.setAttribute("value", '0x%02X' % ((0x100 + v[2]) % 0x10))
                elif v[1] == 'int16' and _hex16:
                    elem.setAttribute("value", '0x%04X' % ((0x10000 + v[2]) % 0x10000))
                elif v[1] == 'int32' and _hex32:
                    elem.setAttribute("value", '0x%08X' % ((0x100000000 + v[2]) % 0x100000000))
                else:
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

with open(sys.argv[2], "wt") as f:
    f.write(xdoc.toprettyxml('  '))

