# fifa-ibx1
encoder/decoder tools for files in IBX1 format used in FIFA games.

## Build

- on Linux/macOS: ```make```
- on Windows: ```build.cmd```

## Usage

### XML --> DAT

```
% ./xml2dat 
FIFA IBX1 Encoder by juce. Version: 1.2-575635bce1a25451663bc17e41704179d855e26a
Usage: ./xml2dat <in-path> <out-path> [options]
Options:
	--debug   : print out extra info for troubleshooting
	--noshare : do not re-use typed values (produces larger IBX1 files)
```

### DAT --> XML

```
% ./dat2xml 
FIFA IBX1 Decoder by juce. Version: 1.2-575635bce1a25451663bc17e41704179d855e26a
Usage: ./dat2xml <in-path> <out-path> [options]
Options:
	--debug : print out extra info for troubleshooting
	--hex8  : output 8-bit integers in hexadecimal format
	--hex16 : output 16-bit integers in hexadecimal format
	--hex32 : output 32-bit integers in hexadecimal format
```
