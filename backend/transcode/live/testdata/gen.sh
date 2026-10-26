#!/usr/bin/env bash
# Regenerates the FLV fixtures used by the live package tests.
# Keyframes every 2s (-g 60 at 30fps), B-frames on so composition time is non-zero.
set -euo pipefail
cd "$(dirname "$0")"

gen() {
	local name=$1 size=$2 audio=$3
	local args=(-y -loglevel error -f lavfi -i "testsrc2=size=${size}:rate=30:duration=6")
	if [ "$audio" = yes ]; then
		args+=(-f lavfi -i "sine=frequency=440:sample_rate=44100:duration=6")
	fi
	args+=(-c:v libx264 -preset veryfast -b:v 300k -maxrate 300k -bufsize 600k -pix_fmt yuv420p -g 60 -keyint_min 60 -sc_threshold 0 -bf 2)
	if [ "$audio" = yes ]; then
		args+=(-c:a aac -ar 44100 -b:a 48k)
	else
		args+=(-an)
	fi
	ffmpeg "${args[@]}" -f flv "$name"
}

gen src_640x360.flv 640x360 yes
gen src_360x640.flv 360x640 yes
gen src_640x360_noaudio.flv 640x360 no
gen src_1280x720.flv 1280x720 yes
