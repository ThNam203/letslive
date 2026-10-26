# LL-HLS Live Pipeline — Design (Phase 1)

Status: draft for review · Date: 2026-10-26 · Service: `backend/transcode`

## Problem

Live playback is about 10–18s behind the streamer, and the transcode service uses more CPU than it needs to.

- One ffmpeg process does everything: it decodes, re-encodes **every** configured quality (including the top one), encodes audio once per quality, cuts `.ts` segments and writes playlists to disk (`transcoder/ffmpeg.go`).
- A watcher polls the disk every 100ms, uploads each segment to MinIO and rewrites the playlists (`watcher/`). Viewers download from MinIO.
- Viewers wait for a whole segment, then the poll, then the upload, and the player buffers about 3 full segments before it starts.
- A second ffmpeg decodes every frame to save one thumbnail per minute. It reads from the same `io.Pipe` through `io.MultiWriter`, so a slow thumbnail stalls the live stream.
- Every configured quality is always produced, so a 480p source is stretched up to 1080p and 720p.
- ffmpeg's HLS muxer cannot produce Apple LL-HLS (partial segments, preload hints, blocking playlist reload).

## Goal

- Glass-to-glass latency of **≤ 4s** with hls.js in low-latency mode.
- ffmpeg is used **only to encode the lower qualities**. It does not handle networking, segments, playlists, disk or audio.
- The source quality is copied, never re-encoded.
- Viewers are offered only qualities at or below what the streamer sends.
- Thumbnails and VOD uploads cannot block the live path.
- The live URL stays the same: `{API_URL}/transcode/{livestreamId}/index.m3u8`. The VOD output keeps the same layout and segment URL format.

## Architecture

```
OBS ─RTMP─► Session (Go, per connection)
              ├─ source H.264 + AAC ───────────────────────────► Stream.rendition[0] "source"
              ├─ FLV (bounded queue, drops on overflow) ─► ffmpeg ─TS pipes─► feeders ─► Stream.rendition[1..n]
              ├─ AAC ───────────────────────────────────────────► every rendition (no audio encode)
              ├─ every keyframe (rate-limited) ─► Thumbnailer (one-shot ffmpeg → JPEG → MinIO)
              └─ every rendition ─► Recorder (TS files at keyframes → MinIO) → VOD playlists on end
Stream = one gohlslib LL-HLS Muxer per rendition + a merged master playlist, served by Registry over HTTP.
```

### Components

| Component | Responsibility |
|---|---|
| `live.ProbeAVCConfig` | Parse the RTMP AVC decoder config (SPS/PPS) into width, height and fps. |
| `live.BuildLadder` | Rendition 0 = source. Then add every configured quality whose short side is **smaller** than the source's short side. Never upscale. |
| `live.EncoderArgs` / `live.Encoder` | ffmpeg: one decode, `split`, `scale`, libx264 `-tune zerolatency`, `-force_key_frames source`, `-an`, one MPEG-TS pipe per rendition (fd 3+). |
| `live.Stream` | One `gohlslib.Muxer` (LL-HLS, fMP4) per rendition, all sharing the source AAC track. Serves the merged master playlist and each muxer's own files. |
| `live.Registry` | Maps `livestreamId` to `*Stream`; the HTTP handler for `/static/{id}/...`. |
| `live.Session` | Per RTMP connection: waits for configs and the first keyframe, then builds the ladder and wires everything. Feeds the encoder through a bounded queue, so a slow or crashed ffmpeg never blocks the source rendition. |
| `live.Recorder` | Per rendition: writes MPEG-TS files cut at keyframes (≥ `recordSegmentDurationMs`), uploads them via `storage.Storage.AddSegment` in the background, and writes VOD playlists at the end. |
| `live.Thumbnailer` | At most once per `thumbnailIntervalMs`: one keyframe → `ffmpeg -frames:v 1` → JPEG → `storage.Storage.AddThumbnail`. Drops the request if it is busy. |

### Decisions

- **One Muxer per rendition.** `gohlslib` v2 supports only one video track per Muxer. The master playlist merges each muxer's own `index.m3u8`: URIs are prefixed with the rendition name and audio `GROUP-ID`s are renamed per rendition.
- **No audio encoding.** The source AAC goes to every rendition. If the source has no AAC config before its first keyframe, every rendition is video-only.
- **Time base.** A shared `Clock` with zero = DTS of the first source keyframe. ffmpeg runs with `-copyts`, so its TS timestamps share the FLV time base. Feeders unwrap 33-bit timestamps and subtract the same zero.
- **Rendition names** (URL path segment): `source`, then `{shortSide}p` (e.g. `720p`, `480p`).
- **Portrait sources.** Compare short sides. Scale with `-2:{short}` for landscape and `{short}:-2` for portrait.
- **Encoder failure.** If ffmpeg exits or its queue overflows, the source rendition keeps playing. The encoded renditions stop updating, and the error is logged with the livestream id.

### HTTP

| Path (transcode webserver) | Served by |
|---|---|
| `/static/vods/...` | existing `http.FileServer(PublicHLSPath)` (VOD unchanged) |
| `/static/{id}/index.m3u8` | `Registry`: merged master playlist (404 until every rendition has a complete segment) |
| `/static/{id}/{rendition}/{file}` | `Registry`: that rendition's `gohlslib.Muxer.Handle` (query string preserved for `_HLS_msn`, `_HLS_part`) |

Envoy already rewrites `/transcode` → `/static` with `timeout: 0s`, so no gateway change is needed.

### Configuration (`transcode.live`, in the config server repo)

| Key | Default | Meaning |
|---|---|---|
| `partMinDurationMs` | 200 | LL-HLS part length |
| `segmentMinDurationMs` | 1000 | segment cut at the first keyframe after this |
| `segmentCount` | 7 | segments kept in memory (minimum 7 for LL-HLS) |
| `recordSegmentDurationMs` | 4000 | VOD segment length |
| `encoderQueuePackets` | 512 | bounded queue between the session and ffmpeg |
| `thumbnailIntervalMs` | 60000 | thumbnail cadence |

The existing `ffmpegSetting.qualities[].resolution` must be `WxH` (e.g. `1280x720`). The service refuses to start otherwise.

## Removed

`transcoder/ffmpeg.go` (live transcoder), `watcher/watcher.go`, `watcher/watcher_utils.go`, `watcher/vod.go`, `watcher/minio/`, and the `github.com/radovskyb/watcher` dependency. The uploaded-video path (`transcoder/file_transcoder.go`, `worker/`) is unchanged.

## Out of scope (later phases)

SRT/WHIP ingest, NVENC, multi-node routing and CDN, resolution changes mid-stream (logged; the session continues with the original ladder), non-H.264 or non-AAC sources (rejected or video-only, see above).

## Libraries

- `github.com/bluenviron/gohlslib/v2` (LL-HLS muxer, MIT, Go 1.26)
- `github.com/bluenviron/mediacommon/v2` v2.9.7 (MPEG-TS reader/writer, H.264 SPS/AVCC, MPEG-4 audio config)
- Existing: `github.com/nareix/joy5` (RTMP, FLV)
