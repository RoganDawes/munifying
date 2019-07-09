# munifying by Marcus Mengs (@MaMe82)

The tool `munifying` could be used to interact with Logitech receivers from USB side (not RF).
This tool was developed during vulnerability research and is provided as-is.

The main purpose of munifying is the demonstration of the extraction of AES link encryption keys and device RF addresses 
of paired devices from a Logitech receiver dongle via USB (CVE-2019-13054 and CVE-2019-13055) or, at least, to accomplish
the re-pairing of devices, which indirectly leads to AES key extraction based on passive RF sniffing during pairing
(CVE-2019-13052).

While direct extraction only requires physical access to the receiver, the re-pairing approach requires access to 
receiver and device. A vendor patch, on the other hand, will only be applied for USB-based AES key extraction.

## supported functionality:

- set receiver into pairing mode
- unpair a single device or all devices
- dump AES keys for devices vulnerable to CVE-2019-13054 or CVE-2019-13055 (disabled in pre-release)
    - working Unifying dongles: CU0008, CU0012 (likely other TI CC2544 based dongles with WPID 0x8808)
    - working presentation clicker dongles: CU0016 for Spotlight (046d:c540), CU0016 for R500 (046d:c53e)
    - not working Unifying dongles: CU0007 (likely other Nordic nRF24 based dongles with WPID 0x8802)
- dump readable memory range with undocumented HID++ command (disabled in pre-release, used for AES key derivation)
- store full receiver dump to file in order to re-import it to `mjackit` (disabled for pre-release)
- show Info of all devices paired to a receiver

## Note on pre-release

In sense of responsible disclosure, I agreed with Logitech, not to publish in-depth content related to CVE-2019-13054
or CVE-2019-13055 (USB based AES key extraction from Texas Instruments based dongles).
The pre-release is only provided in binary form (pre-compiled for Linux, x64).

For Unifying devices, CVE-2019-13052 exists. This vulnerability covers AES key derivation based on a sniffed device
pairing (or re-pairing). **This vulnerability will not be fixed.** Although, AES key extraction is currently disabled,
`munifying` could be used to unpair and pair devices. In conjunction with `LOGITacker` or `mjackit` this allows AES key
sniffing, using passive RF monitoring. For Unifying receivers, unpairing and re-pairing could be done with the official 
Unifying software. **Logitech SPOTLIGHT and R500 presentation clickers have an undocumented pairing functionality, which
ultimately could be used to replicate CVE-2019-13052 for those presentation remotes, too**

In order to re-pair an R500/SPOTLIGHT:

1) Unpair the device from the receiver with `munifying unpairall`
2) Set the receiver to pairing mode `munifying pair`
3) Press and hold the two keys, which are meant to enable Bluetooth mode

In order to retrieve the AES link encryption key, `mjackit` or `LOGITacker` have to be ran in "pair sniffing" mode.
In contrast to `mjackit`, `LOGITacker` is able to bypass the key blacklisting filter of the presentation remotes for
keystroke injection against Windows hosts (see CVE-2019-13054 for details) 

## Requirements

Although written in Go, the `munifying` tool has a native dependency on `libusb-1.0-0`.

## Usage

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

## Supported Logitech receivers (tested)

- Logitech Unifying: CU0007, CU0008, CU0012
- Logitech R500 presentation remote: CU0016
- Logitech SPOTLIGHT presentation remote: CU0016


## Unsupported receivers (no HID++ capabilities)

- CU0010 used by M171 mouse (no key extraction needed, unencrypted)
- CU0014 used by R400 (no key extraction needed, unencrypted)

## DISCLAIMER
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