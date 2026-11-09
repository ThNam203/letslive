# LL-HLS live pipeline — review problems

Date: 2026-11-09 · Branch: `feat/ll-hls-packager` · Compared against: `main`

Problems found while comparing the LL-HLS rework with the old ffmpeg + disk
watcher pipeline. Each fix is its own commit that ticks its entry here.

| # | Problem | Status |
|---|---|---|
| 1 | Wrong stream duration | fixed |
| 2 | Livestream ends late; VOD published even when the recording failed | fixed |
| 3 | Non-AAC audio becomes a silent stream without saying so | fixed |
| 5 | Every live viewer request reaches the transcode service | fixed (needs host nginx apply) |
| 6a | First thumbnail taken at the first keyframe | fixed |
| 6b | VOD master has no `CODECS` attribute | fixed |
| 6c | VOD folder `0` is now the source quality | won't fix (intentional) |
| 7 | Branch far behind `main`, `go.work.sum` conflict | fixed |
| 8 | `/transcode/vods/` lists every VOD id publicly | open |

(#4, iOS LL-HLS over HTTP/2, was dropped: prod serves HTTP/2 and iOS plays it.)

## 1. Wrong stream duration

`rtmp/rtmp.go` computed the duration as wall-clock time since connect minus a
fixed 7 s, *after* `session.Close()`. Close waits for the last VOD segments to
upload (up to 75 s), so that wait was counted, and the 7 s was the old ffmpeg
start-up delay, which the new pipeline does not have.

**Fixed:** the `Session` tracks the last source-video DTS on the shared clock and
reports `Duration()`; the RTMP server uses it and the 7 s offset is gone. A
session that never reached its first keyframe reports 0. Tests:
`TestSessionDuration`, `TestSessionDurationNeverStarted`.

## 2. Livestream ends late; VOD published even when the recording failed

`HandleConnection` closed the session (writing the VOD) before calling
`EndLivestream`, so the livestream stayed "live" in the DB for up to 75 s after
the streamer stopped. It then always sent the VOD playback URL, even when
`Recorder.Finish` failed, and the vod service creates livestream VODs as
`ready` — leaving a VOD that points at a playlist that does not exist. (`main`
had the same false-ready problem.)

**Fixed:** the livestream is ended as soon as the RTMP connection ends
(`rtmp/finish.go`, `finishStream`):

1. `EndLivestream` is sent with `vodStatus: "processing"`, the media duration
   and no playback URL. The vod service creates the VOD as `processing`
   (`status` on `POST /v1/internal/vods`; empty still means `ready`, anything
   other than `ready`/`processing` is a 400), and the livestream service now
   returns `{"vodId": ...}`.
2. The session is closed, which writes the VOD.
3. Transcode marks the VOD `ready` with its playback URL and duration, or
   `failed` (no playback URL) when Close returned an error — including the new
   `live.ErrNothingRecorded` for a session that never reached a keyframe.

Related changes in the same commit:

- `GetPublicVODsByUser` (a channel's public VOD list) now only returns `ready`
  VODs, like the popular and search queries already did, so a VOD that is
  still processing or failed is not shown as an unplayable card. The web mock
  for `GET /vods` mirrors it. The SQL change is compile-checked only; there is
  no database test harness in the vod service.
- The transcode gateway's `EndLivestream` and `UpdateVODStatus` requests now
  carry their context, so the 10 s timeouts actually apply.
- Gateway calls use a context detached from the server's, so a stream that
  ends during a service shutdown is still ended and its VOD still settled.

**Deploy order:** vod, then livestream, then transcode. A new transcode
against an old livestream service gets no VOD id back; it logs that and the
VOD stays as the old service created it (`ready`, no playback URL).

Tests: `TestFinishStream` (rtmp), `TestEndLivestream*` and
`TestUpdateVODStatusHonoursContext` (transcode gateway),
`TestCreateVODInternal*` (vod), `TestEndLivestreamRequestDTOVODStatus`
(livestream), and the updated `TestSessionCloseBeforeFirstKeyframe`.

## 3. Non-AAC audio becomes a silent stream without saying so

joy5 (the RTMP library) only demuxes AAC; any other audio codec is discarded
before the session sees it. This was already true on `main` (its ffmpeg got
joy5's packets), so it is not a regression, and transcoding other codecs would
need a different RTMP library.

**Fixed:** when a stream starts without a usable AAC config the session logs a
Warn saying the stream is video-only and that only AAC audio is supported.
Supporting other codecs stays out of scope (it needs an RTMP library other
than joy5). Test: `TestSessionWarnsWhenVideoOnly`.

## 5. Every live viewer request reaches the transcode service

Live playlists, parts and segments are now served from the transcode process
(segments used to come from MinIO). Viewers of one stream ask for the same
part or blocking playlist at nearly the same moment.

**Fixed:** `configs/nginx.conf` gets two `/transcode/` locations backed by a
cache zone defined in the new `configs/nginx-http.conf` (that file goes in the
host nginx's `http { }` block):

- Parts, segments and init files are cached for as long as gohlslib's
  `Cache-Control` allows (an hour). Errors are not cached.
- Playlists: only LL-HLS blocking requests (with `_HLS_msn`) are cached, for
  1 s, overriding gohlslib's `no-cache`. The master, plain reloads and VOD
  playlists always go to the gateway.
- `proxy_cache_lock` turns concurrent identical requests into one upstream
  request. The cache key includes `Origin`, because Envoy echoes it in
  `Access-Control-Allow-Origin` with credentials.

Verified with nginx 1.27 in Docker against a fake upstream: 20 concurrent
blocking playlist requests → 1 upstream hit; another origin → its own entry
with its own `Access-Control-Allow-Origin`; the same request after 1.5 s → a
new hit; master, plain reload, VOD playlist and 404 → never cached; 21 part
requests → 1 hit.

**To apply:** the host nginx is not deployed by CI. Add
`include .../nginx-http.conf;` to its `http` block, reload it with the new
`nginx.conf`, and create `/var/cache/nginx/letslive_hls` if nginx cannot.

## 6a. First thumbnail taken at the first keyframe

The old thumbnail filter skipped the first 5 s (`gte(t,5)`); the new
thumbnailer used the very first keyframe, often a black or title frame.

**Fixed:** the thumbnailer ignores keyframes for the first 5 s after the stream
starts (`thumbnailStartDelay`); the 60 s cadence follows from the first one it
takes. Test: `TestThumbnailerSkipsStart`.

## 6b. VOD master has no `CODECS` attribute

The old VOD master was ffmpeg's and carried `CODECS`; the recorder's master
had only `BANDWIDTH` and `RESOLUTION`.

**Fixed:** each variant gets `CODECS="avc1.PPCCLL[,mp4a.40.N]"` (before
`RESOLUTION`, as gohlslib orders it), built from that rendition's own SPS and
the source audio config. When the SPS is unknown `CODECS` is left out rather
than listing audio only. Tests: `TestRecorderMasterCodecs`,
`TestAVCCodecString`.

## 6c. VOD folder `0` is now the source quality

Old VODs used folders `0/1/2` = 360p/720p/1080p; new VODs use the ladder index
(`0` = source, then descending). Nothing reads live VOD folder numbers: web,
mobile and the vod service only use the master playlist URL. The one
`0/stream.m3u8` reader in the codebase is the upload worker
(`worker.getHLSDurationSeconds`), on its own uploaded-video output. Old VODs
keep their layout and still play.

**Status:** intentional, no change. A folder name now matches the ladder
index used in the live URLs and the MinIO keys (`{id}/{index}/{n}.ts`).

## 7. Branch far behind `main`, `go.work.sum` conflict

The branch was cut at `4da77e43` and is 79 commits behind `origin/main`; the
only conflict is `backend/go.work.sum`.

**Fixed:** merged `origin/main` into the branch (it is already pushed, so no
rebase); `go.work.sum` is the union of both sides and every module builds.

## 8. `/transcode/vods/` lists every VOD id publicly

The VOD file server is a plain `http.FileServer`, which renders directory
listings: `GET /transcode/vods/` returned every VOD id, including those of
private VODs. Live on `main` in production.

**Fix:** the VOD handler returns 404 for any request that resolves to a
directory; playlists and other files are served as before.
