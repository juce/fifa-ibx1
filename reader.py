import sys
import struct

with open(sys.argv[1],"rb") as f:
    data = f.read()

print(len(data), file=sys.stderr)

sig = data[:4]
num1 = data[4]
num2 = data[5]
print(num1, file=sys.stderr)
print(num2, file=sys.stderr)

strings = []

# read strings
offs = 6
for i in range(num2):
    clen = data[offs]
    value = data[offs+1:offs+1+clen].decode("utf-8")
    zero = data[offs+1+clen]
    offs += clen + 2
    print(hex(i), value, file=sys.stderr)
    strings.append(value)

# read typed values
print(data[offs], file=sys.stderr)
num_typed_values = data[offs+1]
print(num_typed_values, file=sys.stderr)

offs += 2
values = []

def get_typed_value(data, offs):
    typ = data[offs]
    if typ < 0x30:
        # one-byte int
        v = (hex(typ), "int", typ)
        offs += 1
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

print(values, file=sys.stderr)

# encoding
encoding = data[offs]
offs += 1

print(f'encoding: {encoding}', file=sys.stderr)


def get_prop(data, offs):
    ptyp = data[offs]
    name = strings[data[offs+1]]
    vtyp = data[offs+2]
    if ptyp == 0xa0 and vtyp == 0x40:
        value = values[data[offs+3]]
        offs += 4
    elif ptyp == 0xa0:
        value = values[vtyp]
        offs += 3
    else:
        name = strings[data[offs] - 0x80]
        value = values[data[offs+1]]
        offs += 2
    print(('property',name,value), file=sys.stderr)
    return {'Elem': 'property', 'attrs': {'name':name, 'value':value}}, offs

def get_element(data, offs):
    print(f'offs: {hex(offs)}', file=sys.stderr)
    zero = data[offs]
    name = strings[data[offs+1]]
    num_properties = data[offs+2]
    num_children = data[offs+3]
    print(f'Elem: {name}, {num_properties} props, {num_children} children', file=sys.stderr)
    props = []
    offs += 4
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

