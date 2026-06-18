#!/bin/bash
QEMU=qemu-system-x86_64
# Use OVMF firmware
$QEMU -bios /usr/share/ovmf/OVMF.fd -drive file=fat:rw:./,format=raw -net none -nographic -serial mon:stdio
