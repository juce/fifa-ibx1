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


def put(li, item):
    if item not in li:
        li.append(item)


def enumerateStrings(node, strings):
    """
    Walk the XML document structure and find all strings, which are either:
        names of elements, or
        value and name of "property" elements of type "string"
    """
    if node.nodeType == Node.ELEMENT_NODE:
        value = node.nodeName
        if value == 'property':
            put(strings, node.getAttribute('name'))
            if node.getAttribute('type') == 'string':
                value = node.getAttribute('value')
                put(strings, value)
        else:
            put(strings, value)
    # mimic original encoder:
    #   first, process child elements that aren't "property" elements
    #   then, process "property" child elements
    elems, props = [], []
    for x in node.childNodes:
        if x.nodeType == Node.ELEMENT_NODE and x.nodeName == 'property':
            props.append(x)
        else:
            elems.append(x)
    for x in elems:
        enumerateStrings(x, strings)
    for x in props:
        enumerateStrings(x, strings)


# step 1: enumerate all strings
strings = []
enumerateStrings(doc.documentElement, strings)

debug(hex(len(strings)))
for i, s in enumerate(strings):
    debug('%s %s' % (hex(i), s))
