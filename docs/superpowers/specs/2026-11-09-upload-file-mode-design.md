---
ticket: LL-HLS
status: draft
---

# Uploads on the live engine ("file mode") — design

Date: 2026-11-09 · Branch: `feat/ll-hls-packager` · Builds on
[2026-10-26-ll-hls-pipeline-design.md](2026-10-26-ll-hls-pipeline-design.md)

## Why

Recorded livestreams and uploaded videos are made by two different engines:

- **Live:** `backend/transcode/live` (`Session` → ladder, encoder, `Recorder`,
  `Thumbnailer`). It copies the source as the top quality, encodes only the
  qualities below it, never upscales, and names VOD folders by ladder index
  (`0` = source).
- **Uploads:** `backend/transcode/transcoder/file_transcoder.go`
  (`TranscodeFile`), one ffmpeg that re-encodes every configured quality
  (including 1080p from a 720p file) into `0` = 360p, `1` = 720p, `2` = 1080p,
  with everything stored in MinIO.

The owner wants **one engine**, so an upload comes out the same as a
recorded stream. Measured on 2026-11-09, the live engine also used 33–67% less
CPU than re-encoding everything.

## Goals

- Uploads go through the live engine: the same ladder rules, recorder,
  storage layout, playlists (with `CODECS`) and thumbnailer as a recorded
  livestream.
- `transcoder.TranscodeFile` and the upload-only HLS upload code are deleted.
- Uploads keep running as fast as the machine allows (not real time).
- Any input ffmpeg can read is accepted, as today.

## Non-goals

- Re-processing videos uploaded before this change. They keep their MinIO
  playlists and URLs and must keep playing.
- Changing live behaviour or live output in any way.
- The open review issues L17–L20 in `docs/ISSUES.md` (separate work).

## Anti-goals (failure even if it "works")

- A live stream gets worse: more latency, stalls, or dropped quality because
  of this change.
- An upload is marked `ready` while missing a quality, frames or audio. The
  live engine hides such failures on purpose to keep a stream going; an
  upload must fail instead, so the job is retried.
- A portrait phone video comes out sideways.
- An upload with non-AAC audio comes out silent.

## Decisions (owner's)

1. **One engine with a file mode** (option A): the same `Session` code with a
   mode switch, not a second engine and not real-time playback of the file.
2. **File mode uses normal compression settings**, not live's low-latency
   ones (`-tune zerolatency`). Ladder, folders and playlists are identical to
   live; only the encoded files are smaller.
3. **Uploads are capped at 1080p** by short side (portrait included). Larger
   files are scaled down to 1080p on intake, and that version becomes the top
   ("source") quality.
4. **New upload playlists live where recorded streams' do:** on the transcode
   server's `transcode_hls_public` volume, served at
   `{vodPlaybackUrlPrefix}/{vodId}/index.m3u8` (`/transcode/vods/...`).
   Segments stay in MinIO.
5. Old uploads are left as they are.

## Design

### Flow

```
worker job ──► download raw file (as today)
          ──► intake: ffprobe, then one ffmpeg ──FLV pipe──► joy5 FLV demuxer
          ──► Session (file mode, id = vodId) ──► Recorder ─► MinIO segments
                                                          └─► public/vods/{vodId}/ playlists
                                               ──► Thumbnailer ─► MinIO thumbnail
          ──► Close: ok → UpdateVODStatus(ready, playbackURL, thumbnailURL, duration)
                     err → existing markJobFailed (attempts / retry)
```

### 1. Intake

A new step turns any input into what the engine accepts: an FLV stream of
H.264 + AAC, at most 1080p short side, upright, with correct timestamps.

- `ffprobe` the downloaded file: first video stream codec, profile, pixel
  format, size, rotation (display matrix / `rotate` tag), field order and
  colour transfer; first audio stream codec; and the keyframe times of the
  video stream (`-skip_frame nokey`, a fast scan of the whole file).
- **Video is copied** only when all of these hold. Otherwise it is
  **converted once**:
  - H.264, 8-bit 4:2:0, SDR;
  - short side ≤ 1080;
  - not rotated;
  - progressive (not interlaced);
  - no gap between keyframes longer than 2 × `recordSegmentDurationMs`
    (8 s with today's 4 s segments). The recorder only cuts segments at a
    keyframe, so longer gaps would give long, uneven segments and poor
    seeking. *(My call, including the threshold.)*
- **Converting** means `libx264` at the source size (scaled to 1080p short
  side when larger), auto-rotated, deinterlaced when interlaced, with
  keyframes at most every 2 s, using file mode's normal compression
  settings (preset/CRF from `ffmpegSetting`).
- **HDR / 10-bit input** (the iPhone default: HEVC HLG, Dolby Vision, PQ) is
  tone-mapped to SDR BT.709 with `zscale` + `tonemap` during conversion, so
  it does not come out washed out. Both filters ship in the production
  image's ffmpeg (Alpine `ffmpeg` 8.1.3, checked 2026-11-09); no new
  dependency. *(My call; the old transcoder did not tone-map.)*
- **Audio is copied** when it is AAC, otherwise converted to AAC. Only the
  first audio stream is used; a file without audio becomes a video-only VOD
  (as live). *(My call.)*
- A file with no video stream fails the job (the engine needs H.264
  keyframes; `ErrNothingRecorded`). *(My call; the old transcoder failed too.)*
- MP4 edit lists and start offsets are applied by ffmpeg's demuxer, so
  timestamps start at the first presented frame.
- The FLV goes through a pipe into `joy5/format/flv.NewDemuxer`, whose
  packets feed `Session.HandlePacket`, the same packet type RTMP delivers.

The bitrate of a file ≤ 1080p that is copied is not capped. *(My call; the
owner capped resolution only.)*

### 2. Session file mode

`SessionDeps` gains a mode (live, the default, or file). Live behaviour is
untouched; in file mode:

| | Live (unchanged) | File mode |
|---|---|---|
| Full encoder / recorder queue | drop until next keyframe | **wait** (bounded by the session context) |
| Live LL-HLS `Stream` + `Registry` | created and registered | **none**; `Registry` not required |
| Encoder args | `-tune zerolatency`, live probe/flush flags | **normal x264 tuning**; preset/CRF/maxrate/bufsize/GOP/FPS rules as live |
| Encoder or rendition failure | drop the rendition, keep going | **`Close` returns an error** (job fails, is retried) |
| Recorder drops, abandoned segments, failed segment uploads | logged | **`Close` returns an error** |
| Unparsable video access units, mid-stream resolution change | logged, ladder kept | **`Close` returns an error** |
| Packets before the first keyframe, audio just before the first video frame, empty NAL units | discarded | discarded (normal in good files, not an error) |
| Thumbnails | wall clock: first after 5 s, then every 60 s | **media time:** one thumbnail, first keyframe at ≥ 5 s, or the first keyframe when the video is shorter *(my call)* |
| Thumbnail URL | discarded | **returned** to the caller after `Close` |
| Log wording | "livestream {id}" | "upload {id}" |

Everything else is shared: ladder (`BuildLadder`; source copied, only lower
qualities encoded), `Clock` media timeline, `Recorder` segment length and
layout (`{id}/{index}/{n}.ts` in MinIO, playlists and master with `CODECS`
under `public/vods/{id}/`), `Session.Duration()`.

**Retry with conversion** *(my call)*: when a file-mode attempt fails on the
input (unparsable access units, resolution change, ffmpeg failure on a copied
stream), the next attempt of the same job forces the convert path, so a
quirky H.264 file that slipped past the copy rules still ends up `ready`
instead of using up every attempt the same way.

### 3. Worker

`doTranscode` keeps job claiming, download, temp dir cleanup and
`markJobFailed` as they are, and replaces `TranscodeFile`,
`getHLSDurationSeconds` and `uploadHLSToStorage` with intake + file-mode
session. On success it sends `ready` with:

- playback URL `{vodPlaybackUrlPrefix}/{vodId}/index.m3u8` (same form as
  `finishStream` uses for recorded streams),
- the thumbnail URL from the session,
- the duration from `Session.Duration()` (seconds, rounded up like live).

### 4. Sharing the machine with live streams

A file-mode encode runs flat out on the same host as live sessions. It must
not starve live encoders (which would then drop packets). The worker already
runs one job at a time; the builder chooses the mechanism on top (an ffmpeg
`-threads` limit and/or lower process priority), configurable with a sane
default.

## Effects outside transcode

- **vod service:** none. The status flow is unchanged (`processing` →
  `ready`/`failed` via `PATCH /v1/internal/vods/{id}/status`).
- **web / mobile:** none expected; new upload playback URLs are the same
  `/transcode/vods/...` form recorded streams already use. Old uploads keep
  their MinIO URLs.
- **nginx:** `/transcode/vods/` is already an uncached pass-through.

## Testing

- Unit: file mode never drops under input faster than the encoder; no
  `Stream` is registered; an encoder failure, a dropped rendition or a failed
  segment upload makes `Close` fail; thumbnails follow media time (short and
  long inputs); live-mode tests still pass unchanged.
- Intake: the copy-or-convert decision as a table over probe results.
- End to end, with the real ffmpeg on generated files: H.264/AAC MP4 (copied),
  HEVC MP4, VP9/Opus WebM, a rotated portrait MP4, a 4K file, a file without
  audio, a file without video. Assert ladder, folder order, `CODECS`,
  durations, upright output, audio present, thumbnail URL, and the error case.
- Benchmark against the deleted `TranscodeFile` on the same file (CPU, wall
  time, size), as in the 2026-11-09 comparison.

## Code facts the builder must handle

Checked against the code by the design reviewer:

- `sessionEncoder.stop()` ends ffmpeg with SIGTERM instead of draining it;
  file mode must close ffmpeg's input and wait, or the end of the encoded
  renditions is lost.
- `FeedTS` and the 10 s encoded-ready deadline depend on a live `Stream`,
  which file mode does not have.
- `Thumbnailer` uses wall-clock time (`time.Now`) for its start delay and
  interval.
- `Recorder.WriteVideo`/`WriteAudio` drop when the queue is full, and
  `Finish` gives up after 60 s; a long upload needs file-mode waiting and a
  finish bound that fits it.

## Left to the builder

Package and type layout, how waiting is implemented, ffprobe invocation and
parsing, exact non-live x264 flags, GOP/FPS of converted output within the
rules above, audio channel handling, cleanup after a failed attempt, what a
thumbnail failure does (it must not fail the job), the CPU-sharing mechanism
and its default, and log wording beyond the "upload {id}" prefix.
