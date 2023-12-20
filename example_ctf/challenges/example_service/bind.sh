#! /bin/sh

LPORT=42023
yes "CTF[ITSASERVICE]" | socat TCP-LISTEN:$LPORT,reuseaddr,fork STDIN