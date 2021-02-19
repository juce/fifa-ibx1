#!/usr/bin/env python3

"""
IBX1 encoder for FIFA
"""

import sys
import struct
from xml.dom import minidom, Node

if len(sys.argv)<3:
    print(f"Usage {sys.argv[0]} <input.xml> <output.DAT> [--debug]")
    sys.exit(0)

with open(sys.argv[1],"rb") as f:
    doc = minidom.parse(f)

# optional debug flag
_debug = False
if len(sys.argv) > 3 and sys.argv[3] == "--debug":
    _debug = True

def debug(*args, **kwargs):
    if _debug:
        print(*args, **kwargs)

debug('xml file successfully parsed')


def get(li, item):
    for i, x in enumerate(li):
        if x == item:
            return i
    raise IndexError(f'no such element: {item}')


def put(li, item):
    if item not in li:
        li.append(item)


def add(li, item):
    li.append(item)


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


def enumerate_typed_values(node, strings, typed_values):
    """
    Walk the XML document structure and find all typed values,
        which are located in 'property' elements
    """
    if node.nodeName == 'property':
        typ = node.getAttribute('type')
        value = node.getAttribute('value')
        if typ == 'float':
            value = float(value)
            add(typed_values, (0xb0, 'float', value, struct.pack('<f', value)))
        elif typ == 'string':
            # lookup string index
            idx = get(strings, value)
            if idx < 16:
                add(typed_values, (0xc0 + idx, 'string', idx, None))
            elif idx < 256:
                add(typed_values, (0xd0, 'string', idx, struct.pack('>B', idx)))
            else:
                add(typed_values, (0xe0, 'string', idx, struct.pack('>H', idx)))
        elif typ == 'int':
            value = int(value)
            if value >= 0 and value < 16:
                add(typed_values, (value, 'int', value, None))
            elif value >= 0 and value < 256:
                add(typed_values, (0x10, 'int', value, struct.pack('<b', value)))
            elif value >= 0 and value < 65536:
                add(typed_values, (0x20, 'int', value, struct.pack('<h', value)))
            else:
                add(typed_values, (0x30, 'int', value, struct.pack('<i', value)))
        elif typ == 'bool':
            if value.lower() == 'true':
                add(typed_values, (0x41, 'bool', value, None))
            else:
                add(typed_values, (0x40, 'bool', value, None))

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
        enumerate_typed_values(x, strings, typed_values)
    for x in props:
        enumerate_typed_values(x, strings, typed_values)


def encode_number(x):
    if x < 0x40:
        return struct.pack('>B', x)
    if x < 0x80:
        return b'\x40' + struct.pack('>B', x)
    return b'\x80' + struct.pack('>H', x)


def create_tree(node, strings, value_index):
    """
    Walk the XML document structure and find all strings, which are either:
        names of elements, or
        value and name of "property" elements of type "string"
    """
    name = node.nodeName
    if name == 'property':
        name = node.getAttribute('name')
        name_index = get(strings, name)
        return {'kind': 'prop', 'name_index': name_index, 'value_index': value_index}, value_index+1

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
        elem, value_index = create_tree(x, strings, value_index)
        child_elems.append(elem)
    child_props = []
    for x in props:
        prop, value_index = create_tree(x, strings, value_index)
        child_props.append(prop)
    return {'kind': 'Elem', 'name_index': name_index, 'props': child_props, 'elems': child_elems}, value_index


def output_tree(tree, f):
    if tree['kind'] == 'prop':
        name_index = tree['name_index']
        value_index = tree['value_index']
        if name_index + 0x80 < 0xa0:
            f.write(struct.pack('>B', 0x80 + name_index))
            f.write(encode_number(value_index))
        elif name_index < 256:
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


# step 1: enumerate all strings
strings = []
enumerate_strings(doc.documentElement, strings)

debug(hex(len(strings)))
for i, s in enumerate(strings):
    debug('%s %s' % (hex(i), s))

# step 2: enumerate typed values
typed_values = []
enumerate_typed_values(doc.documentElement, strings, typed_values)

debug(hex(len(typed_values)))
for i, v in enumerate(typed_values):
    typ, name, value, bstr = v
    if name == 'string':
        value = strings[value]
    debug("%s ('%s', %s, '%s')" % (hex(i), hex(typ), name, value))

# step 3: create structure
tree, value_index = create_tree(doc.documentElement, strings, 0)
debug(hex(value_index))
debug(tree)


# final step: output
with open(sys.argv[2], 'wb') as f:
    f.write(b'IBX1')
    # strings
    f.write(encode_number(len(strings)))
    for s in strings:
        f.write(struct.pack('>B', len(s)))
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
