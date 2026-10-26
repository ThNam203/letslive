# LL-HLS Live Pipeline Phase 1 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Serve live streams as LL-HLS from memory with ≤ 4s latency. ffmpeg only encodes the lower qualities, and the source quality is copied.

**Architecture:** A new package `backend/transcode/live` replaces the live ffmpeg transcoder and the disk watcher. Each RTMP connection gets a `Session`. The Session copies source H.264/AAC into a per-rendition `gohlslib` LL-HLS Muxer. It feeds ffmpeg (video only, one MPEG-TS pipe per lower rendition) through a bounded queue. It records every rendition to MPEG-TS files for VOD and sends rate-limited keyframes to a one-shot thumbnail ffmpeg. A `Registry` serves the merged master playlist and each muxer's files at the existing `/static/{id}/...` path.

**Tech Stack:** Go 1.26, `github.com/bluenviron/gohlslib/v2` v2.4.7, `github.com/bluenviron/mediacommon/v2` v2.9.7, `github.com/nareix/joy5` (existing), ffmpeg 8 (libx264), MinIO (existing `storage.Storage`).

**Spec:** `docs/superpowers/specs/2026-10-26-ll-hls-pipeline-design.md`

## Global Constraints

- Module `sen1or/letslive/transcode`; new code lives in `backend/transcode/live/` (package `live`).
- New dependencies: `github.com/bluenviron/gohlslib/v2@v2.4.7` and `github.com/bluenviron/mediacommon/v2@v2.9.7` only. Remove `github.com/radovskyb/watcher`.
- Live URL unchanged: web requests `{API_URL}/transcode/{id}/index.m3u8`; envoy rewrites it to `/static/{id}/index.m3u8` on transcode port `config.Webserver.Port`.
- VOD output unchanged: `{PublicHLSPath}/vods/{id}/{MasterFileName}` and `{PublicHLSPath}/vods/{id}/{i}/stream.m3u8`. Segment lines are `{remoteURL}?fileName={file}`, where `remoteURL` is returned by `storage.Storage.AddSegment(ctx, path, id, i)`.
- Thumbnail object name stays `thumbnail.jpg`, uploaded through `storage.Storage.AddThumbnail(ctx, path, id, "image/jpeg")`.
- Rendition names: index 0 = `source`; the others are `{shortSide}p`. Names must match `^[a-z0-9]+$`.
- Clock rates: H.264 = 90000; AAC = the sample rate from its AudioSpecificConfig (gohlslib rejects anything else).
- ffmpeg never encodes audio (`-an`) and never writes files on the live path.
- The live path never blocks on ffmpeg, the thumbnailer or MinIO. Each one gets a bounded queue that drops work and logs when full.
- Every error is handled; when it can't be returned, log it with `logger.Errorf(ctx, "...: %v", err)` and include the livestream id.
- Tests that need ffmpeg call `requireFFmpeg(t)`, which uses `t.Skip` when `exec.LookPath("ffmpeg")` fails. CI must have ffmpeg installed, or these tests prove nothing.
- Work on branch `feat/ll-hls-packager` (cut from fresh `main`; this repo has no `release` branch). Every commit body ends with `Refs: docs/superpowers/specs/2026-10-26-ll-hls-pipeline-design.md`.
- Verification per task: `cd backend/transcode && go vet ./... && go test ./live/... -race -count=1`.

## Review Focus

1. **Low-resolution source (e.g. 480p) with a 1080/720/480 ladder.** Expected: only `source` is offered, ffmpeg is not started, and nothing is upscaled. Pinned in Task 2 and Task 9.
2. **Portrait source (360x640).** Expected: short sides are compared, scaling uses `{short}:-2`, and the output stays portrait. Pinned in Task 2 and Task 3.
3. **ffmpeg slow, crashed or missing.** Expected: `source` keeps getting new parts; encoded renditions stop; the error is logged; the RTMP read loop never blocks. Pinned in Task 9.
4. **Source without audio.** Expected: every rendition is video-only, and the master playlist has no `EXT-X-MEDIA` line. Pinned in Task 5.
5. **Disconnect mid-segment, then reconnect.** Expected: the Registry drops the old id at once (404); the partial VOD segment is uploaded and listed; the new connection gets a new livestream id. Pinned in Task 7 and Task 9.

---

### Task 1: Config, fixtures, dependencies

**Files:**
- Modify: `backend/transcode/config/config.go` (extract `Quality`, add `Live`)
- Create: `backend/transcode/config/live.go`, `backend/transcode/config/live_test.go`
- Create: `backend/transcode/live/testdata/gen.sh` and its outputs `src_640x360.flv`, `src_360x640.flv`, `src_640x360_noaudio.flv`, `src_1280x720.flv`
- Create: `backend/transcode/live/helpers_test.go`
- Modify: `backend/transcode/go.mod`, `go.sum`

**Interfaces:**
- Produces:
  - `type Quality struct { Resolution, MaxBitrate string; FPS int; BufSize string }` (same yaml tags as today). `FFMpegSetting.Qualities` becomes `[]Quality`.
  - `func (q Quality) Size() (w, h int, err error)` parses `WxH`.
  - `type Live struct { PartMinDurationMs, SegmentMinDurationMs, SegmentCount, RecordSegmentDurationMs, EncoderQueuePackets, ThumbnailIntervalMs int }` with yaml keys from the spec. `Transcode.Live Live \`yaml:"live"\``.
  - `func (l Live) WithDefaults() Live` fills zeros with 200 / 1000 / 7 / 4000 / 512 / 60000.
  - `func Validate(t Transcode) error`.
  - Test helpers in `live/helpers_test.go`: `requireFFmpeg(t)`, `readFLV(t, name string) []av.Packet` (joy5 `flv.NewDemuxer`), `type recordingStorage` implementing `storage.Storage`: it copies files into `t.TempDir()`, records calls, and returns `http://minio.test/bucket/{id}/{q}/{base}`.

- [ ] **Step 1: Create the branch.** Run `git fetch origin && git switch -c feat/ll-hls-packager origin/main`.
- [ ] **Step 2: Add the dependencies.** Run `cd backend/transcode && go get github.com/bluenviron/gohlslib/v2@v2.4.7 github.com/bluenviron/mediacommon/v2@v2.9.7`. Expected: go.mod lists both.
- [ ] **Step 3: Write `testdata/gen.sh` and run it once.** Each fixture is 6s, 30fps, `-c:v libx264 -g 60 -keyint_min 60 -sc_threshold 0` (B-frames on, so CTime is non-zero), `-c:a aac -ar 44100`, `-f flv`, from `lavfi testsrc2` plus `sine`. The noaudio variant drops `sine` and uses `-an`. Commit the generated files (each < 1 MB).
- [ ] **Step 4: Write failing tests in `config/live_test.go`.**
  - `TestLiveWithDefaults`: `Live{}.WithDefaults()` equals `Live{200,1000,7,4000,512,60000}`; `Live{PartMinDurationMs: 300}.WithDefaults().PartMinDurationMs == 300`.
  - `TestQualitySize`: `"1280x720"` → `1280, 720, nil`; `"hd720"`, `"1280"`, `"0x720"` → error.
  - `TestValidate`: `SegmentCount: 6` → error containing `"segmentCount must be >= 7"`; a quality with resolution `"hd720"` → error containing `"qualities[0].resolution"`.
- [ ] **Step 5: Run `go test ./config/ -run 'Live|Quality|Validate'`.** Expected: FAIL (undefined).
- [ ] **Step 6: Implement `Quality`, `Live`, `WithDefaults`, `Size`, `Validate`.** Call `Validate` at the end of `PostProcess` and return its error, so the service refuses to start on bad config.
- [ ] **Step 7: Run `go test ./config/` and `go build ./...`.** Expected: PASS and a clean build (`file_transcoder.go` still compiles against `[]Quality`).
- [ ] **Step 8: Commit** `feat(transcode): live config, fixtures and LL-HLS deps`.

### Task 2: Source probe and ladder

**Files:**
- Create: `backend/transcode/live/source.go`, `ladder.go`, `ladder_test.go`, `source_test.go`

**Interfaces:**
- Consumes: `config.Quality`, `readFLV`.
- Produces:
  - `type SourceInfo struct { Width, Height int; FPS float64; SPS, PPS []byte }`, plus `func (s SourceInfo) ShortSide() int` and `func (s SourceInfo) Portrait() bool` (Height > Width).
  - `func ProbeAVCConfig(vseqhdr []byte) (SourceInfo, error)` parses the AVCDecoderConfigurationRecord and the SPS (`mediacommon/v2/pkg/codecs/h264.SPS`: `Width()`, `Height()`, `FPS()`). FPS is 0 when the SPS has no timing info.
  - `type Rendition struct { Name string; ShortSide int; Encoded bool; MaxBitrate, BufSize string; FPS int }`.
  - `func BuildLadder(src SourceInfo, qs []config.Quality) ([]Rendition, error)`.

- [ ] **Step 1: Write failing tests.**
  - `TestProbeAVCConfig`: the first `H264DecoderConfig` packet of `src_640x360.flv` → `Width 640, Height 360`, `FPS` within 0.01 of 30, non-empty SPS and PPS. Of `src_360x640.flv` → `Portrait() == true`. `ProbeAVCConfig([]byte{1,2})` → error.
  - `TestBuildLadder`, table-driven, with qualities 1920x1080, 1280x720, 854x480:
    - src 1920x1080 → names `[source 720p 480p]`; only index 0 has `Encoded == false`.
    - src 1280x720 → `[source 480p]`.
    - src 854x480 → `[source]` (**Review Focus 1**).
    - src 360x640 portrait → `[source]`; src 1080x1920 portrait → `[source 720p 480p]` (**Review Focus 2**).
    - Unsorted config input gives a descending result. Two qualities with the same short side → error `"duplicate rendition 720p"`.
    - `Rendition.FPS` = `min(q.FPS, round(src.FPS))`; when `src.FPS == 0` it is `q.FPS`.
- [ ] **Step 2: Run `go test ./live/ -run 'Probe|Ladder'`.** Expected: FAIL.
- [ ] **Step 3: Implement.** Short side of a quality = `min(w,h)` from `Size()`. Include a quality only when its short side is strictly less than `src.ShortSide()`.
- [ ] **Step 4: Run the tests.** Expected: PASS.
- [ ] **Step 5: Commit** `feat(transcode): probe source and build no-upscale ladder`.

### Task 3: Encoder (ffmpeg, video-only, TS pipes)

**Files:**
- Create: `backend/transcode/live/encoder.go`, `encoder_args.go`, `encoder_test.go`

**Interfaces:**
- Consumes: `Rendition`, `SourceInfo`, `requireFFmpeg`, `readFLV`.
- Produces:
  - `func EncoderArgs(src SourceInfo, encoded []Rendition, preset string, crf int) []string` takes only the renditions with `Encoded == true`, in ladder order.
  - `type Encoder struct` with `func StartEncoder(ctx context.Context, ffmpegPath string, args []string, outputs int) (*Encoder, error)`, `Input() io.WriteCloser`, `Outputs() []io.ReadCloser` (index i = rendition `encoded[i]`), `Wait() error`, `Stop()` (SIGTERM, then SIGKILL after 3s).

Exact argument order (copy into the test as the expected slice for 720p + 480p, landscape, preset `veryfast`, crf 23):
```
-hide_banner -loglevel error -fflags nobuffer -f flv -i pipe:0 -an -copyts
-filter_complex [0:v]split=2[s0][s1];[s0]scale=-2:720[v0];[s1]scale=-2:480[v1]
-map [v0] -c:v libx264 -preset veryfast -tune zerolatency -pix_fmt yuv420p -crf 23 -maxrate {q.MaxBitrate} -bufsize {q.BufSize} -force_key_frames source -sc_threshold 0 -g 300 [-r {fps} only if < src fps] -muxdelay 0 -muxpreload 0 -flush_packets 1 -f mpegts pipe:3
-map [v1] ... -f mpegts pipe:4
```
With one rendition the graph is `[0:v]scale=-2:480[v0]`. For a portrait source, scale is `{short}:-2`.

- [ ] **Step 1: Write failing tests.**
  - `TestEncoderArgsLandscape` and `TestEncoderArgsSingle`: exact slices as above.
  - `TestEncoderArgsPortrait`: the graph contains `scale=720:-2` (**Review Focus 2**).
  - `TestEncoderArgsFPSOnlyWhenLower`: with src 30fps and q.FPS 30 there is no `-r`; with q.FPS 24, `-r 24` follows `-g 300`.
  - `TestEncoderEndToEnd` (`requireFFmpeg`): write the packets of `src_1280x720.flv` through `flv.NewMuxer(enc.Input())`, close the input, and parse `Outputs()[0]` with `mpegts.Reader`. Expect exactly one H.264 track, SPS height 480, and a first video PTS within 3000 ticks (one frame) of `firstKeyframe.Time` in 90 kHz. This proves `-copyts` alignment. `Wait()` returns nil.
- [ ] **Step 2: Run `go test ./live/ -run Encoder`.** Expected: FAIL.
- [ ] **Step 3: Implement.** Outputs use `os.Pipe()` pairs passed through `exec.Cmd.ExtraFiles` (fd 3 = ExtraFiles[0]). Close the parent's copy of each write end after `Start`. Drain stderr in a goroutine to `logger.Errorf`.
- [ ] **Step 4: Run the tests.** Expected: PASS.
- [ ] **Step 5: Commit** `feat(transcode): video-only ffmpeg encoder with TS pipes`.

### Task 4: Clock and RTMP packet conversion

**Files:**
- Create: `backend/transcode/live/clock.go`, `rtmp_packet.go`, `clock_test.go`, `rtmp_packet_test.go`

**Interfaces:**
- Produces:
  - `type Clock struct` with `func NewClock(start time.Time, zero time.Duration) *Clock`, `Video90k(t time.Duration) int64` (= `(t - zero) * 90000 / time.Second`), `Audio(t time.Duration, sampleRate int) int64`, `NTP(pts90k int64) time.Time`.
  - `type TSUnwrapper struct` with `func (c *Clock) NewTSUnwrapper() *TSUnwrapper` and `Rel(raw int64) int64`. It returns `mpegts.TimeDecoder.Decode(raw)` plus `firstRaw - zero90k`, so unwrapped ffmpeg timestamps share the Clock's zero.
  - `func VideoAU(pkt av.Packet) ([][]byte, error)` turns AVCC into NAL units (`h264.AVCC.Unmarshal`).
  - `func AudioConfig(pkt av.Packet) (*mpeg4audio.AudioSpecificConfig, error)` reads from `pkt.ASeqHdr` / `pkt.Data`.

- [ ] **Step 1: Write failing tests.**
  - `TestClockVideo90k`: zero 1s, t 1.5s → 45000. `TestClockAudio`: zero 1s, t 2s, rate 44100 → 44100. `TestClockNTP`: start + 1s for 90000.
  - `TestTSUnwrapperWrap`: zero 0; raws `0x1FFFFFF00` then `0x100` → second minus first equals `0x200`.
  - `TestVideoAU`: the first `H264` keyframe of `src_640x360.flv` gives ≥ 1 NAL unit, and `h264.IsRandomAccess(au)` is true.
  - `TestAudioConfig`: the `AACDecoderConfig` packet gives `SampleRate 44100`.
- [ ] **Step 2: Run them.** Expected: FAIL. **Step 3: Implement.** **Step 4: Run them.** Expected: PASS.
- [ ] **Step 5: Commit** `feat(transcode): shared clock and RTMP packet conversion`.

### Task 5: Stream (per-rendition muxers, master playlist) and Registry

**Files:**
- Create: `backend/transcode/live/master.go`, `stream.go`, `registry.go`, `master_test.go`, `stream_test.go`

**Interfaces:**
- Consumes: `Rendition`, `Clock`, `config.Live`.
- Produces:
  - `func MergeMultivariant(parts []NamedPlaylist) ([]byte, error)` with `type NamedPlaylist struct { Name string; Body []byte }`. It keeps one `#EXTM3U`, the highest `#EXT-X-VERSION`, and one `#EXT-X-INDEPENDENT-SEGMENTS`. It rewrites `GROUP-ID="audio"` → `GROUP-ID="audio-{name}"` and `AUDIO="audio"` → `AUDIO="audio-{name}"`, prefixes `URI="x"` → `URI="{name}/x"` and each URI line after `#EXT-X-STREAM-INF` → `{name}/x`, and keeps input order.
  - `func NewStream(id string, rs []Rendition, audio *mpeg4audio.AudioSpecificConfig, cfg config.Live) (*Stream, error)` creates one `gohlslib.Muxer` per rendition (`Variant: MuxerVariantLowLatency`, `SegmentCount`, `SegmentMinDuration`, `PartMinDuration` from cfg). When `audio != nil`, each muxer's tracks are `[H264, MPEG4Audio]`.
  - `(*Stream).WriteVideo(r int, ntp time.Time, pts90k int64, au [][]byte) error`, `WriteAudio(ntp time.Time, pts int64, aus [][]byte) error` (writes to every muxer), `ServeHTTP(w, req)`, `Close()`.
  - Readiness: rendition `r` is ready after it has received its 2nd random-access AU (the first segment is closed). The master returns `404` with `Cache-Control: no-store` until every rendition is ready. After that it is built once from each muxer's in-process `index.m3u8` (`httptest.NewRecorder`) and cached.
  - `type Registry struct` with `NewRegistry()`, `Add(*Stream) error` (error on duplicate id), `Remove(id string)`, `Handler() http.Handler`. The handler serves `/{id}/index.m3u8` → master and `/{id}/{rendition}/{file}` → `muxer.Handle` (query preserved); anything else → 404. `id` must parse as a UUID, otherwise 400.

- [ ] **Step 1: Write failing tests.**
  - `TestMergeMultivariant`: two inputs shaped like gohlslib's LL output (copy the `"video+audio" && "lowLatency"` expectation from gohlslib `muxer_test.go` as each body). The expected output has exact lines with `audio-source`/`audio-480p` and `source/video1_stream.m3u8`/`480p/video1_stream.m3u8`.
  - `TestMergeMultivariantVideoOnly`: bodies without `EXT-X-MEDIA` → output has no `EXT-X-MEDIA` (**Review Focus 4**).
  - `TestStreamServesLLHLS`: build a Stream from `src_640x360.flv` with `[source]` and audio; feed every packet through `Clock`/`VideoAU`; serve it with `httptest.NewServer(registry.Handler())`. GET `/{id}/index.m3u8` → 200, containing `source/`. GET the media playlist → contains `#EXT-X-PART:` and `#EXT-X-PRELOAD-HINT`. GET with `?_HLS_msn={last+1}` from a goroutine returns only after one more keyframe is written (assert that it does not return within 100ms, then does return within 3s of writing).
  - `TestStreamMasterNotReady`: a fresh Stream → master 404 within 100ms.
  - `TestRegistry`: duplicate Add → error; after `Remove`, GET → 404; id `../x` → 400.
- [ ] **Step 2: Run them.** Expected: FAIL.
- [ ] **Step 3: Implement.**
- [ ] **Step 4: Run with `-race`.** Expected: PASS.
- [ ] **Step 5: Commit** `feat(transcode): LL-HLS stream with merged master playlist`.

### Task 6: TS feeder (ffmpeg output → Stream)

**Files:**
- Create: `backend/transcode/live/feeder.go`, `feeder_test.go`

**Interfaces:**
- Consumes: `Stream.WriteVideo`, `Clock.NewTSUnwrapper`, `Clock.NTP`.
- Produces: `func FeedTS(ctx context.Context, r io.Reader, s *Stream, rendition int, clk *Clock, onAU func(pts90k, dts90k int64, au [][]byte)) error`. It reads with `mpegts.Reader{R: r}`, uses the first H.264 track, and calls `WriteVideo` and then `onAU` (the recorder hook; may be nil). It returns `nil` on `io.EOF` and stops when `ctx` is done.

- [ ] **Step 1: Write a failing test.** `TestFeedTS` (`requireFFmpeg`): encode `src_1280x720.flv` to 480p with `StartEncoder`, feed `Outputs()[0]` into a Stream `[source 480p]` while also writing source packets. Expect `480p` to become ready, and the master to list both renditions.
- [ ] **Step 2: Run it.** Expected: FAIL. **Step 3: Implement.** **Step 4: Run it.** Expected: PASS.
- [ ] **Step 5: Commit** `feat(transcode): feed encoded TS into LL-HLS stream`.

### Task 7: Recorder (VOD)

**Files:**
- Create: `backend/transcode/live/recorder.go`, `recorder_test.go`

**Interfaces:**
- Consumes: `storage.Storage`, `Rendition`, `recordingStorage`.
- Produces:
  - `func NewRecorder(ctx context.Context, id string, rs []Rendition, audio *mpeg4audio.AudioSpecificConfig, st storage.Storage, dir string, segDur time.Duration) *Recorder`.
  - `WriteVideo(r int, pts90k, dts90k int64, au [][]byte)` and `WriteAudio(pts int64, aus [][]byte)` never block. Each rendition has a 1024-item channel; on overflow it drops the item and logs.
  - `Finish(publicHLSPath, masterFileName string) error`: closes the open segments, waits for uploads, and writes the VOD playlists in the layout from Global Constraints. Each variant playlist has `#EXT-X-VERSION:3`, `#EXT-X-TARGETDURATION:{ceil(max seg)}`, `#EXT-X-PLAYLIST-TYPE:VOD`, `#EXTINF:{dur:.3f},` per segment, and `#EXT-X-ENDLIST`. The master has one `#EXT-X-STREAM-INF:BANDWIDTH={bits/s},RESOLUTION={w}x{h}` per rendition, pointing at `{i}/stream.m3u8`. `BANDWIDTH` = total uploaded bytes × 8 / total duration; `{w}x{h}` comes from the first SPS NAL unit seen in that rendition (`h264.SPS.Unmarshal`).
  - A segment file `{dir}/{id}/{i}/{n}.ts` is cut at the first random-access AU once at least `segDur` has passed. It is written with `mpegts.Writer` and uploaded through `st.AddSegment(ctx, path, id, i)`; the local file is removed after upload. A failed upload is logged and that segment is left out of the playlist.

- [ ] **Step 1: Write failing tests.**
  - `TestRecorderVOD`: feed `src_640x360.flv` (6s) as `[source]` with `segDur` 2s. `Finish` writes `vods/{id}/index.m3u8` and `vods/{id}/0/stream.m3u8` with 3 `#EXTINF` lines. Each line after it is `http://minio.test/bucket/{id}/0/{n}.ts?fileName={n}.ts`, and `recordingStorage` recorded 3 `AddSegment` calls. The master contains `RESOLUTION=640x360` and a `BANDWIDTH` greater than 0.
  - `TestRecorderPartialSegmentOnFinish`: stop at 3s → 2 segments, the last one about 1s long (**Review Focus 5**).
  - `TestRecorderUploadFailure`: storage returns an error for segment 1 → the playlist has 2 entries and `Finish` returns nil.
- [ ] **Step 2: Run them.** Expected: FAIL. **Step 3: Implement.** **Step 4: Run with `-race`.** Expected: PASS.
- [ ] **Step 5: Commit** `feat(transcode): record renditions to VOD without blocking live`.

### Task 8: Thumbnailer

**Files:**
- Create: `backend/transcode/live/thumbnail.go`, `thumbnail_test.go`

**Interfaces:**
- Produces: `func NewThumbnailer(ctx context.Context, id, ffmpegPath string, src SourceInfo, st storage.Storage, every time.Duration) *Thumbnailer`, `Offer(au [][]byte)` (non-blocking; ignored if called before `every` has passed since the last accepted one, or if a job is running), `Close()`.
- The job writes Annex-B `SPS, PPS, au` to the stdin of `ffmpeg -hide_banner -loglevel error -f h264 -i pipe:0 -frames:v 1 -vf scale=640:-2 -q:v 2 -f image2 {tmp}/thumbnail.jpg`, then calls `st.AddThumbnail(ctx, path, id, "image/jpeg")`.

- [ ] **Step 1: Write failing tests.** `TestThumbnailer` (`requireFFmpeg`): offer 3 keyframes of `src_640x360.flv` within 10ms with `every` 1h → exactly 1 `AddThumbnail` call, and the file starts with `FF D8`. `TestThumbnailerOfferNeverBlocks`: with a storage that sleeps 2s, 100 `Offer` calls return in under 50ms in total.
- [ ] **Step 2: Run them.** Expected: FAIL. **Step 3: Implement.** **Step 4: Run them.** Expected: PASS.
- [ ] **Step 5: Commit** `feat(transcode): rate-limited non-blocking thumbnails`.

### Task 9: Session (per-connection orchestration)

**Files:**
- Create: `backend/transcode/live/session.go`, `session_test.go`

**Interfaces:**
- Consumes: everything above.
- Produces:
  - `type SessionDeps struct { Registry *Registry; Storage storage.Storage; Transcode config.Transcode; Now func() time.Time }`.
  - `func NewSession(ctx context.Context, id string, d SessionDeps) *Session`.
  - `HandlePacket(pkt av.Packet) error`. It must never block on ffmpeg, the recorder or the thumbnailer.
  - `Close(ctx context.Context) error`: `Registry.Remove`, stop the encoder, then `Recorder.Finish`.
- Start rule: buffer the `H264DecoderConfig` and an optional `AACDecoderConfig`, and drop media before the first `H264` keyframe. At that keyframe: `ProbeAVCConfig`, then `BuildLadder`, `NewClock(Now(), pkt.Time)`, `NewStream`, `Registry.Add`, `NewRecorder` (with `dir` = `Transcode.PrivateHLSPath`), `NewThumbnailer`. If any rendition is encoded, also `StartEncoder` and one `FeedTS` goroutine per output. The encoder gets the config packets first, then every packet, through a `chan av.Packet` of size `EncoderQueuePackets`. A full queue drops packets until the next keyframe and logs once per drop run. When the encoder exits, the error is logged and the source keeps going.
- A later `H264DecoderConfig` with a different width or height logs `source resolution changed` and is otherwise ignored.

- [ ] **Step 1: Write failing tests.**
  - `TestSessionSourceOnly`: `src_640x360.flv` with 1080/720/480 qualities → the ladder is `[source]`, no encoder starts, and the master lists only `source/` (**Review Focus 1**).
  - `TestSessionFullLadder` (`requireFFmpeg`): `src_1280x720.flv` → the master lists `source/` and `480p/`.
  - `TestSessionEncoderMissing`: `FFMpegPath: "/nonexistent"` with `src_1280x720.flv` → `HandlePacket` never returns an error and `source` becomes ready (**Review Focus 3**).
  - `TestSessionEncoderStalls`: replace ffmpeg with `sh -c 'sleep 30'` (it reads nothing) through `FFMpegPath` pointing at a temp script → feeding all packets takes under 1s in total (**Review Focus 3**).
  - `TestSessionCloseThenReconnect`: Close → GET master for the id → 404, and the VOD master exists. A new Session with a new id works (**Review Focus 5**).
- [ ] **Step 2: Run them.** Expected: FAIL. **Step 3: Implement.** **Step 4: Run with `-race`.** Expected: PASS.
- [ ] **Step 5: Commit** `feat(transcode): live session wiring source, encoder, recorder, thumbnails`.

### Task 10: Wire into the service, remove the old path

**Files:**
- Modify: `backend/transcode/rtmp/rtmp.go` (`HandleConnection`, `onConnect`, `onDisconnect`, `RTMPServerConfig`)
- Modify: `backend/transcode/cmd/main.go`, `backend/transcode/webserver/web-server.go`
- Delete: `backend/transcode/transcoder/ffmpeg.go`, `watcher/watcher.go`, `watcher/watcher_utils.go`, `watcher/vod.go`, `watcher/minio/` (first confirm with `grep -rn "watcher\." backend/transcode` that only these files and main.go use them)

**Interfaces:**
- Consumes: `live.NewRegistry`, `live.NewSession`, `live.SessionDeps`.
- `RTMPServerConfig` drops `VODHandler` and gains `Registry *live.Registry` and `Storage storage.Storage`.
- `NewWebServer(port, base string, live http.Handler)`: `/static/vods/` → FileServer(base) with prefix `/static/`; `/static/{id}/...` → `http.StripPrefix("/static", live)`; `/v1/health` stays. `WriteTimeout` becomes 30s, so held playlist requests are not cut off.

- [ ] **Step 1: Make the edits.** `HandleConnection` creates a Session after `onConnect`, calls `HandlePacket` for each `c.ReadPacket()`, and calls `Close` on EOF or error before `onDisconnect`. Keep the duration and livestream gateway calls as they are; remove `removeLiveGeneratedFiles` for the private path.
- [ ] **Step 2: Run `go mod tidy && go vet ./... && go test ./... -race -count=1 && go build ./cmd/`.** Expected: all pass; `radovskyb/watcher` is gone from go.mod.
- [ ] **Step 3: Run `docker build -f backend/transcode/Dockerfile backend`.** Expected: success (alpine `ffmpeg` already installed).
- [ ] **Step 4: Commit** `refactor(transcode): serve live via LL-HLS packager, remove disk watcher`.

### Task 11: Config repo and end-to-end check

**Files:**
- Config is already done: `../letslive-configs` branch `feat/ll-hls-live-config` adds `transcode.live` to `transcode_service-dev.yml` and `-prod.yml`. Existing qualities are `640x360`, `1280x720` and `1920x1080`, so they pass `Validate`. `hlsTime` stays (the uploaded-video transcoder uses it).

- [ ] **Step 1: Point the dev config server at `feat/ll-hls-live-config` (or merge it) and restart the config server.**
- [ ] **Step 2: Run `docker compose -f docker-compose-dev.yaml up --build transcode` plus its dependencies.** Expected log: `transcode` starts with no config validation error.
- [ ] **Step 3: Stream with OBS at 1280x720, keyframe interval 2s, x264.** `curl -s {API}/transcode/{id}/index.m3u8` lists `source/` and `480p/` (plus `720p/` only if the source is 1080p). `curl -s {API}/transcode/{id}/source/video1_stream.m3u8` contains `#EXT-X-PART:`.
- [ ] **Step 4: Measure latency.** Show an on-screen clock in OBS (Text source with a time script, or a phone stopwatch filmed by the camera), open the web player, take a screenshot showing both, and record the difference. Expected: ≤ 4s. Repeat with OBS at 854x480 and confirm only `source/` is offered.
- [ ] **Step 5: Compare CPU.** Record `docker stats transcode` (60s average) for the same 720p OBS stream on `main` and on this branch. Record both numbers in the PR description.
- [ ] **Step 6: End the stream.** The VOD plays from the livestream page, and `thumbnail.jpg` exists in MinIO under the id.
