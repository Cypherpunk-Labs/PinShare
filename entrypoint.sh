#!/bin/bash 
/usr/local/bin/start_ipfs daemon --migrate=true --agent-version-suffix=docker &
/opt/pinshare/bin/pinshare
