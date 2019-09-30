#!/bin/sh
free -m | awk 'NR==2{printf "%.2f",$3*100/$2 }'
