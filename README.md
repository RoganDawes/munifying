# munifying by Marcus Mengs

The tool `munifying` could be used to interact with Logitech receivers from USB end.
It was developed during vulnerability research and is provided as-is.

The main purpose is to demo extraction of AES link encryption keys and device RF addresses from the dongle via USB
(CVE-2019-13054 and CVE-2019-13055) or at least support re-pairing of devices, which again leads to AES key extraction
using if the RF part of pairing is sniffed (CVE-2019-13052).

The tool was tested for the following receivers:

- Logitech Unifying: CU0007, CU0008, CU0012
- Logitech R500 presentation remote: CU0016
- Logitech SPOTLIGHT presentation remote: CU0016

Supported functionality:
- set receiver into pairing mode
- unpair a single device or all devices
- dump AES keys for devices vulnerable to CVE-2019-13054 or CVE-2019-13055 *(disabled in pre-release)*
    - working **Unifying dongles**: CU0008, CU0012 (likely other TI CC2544 based dongles with WPID 0x8808)
    - working **presentation clicker dongles**: CU0016 for Spotlight (046d:c540), CU0016 for R500 (046d:c53e)
    - not working Unifying dongles: CU0007 (likely other Nordic nRF24 based dongles with WPID 0x8802)
- dump readable memory range with undocumented HID++ command (used for AES key derivation) *(disabled in pre-release)*
- store full receiver dump to file in order to re-import it to `mjackit` *(disabled in pre-release)*
- show Info of all devices paired to a receiver

## Note on pre-release

In sense of responsible disclosure, I agreed with Logitech, not to publish in-depth content related to CVE-2019-13054
or CVE-2019-13055 (USB based AES key extraction from Texas Instruments based dongles) before the respective patch is
released in August 2019.

**The munifying pre-release is only provided in binary form (pre-compiled for Linux, x64) and is fairly limited.**

So why bother ?

For Unifying devices CVE-2019-13052 covers AES key derivation based on a sniffed device pairing (or re-pairing). 
**This vulnerability will not be addressed by Logitech.** Although, direct USB AES key extraction is currently 
disabled, `munifying` could be used to unpair and pair devices. In conjunction with `LOGITacker` or `mjackit` this 
allows AES key sniffing during pairing, using passive RF monitoring. For Unifying receivers, unpairing and re-pairing 
could, of course, be done with the official Unifying software. **But Logitech SPOTLIGHT and R500 presentation clickers 
provide undocumented pairing functionality, which ultimately could be used to replicate CVE-2019-13052 for those 
presentation remotes, too. Munifying is capable to put the respective receivers into pairing mode.**

In order to re-pair an R500/SPOTLIGHT device to its receiver:

1) Unpair the device from the receiver with `munifying unpairall`
2) Set the receiver to pairing mode `munifying pair`
3) Press and hold the two keys, which are meant to enable Bluetooth mode

In order to retrieve the AES link encryption key, `mjackit` or `LOGITacker` have to be ran in "pair sniffing" mode.
In contrast to `mjackit`, `LOGITacker` is able to bypass the key blacklisting filter of the presentation remotes for
keystroke injection against Windows hosts (see CVE-2019-13054 for details) 

Full release will be published at: 

# Requirements

Although written in Go, the `munifying` tool has a native dependency on `libusb-1.0-0`.

# Usage

```
Usage:
  munifying [command]

Available Commands:
  dump        Dump dongle memory utilizing secret HID++ command
  help        Help about any command
  info        Lists relevant information of first receiver found on USB
  pair        Pair new devices to first receiver found on USB
  patchdump   Dumps RAM using firmwaremod for CU0007 (not published)
  store       Store relevant information of first receiver found on USB to file (usable with 'mjackit')
  unpair      Unpair devices of first receiver found on USB
  unpairall   Unpair all paired devices of first receiver found on USB

Flags:
  -h, --help   help for munifying

Use "munifying [command] --help" for more information about a command.
```

Note:
There is no support for multiple receivers connected to USB at the same time. The tool always interacts with
the first receiver discovered on USB bus.

# Unsupported receivers (no HID++ capabilities)

- CU0010 used by M171 mouse (no key extraction needed, unencrypted)
- CU0014 used by R400 (no key extraction needed, unencrypted)

# DISCLAIMER
- no responsibility for damage or destroyed devices 
- no illegal use, meant for demonstration and educational purposes

**munifying** is a Proof of Concept and should be used for authorized testing and/or 
educational purposes only. The only exception is using it against devices
or a network, owned by yourself.

I take no responsibility for the abuse of munifying or any information given in
the related documents. 

**I DO NOT GRANT PERMISSIONS TO USE munifying TO BREAK THE LAW.**

As munifying is meant as a Proof of Concept, it is likely that bugs occur.
I disclaim any warranty for munifying, it is provided "as is".