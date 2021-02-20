#!/usr/bin/env python3

"""
IBX1 encoder for FIFA
"""

import sys
import struct
from xml.dom import minidom, Node


if len(sys.argv)<3:
    print(f"Usage {sys.argv[0]} <input.xml> <output.DAT> [--noshare] [--debug]")
    sys.exit(0)

infile = sys.argv[1]
outfile = sys.argv[2]

_share = True
_debug = False
for x in sys.argv[3:]:
    if x == '--noshare':
        # do not reuse typed-values: this will create larger DAT files
        _share = False
    if x == '--debug':
        # produce debug output
        _debug = True

def debug(*args, **kwargs):
    if _debug:
        print(*args, **kwargs)


with open(infile, 'rb') as f:
    doc = minidom.parse(f)

#debug('xml file successfully parsed')


class Li(list):
    def __init__(self):
        self.di = dict()


def get(li, item):
    return li.di[item]


def put(li, item):
    val = li.di.get(item)
    if not val:
        val = len(li)
        li.di[item] = val
        li.append(item)
    return val


def add(li, item, key=None):
    key = key or item
    val = li.di.get(key) if _share else None
    if val is None:
        val = len(li)
        li.di[key] = val
        li.append(item)
    return val


def count_extra_attributes(node):
    count = 0
    if node.nodeType == Node.ELEMENT_NODE and node.nodeName != "property":
        attrs = node.attributes
        count += attrs.length
    for x in node.childNodes:
        count += count_extra_attributes(x)
    return count


def enumerate_strings(node, strings):
    """
    Walk the XML document structure and find all strings, which are either:
        names of elements, or
        value and name of "property" elements of type "string"
    """
    value = node.nodeName
    if value == 'property':
        put(strings, node.getAttribute('name'))
        if node.getAttribute('type') == 'string':
            value = node.getAttribute('value')
            put(strings, value)
    else:
        put(strings, value)

    # mimic fifa walking order:
    #   first, process child elements that aren't "property" elements
    #   then, process "property" child elements
    elems, props = [], []
    for x in node.childNodes:
        if x.nodeType == Node.ELEMENT_NODE:
            if x.nodeName == 'property':
                props.append(x)
            else:
                elems.append(x)
    for x in elems:
        enumerate_strings(x, strings)
    for x in props:
        enumerate_strings(x, strings)


def encode_number(x):
    if x < 0x40:
        return struct.pack('>B', x)
    if x < 0x100:
        return b'\x40' + struct.pack('>B', x)
    if x < 0x10000:
        return b'\x80' + struct.pack('>H', x)
    return b'\xc0' + struct.pack('>I', x)


def create_tree(node, strings, values):
    """
    Walk the XML document structure and find all strings, which are either:
        names of elements, or
        value and name of "property" elements of type "string"
    """
    name = node.nodeName
    if name == 'property':
        name = node.getAttribute('name')
        typ = node.getAttribute('type')
        value = node.getAttribute('value')

        name_index = get(strings, name)

        tval = None
        if typ == 'float':
            value = float(value)
            tval = (0xb0, 'float', value, struct.pack('<f', value))
        elif typ == 'string':
            # lookup string index
            idx = get(strings, value)
            if idx < 0x10:
                tval = (0xc0 + idx, 'string', idx, None)
            elif idx < 0x100:
                tval = (0xd0, 'string', idx, struct.pack('>B', idx))
            else:
                tval = (0xe0, 'string', idx, struct.pack('>H', idx))
        elif typ in ['byte', 'int8', 'uint8']:
            fmt = '<B' if typ == 'uint8' else '<b'
            try:
                value = int(value)
            except ValueError:
                value = int(value, 16)
                fmt = '<B'
            if value >= 0 and value < 16:
                tval = (value, typ, value, None)
            else:
                tval = (0x10, typ, value, struct.pack(fmt, value))
        elif typ in ['short', 'int16', 'uint16']:
            fmt = '<H' if typ == 'uint16' else '<h'
            try:
                value = int(value)
            except ValueError:
                value = int(value, 16)
                fmt = '<H'
            tval = (0x20, typ, value, struct.pack(fmt, value))
        elif typ in ['int','int32', 'uint32']:
            fmt = '<I' if typ == 'uint32' else '<i'
            try:
                value = int(value)
            except ValueError:
                value = int(value, 16)
                fmt = '<I'
            tval = (0x30, typ, value, struct.pack(fmt, value))
        elif typ == 'bool':
            if value.lower() == 'true':
                tval = (0x41, 'bool', value, None)
            else:
                tval = (0x40, 'bool', value, None)

        value_index = add(values, tval, key=(tval[0], tval[3]))
        return {'kind': 'prop', 'name_index': name_index, 'value_index': value_index}

    name = node.nodeName
    name_index = get(strings, name)

    # mimic fifa walking order:
    #   first, process child elements that aren't "property" elements
    #   then, process "property" child elements
    elems, props = [], []
    for x in node.childNodes:
        if x.nodeType == Node.ELEMENT_NODE:
            if x.nodeName == 'property':
                props.append(x)
            else:
                elems.append(x)
    child_elems = []
    for x in elems:
        elem = create_tree(x, strings, values)
        child_elems.append(elem)
    child_props = []
    for x in props:
        prop = create_tree(x, strings, values)
        child_props.append(prop)
    return {'kind': 'Elem', 'name_index': name_index, 'props': child_props, 'elems': child_elems}


def output_tree(tree, f):
    if tree['kind'] == 'prop':
        name_index = tree['name_index']
        value_index = tree['value_index']
        if name_index + 0x80 < 0xa0:
            f.write(struct.pack('>B', 0x80 + name_index))
            f.write(encode_number(value_index))
        elif name_index < 0x100:
            f.write(b'\xa0')
            f.write(struct.pack('>B', name_index))
            f.write(encode_number(value_index))
        else:
            f.write(b'\xc0')
            f.write(struct.pack('>H', name_index))
            f.write(encode_number(value_index))
    else:
        # element
        f.write(b'\0')
        name_index = tree['name_index']
        f.write(encode_number(name_index))
        f.write(encode_number(len(tree['props'])))
        f.write(encode_number(len(tree['elems'])))

        # mimic fifa walking order:
        #   first, process "property" child elements
        #   then, process child elements that aren't "property" elements
        for x in tree['props']:
            output_tree(x, f)
        for x in tree['elems']:
            output_tree(x, f)


# step 0: sanity-check XML file
extra_attrs = count_extra_attributes(doc.documentElement)
if extra_attrs > 0:
    # cannot encode. Just mirror as-is
    with open(infile, 'rb') as inf:
        with open(outfile, 'wb') as outf:
            outf.write(inf.read())
    sys.exit(0)

# step 1: enumerate all strings
strings = Li()
enumerate_strings(doc.documentElement, strings)

debug(hex(len(strings)))
for i, s in enumerate(strings):
    debug('%s %s' % (hex(i), s))

# step 2: create tree and enumerate typed values
typed_values = Li()
tree = create_tree(doc.documentElement, strings, typed_values)

debug(hex(len(typed_values)))
for i, v in enumerate(typed_values):
    typ, name, value, bstr = v
    if name == 'string':
        value = strings[value]
    debug("%s ('%s', %s, %s)" % (hex(i), hex(typ), repr(name), repr(value)))

debug(tree)


# final step: output
with open(outfile, 'wb') as f:
    f.write(b'IBX1')
    # strings
    f.write(encode_number(len(strings)))
    for s in strings:
        f.write(encode_number(len(s)))
        f.write(s.encode('utf-8'))
        f.write(b'\0')
    # typed values
    f.write(encode_number(len(typed_values)))
    for v in typed_values:
        typ, name, value, bstr = v
        f.write(struct.pack('>B', typ))
        if bstr:
            f.write(bstr)
    # encoding
    f.write(b'\x01')
    # document structure
    output_tree(tree, f)
