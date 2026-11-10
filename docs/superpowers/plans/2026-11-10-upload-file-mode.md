# Uploads on the live engine (file mode) Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Run uploaded videos through `backend/transcode/live`, so an upload
comes out exactly like a recorded livestream, and delete the old upload
transcoder.

**Architecture:** `Session` gets a `Mode` value: a struct of strategies built
by `LiveMode(...)` or `FileMode(...)`. The session, encoder, recorder and
thumbnailer call those strategies and never branch on the mode. Live
strategies are today's code, moved behind the strategies. A new `upload`
package probes the file and runs one ffmpeg that turns it into H.264/AAC FLV
on a pipe. Packets from the joy5 FLV demuxer are fed to a file-mode `Session`.

**Tech Stack:** Go (go.work), ffmpeg/ffprobe, joy5 FLV, mediacommon v2,
gohlslib v2.4.7, MinIO storage interface.

**Spec:** `docs/superpowers/specs/2026-11-09-upload-file-mode-design.md`

## Global Constraints

- **No mode branches.** No `if mode == …`, no `isFile` / `isLive` bool, no
  type switch on the mode anywhere in `live/`. Differences live only in the
  `Mode` strategies (spec §2).
- **Live behaviour and output are unchanged.** Every existing test in `live/`
  and `rtmp/` keeps its assertions; only constructor wiring may change.
- Uploads are capped at **1080** short side (portrait included).
- The max keyframe gap for copying is **2 × `recordSegmentDurationMs`**.
  Converted output gets keyframes at most every **2 s**.
- File mode uses normal x264 tuning: no `-tune zerolatency`, no
  `-probesize 32 -analyzeduration 0`, no `-flush_packets 1`. Preset and CRF
  come from `ffmpegSetting`.
- HDR/10-bit input is tone-mapped to SDR BT.709 with `zscale` + `tonemap`.
- New uploads are played from `{vodPlaybackUrlPrefix}/{vodId}/index.m3u8`.
  Old uploads are left as they are.
- No new Go module dependency.
- Commits: `Refs: docs/superpowers/plans/2026-11-10-upload-file-mode.md`, and
  each ends with `Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>`.
  Never use `--no-verify`. Do not push.
- Run from `backend/transcode`: `gofmt -l .` must print nothing, and
  `go vet ./...` and `go test -race ./...` must pass at the end of every task.

## Review Focus

1. **A file faster than the encoder** (every upload). Nothing may be
   dropped; the encoder and recorder queues wait. Test: Task 2,
   `TestFileModeNeverDrops` with a 2-packet queue.
2. **The tail of encoded renditions at Close.** Draining must keep the last
   GOP of the 360p/720p renditions, or the VOD is shorter than the source.
   Test: Task 3, `TestDrainKeepsEncodedTail` checks that encoded and source
   durations differ by at most one segment.
3. **A phone video: portrait, rotated, VFR, HEVC, sometimes HDR.** It must
   come out upright, at a fixed frame rate, H.264, and not washed out.
   Tests: Task 7 e2e cases `rotated_portrait`, `hevc`, `hdr` (`hdr` skips
   when local ffmpeg lacks `zscale`).
4. **A live stream running while an upload encodes.** The upload's ffmpeg
   processes run niced (`upload.niceness`, default 10). Test: Task 3,
   `TestStartEncoderAppliesNiceness` reads the priority back with
   `syscall.Getpriority`.
5. **A failed attempt leaves a half-written VOD.** A retry must start clean:
   the recorder removes its local tree (it does today), and the worker
   removes `public/vods/{vodId}` before each attempt. Test: Task 8,
   `TestPrepareAttemptRemovesOldVODDir`.

---

### Task 1: `Mode`, `Output` and labels (live only, no behaviour change)

**Files:**
- Create: `live/mode.go`, `live/output.go`, `live/output_test.go`
- Modify: `live/session.go`, `live/session_config.go`,
  `live/session_encoder.go`, `live/feeder.go`, `live/recorder.go`,
  `live/thumbnail.go`, `live/encoder.go`, `live/stream.go`,
  `rtmp/rtmp.go`, and the test rigs in `live/session_test.go` and
  `live/feeder_test.go`

**Interfaces:**
- Produces:
  - `type Mode struct` (unexported fields; one value per Session).
  - `func LiveMode(reg *Registry, cfg config.Live) (Mode, error)`: returns
    an error when `reg` is nil.
  - `func (m Mode) label(id string) string`: `"livestream "+id`.
  - `func (m Mode) open(id string, rs []Rendition, audio *mpeg4audio.AudioSpecificConfig) (Output, error)`
  - `type Output interface { WriteVideo(r int, ntp time.Time, pts90k int64, au [][]byte) error; WriteAudio(ntp time.Time, pts int64, aus [][]byte) error; Ready(r int) bool; DropRendition(r int); Unpublish(); Close() }`
  - `liveOutput{*Stream; reg *Registry}`, where `Unpublish` calls
    `reg.Remove(id)`.
  - `noOutput struct{}`: every write returns nil, `Ready` returns true, the
    rest do nothing.
  - `SessionDeps{Mode Mode; Storage storage.Storage; Transcode config.Transcode; Now func() time.Time}`.
    `Registry` moves into `LiveMode`.
  - `FeedTS(ctx context.Context, r io.Reader, out Output, rendition int, name, label string, clk *Clock, onAU func(pts90k, dts90k int64, au [][]byte)) error`.
    The range check moves to the caller, which knows the ladder.
  - `NewRecorder` and `NewThumbnailer` take `label string` in place of
    building `"livestream "+id` themselves. `id` is still used for storage
    paths.

- [ ] **Step 1: Write the failing test** in `live/output_test.go`:
  - `TestNoOutputAcceptsEverything`: writes return nil and `Ready(0..3)` is
    true.
  - `TestLiveModeNeedsRegistry`: `LiveMode(nil, cfg)` returns an error.
  - `TestLiveModeOpenRegistersAndUnpublishRemoves`: after `open`, the
    registry serves `/{id}/index.m3u8` (not 404). After `Unpublish` it is
    404.
  - `TestLiveModeOpenDuplicateID`: a second `open` with the same id returns
    an error.
- [ ] **Step 2:** Run `go test ./live -run 'NoOutput|LiveMode' -v`. Expected:
  FAIL (undefined: `LiveMode`).
- [ ] **Step 3: Implement.**
  - `Session` keeps an `out Output` and a `label string` in place of
    `stream *Stream`.
  - `start` calls `s.d.Mode.open`; `beginClose` calls `s.out.Unpublish()`.
  - `shutdown` calls `s.out.Close()`.
  - `encoderWiring.stream` becomes `out Output`.
  - Replace every `"livestream %s` log prefix in `live/` with `"%s`,
    passing the label. Do not change the wording after the prefix, because
    tests match on it.
  - `rtmp/rtmp.go` builds `live.LiveMode(s.liveRegistry, s.config.Transcode.Live)`.
    An error from it is logged and the connection closed, as with
    `onConnect` errors.
  - The test rigs pass `Mode: mustLiveMode(t, reg, tc.Live)`.
- [ ] **Step 4:** Run `go test -race ./...`. Expected: all PASS, including
  every existing session/feeder/recorder/thumbnail test, with no assertion
  edits. Then run `grep -rn '"livestream ' live/*.go | grep -v _test`.
  Expected: matches only in `mode.go`.
- [ ] **Step 5: Commit** `refactor(transcode): put the live stream behind an Output chosen by the session mode`.

### Task 2: `Backpressure` and `Failures`

**Files:**
- Create: `live/policy.go`, `live/policy_test.go`
- Modify: `live/mode.go`, `live/session.go`, `live/session_config.go`,
  `live/session_encoder.go`, `live/recorder.go`, `live/feeder.go`
- Test: `live/session_test.go` (new tests only)

**Interfaces:**
- Consumes: `Mode` (Task 1).
- Produces:
  - `type Backpressure interface { giveUp() <-chan struct{} }`.
    `dropWhenFull{}` returns a package-level channel that is already
    closed. `waitUntil{ctx}` returns `ctx.Done()`.
  - `func send[T any](bp Backpressure, ch chan<- T, v T) bool`: a
    non-blocking try first, then `select` on `ch <- v` and `bp.giveUp()`.
  - `type Failures interface { Report(err error); Err() error }`.
    `logOnly{}` does nothing and `Err` returns nil (the call sites keep
    their own logging). `firstError{cancel context.CancelFunc}` is
    mutex-guarded: it keeps the first error and calls `cancel`.
  - `Mode` fields: `bp Backpressure`, `fail Failures`.
  - Test-only constructor `testMode(out Output, bp Backpressure, f Failures) Mode`
    in a `_test.go` file.
  - `Session.HandlePacket` returns `s.d.Mode.fail.Err()` (wrapped with the
    label) once it is non-nil.
- Call sites: `sessionEncoder.offer`, `offerConfig` and `flushPending` use
  `send`, and the recorder's `enqueue` uses `send`. `Report` is called at:
  - `sessionEncoder.fail`, and the `disable` path in `startSessionEncoder`;
  - `Recorder.DropRendition`;
  - every `settle(seg, false, …)` and `abandonSegment`;
  - the "abandoned unfinished" branch of `logSummary`;
  - a full queue in `enqueue` (only reachable once `giveUp` fires);
  - the `sourceAU` error in `Session.writeVideo`;
  - `onVideoConfig` returning `configInvalid` or `configResized`;
  - `readTS`'s `OnDecodeError`.

  Each `Report` call is placed next to the existing log line; the log line
  stays.

- [ ] **Step 1: Write the failing tests.**
  - `TestSendDropWhenFull`: a full chan of cap 1 returns false at once.
  - `TestSendWaitUntilRoom`: a full chan returns true after a reader takes
    one item.
  - `TestSendWaitGivesUpOnCancel`: returns false within 100 ms of
    `cancel()`.
  - `TestFirstErrorKeepsFirstAndCancels`.
  - `TestFileModeNeverDrops`: a session with
    `testMode(noOutput{}, waitUntil{ctx}, firstError)` and
    `EncoderQueuePackets = 2` is fed all of `src_1280x720.flv`. Assert
    `rec.Dropped() == 0`, no "encoder queue full" log line, and the
    encoded-rendition VOD playlists exist.
  - `TestFailuresReportedOnEncoderFailure`: with
    `FFMpegPath = "/nonexistent"`, `Err()` is non-nil.
  - `TestFailuresReportedOnSegmentUploadFailure`: `recordingStorage`
    returns an error from `AddSegment`.
  - `TestFailuresReportedOnResize`: reuse the resize fixture from
    `TestSessionAVCConfigResolutionChange`.
  - `TestHandlePacketReturnsReportedFailure`.
- [ ] **Step 2:** Run `go test ./live -run 'Send|FirstError|FileModeNeverDrops|FailuresReported|ReportedFailure' -v`.
  Expected: FAIL (undefined).
- [ ] **Step 3: Implement** as in Interfaces. `LiveMode` sets
  `dropWhenFull{}` and `logOnly{}`.
- [ ] **Step 4:** Run `go test -race ./...`. Expected: PASS, including
  `TestSessionEncoderQueueDropsUntilKeyframe` and
  `TestSessionEncoderStalls` unchanged.
- [ ] **Step 5: Commit** `feat(transcode): queue backpressure and failure reporting as session strategies`.

### Task 3: `EncoderProfile`, draining stop, niceness

**Files:**
- Modify: `live/encoder_args.go`, `live/encoder.go`,
  `live/session_encoder.go`, `live/mode.go`, `config/live.go` (add
  `Upload`), `config/config.go`
- Test: `live/encoder_test.go`, `live/session_test.go`,
  `config/live_test.go`

**Interfaces:**
- Produces:
  - `type EncoderProfile struct { Tune []string; Input []string; Output []string; Threads int; Niceness int }`.
    `liveProfile` holds today's exact flags:
    - `Tune`: `-tune zerolatency`;
    - `Input`: `-probesize 32 -analyzeduration 0`;
    - `Output`: `-muxdelay 0 -muxpreload 0 -flush_packets 1`.

    `fileProfile(u config.Upload)` sets `Tune` and `Input` to nil, keeps
    `-muxdelay 0 -muxpreload 0` in `Output`, and sets
    `Threads: u.Threads, Niceness: u.Niceness`.
  - `EncoderArgs(src SourceInfo, encoded []Rendition, preset string, crf int, p EncoderProfile) []string`.
    It adds `-threads N` per output only when `p.Threads > 0`. That is a
    value check, not a mode check.
  - `StartEncoder(..., niceness int)`. After `cmd.Start` it calls
    `syscall.Setpriority(syscall.PRIO_PROCESS, pid, niceness)` when
    `niceness != 0`; a failure is logged, not returned.
  - `type encoderEnd func(e *sessionEncoder, ctx context.Context)`.
    `terminate` is today's `stop` body. `drain`:
    1. sets `closing` and closes a new `inputDone` channel;
    2. `writeLoop` writes what is left in the queue, then closes ffmpeg's
       stdin;
    3. waits for `enc.Wait()` and the feeders (they return on EOF);
    4. if ctx ends first, falls back to `terminate`.
  - `Mode` fields: `profile EncoderProfile`, `endEncoder encoderEnd`.
    `Session.shutdown` calls `s.d.Mode.endEncoder(s.enc, ctx)` with a
    context bounded by the Mode's `finishTimeout` (added to `Mode` here:
    live 60 s, today's `recorderFinishTimeout`; `Recorder.Finish` reads it
    from a new `finishTimeout time.Duration` parameter).
  - `config.Upload{Niceness int \`yaml:"niceness"\`; Threads int \`yaml:"threads"\`}`
    on `config.Transcode` as `Upload Upload \`yaml:"upload"\``.
    `WithDefaults` sets Niceness to 10; Threads 0 means ffmpeg decides.
- [ ] **Step 1: Write the failing tests.**
  - `TestEncoderArgsLiveProfileUnchanged`: the existing golden args, built
    through `liveProfile`, are byte-identical.
  - `TestEncoderArgsFileProfile`: no `zerolatency`, `probesize` or
    `flush_packets`; `-threads 2` when `Threads = 2`.
  - `TestStartEncoderAppliesNiceness`: `requireFFmpeg`; start with
    niceness 5, then `syscall.Getpriority` on the pid returns 5 (macOS
    returns the nice value; on Linux the syscall returns `20 - nice`, so
    compare via a helper).
  - `TestDrainKeepsEncodedTail`: a `testMode` with `drain` is fed
    `src_1280x720.flv`, then closed. Each encoded rendition's playlist
    `#EXTINF` total is within `recordSegmentDurationMs` of the source
    playlist total.
  - `TestUploadDefaults`: `Upload{}.WithDefaults().Niceness == 10`.
- [ ] **Step 2:** Run `go test ./live ./config -run 'EncoderArgs|Niceness|Drain|UploadDefaults' -v`.
  Expected: FAIL.
- [ ] **Step 3: Implement** as in Interfaces. `LiveMode` uses
  `liveProfile` and `terminate`.
- [ ] **Step 4:** Run `go test -race ./...`. Expected: PASS.
- [ ] **Step 5: Commit** `feat(transcode): encoder profile and end strategy, drain for files`.

### Task 4: `ThumbSchedule` and the thumbnail URL

**Files:**
- Modify: `live/thumbnail.go`, `live/session.go`, `live/mode.go`
- Test: `live/thumbnail_test.go`

**Interfaces:**
- Produces:
  - `type ThumbSchedule interface { Want(media time.Duration) bool; KeepFirst() bool }`.
    `Want` is called only when the thumbnailer is idle, and a true result
    is recorded as taken.
  - `wallSchedule{now func() time.Time; notBefore time.Time; every time.Duration; last time.Time; taken bool}`
    is today's rule moved out of `Offer`: the first thumbnail after 5 s,
    then every `every`. It ignores `media` and `KeepFirst` is false.
  - `mediaSchedule{min time.Duration; taken bool}` takes the first keyframe
    with `media >= 5s` once. `KeepFirst` is true.
  - `Thumbnailer.Offer(au [][]byte, media time.Duration)`. When
    `sched.KeepFirst()` and nothing has been kept yet, a copy of the first
    offered access unit is kept.
  - `Thumbnailer.Close()`: if no job ever ran and a first access unit was
    kept, it runs that one and waits for it.
  - `Thumbnailer.URL() string` returns the last URL returned by
    `AddThumbnail` (`""` if none).
  - `Session.ThumbnailURL() string` is valid after `Close`.
  - `Mode` field: `thumbs func(now func() time.Time, every time.Duration) ThumbSchedule`.
  - The Session passes `pkt.Time - s.clk.zero` as `media`.
- [ ] **Step 1: Write the failing tests.**
  - `TestMediaScheduleTakesFirstKeyframeAfter5s`: `Want(0)`, `Want(4s)`,
    then `Want(6s)`; only `6s` is true; `Want(70s)` is false.
  - `TestThumbnailerShortVideoUsesFirstKeyframe`: with a media schedule,
    offer one keyframe at `media = 1s`, then `Close`. Exactly one
    `AddThumbnail` call, and `URL()` equals its return value.
  - `TestThumbnailerURLEmptyWhenUploadFails`.
- [ ] **Step 2:** Run `go test ./live -run 'MediaSchedule|ShortVideo|URLEmpty' -v`.
  Expected: FAIL.
- [ ] **Step 3: Implement.** `LiveMode` returns a `wallSchedule` built from
  `now`. Existing thumbnail tests that set `notBefore` now set it on the
  `wallSchedule`.
- [ ] **Step 4:** Run `go test -race ./...`. Expected: PASS.
- [ ] **Step 5: Commit** `feat(transcode): thumbnail schedule strategy and thumbnail URL`.

### Task 5: `FileMode`

**Files:**
- Modify: `live/mode.go`
- Test: `live/session_file_test.go` (new)

**Interfaces:**
- Consumes: Tasks 1–4.
- Produces:
  `func FileMode(ctx context.Context, u config.Upload) (m Mode, cancel context.CancelFunc)`
  sets:
  - label `"upload "+id`;
  - `open` returning `noOutput{}`;
  - `waitUntil{fctx}`, `firstError{cancel}`;
  - `fileProfile(u)`, `drain`;
  - `mediaSchedule{min: 5s}`;
  - `finishTimeout` = 10 min (`fileFinishTimeout`).

  `fctx` is derived from `ctx`; the caller defers `cancel`.

- [ ] **Step 1: Write the failing tests**, using `FileMode` and
  `src_1280x720.flv`:
  - `TestFileSessionWritesVODLikeLive`: the master lists ladder indexes
    `0..n` in the same order and with the same `CODECS` as a live session
    fed the same fixture. Compare the two master files with segment URIs
    stripped.
  - `TestFileSessionNotRegistered`: run a live registry beside it; the id
    serves 404 the whole time.
  - `TestFileSessionCloseFailsOnEncoderFailure`
  - `TestFileSessionReturnsThumbnailURL`
  - `TestFileSessionLogsUploadLabel`: a log line starts `upload <id>:`.
- [ ] **Step 2:** Run `go test ./live -run FileSession -v`. Expected: FAIL
  (undefined `FileMode`).
- [ ] **Step 3:** Implement `FileMode`.
- [ ] **Step 4:** Run `go test -race ./...`, then run
  `grep -rnE 'isFile|isLive|mode ==' live/*.go`. Expected: PASS, and no
  grep matches.
- [ ] **Step 5: Commit** `feat(transcode): file mode for the live session`.

### Task 6: Upload probe and copy-or-convert decision

**Files:**
- Create: `upload/probe.go`, `upload/plan.go`, `upload/args.go`,
  `upload/plan_test.go`, `upload/args_test.go`

**Interfaces:**
- Produces:
  - `type Probe struct { HasVideo bool; VideoCodec, Profile, PixFmt, FieldOrder, ColorTransfer string; Width, Height, Rotation int; FrameRate float64; MaxKeyframeGap time.Duration; HasAudio bool; AudioCodec string; AudioChannels int }`
  - `func RunProbe(ctx context.Context, ffprobePath, input string) (Probe, error)`
    makes two `ffprobe -of json` calls:
    1. `-show_streams` for `v:0` and `a:0`; rotation comes from the
       `side_data_list` display matrix `rotation` or the `tags.rotate`
       value.
    2. `-select_streams v:0 -skip_frame nokey -show_entries frame=pts_time`
       for the keyframe times.

    `ffprobePath` is `filepath.Join(filepath.Dir(FFMpegPath), "ffprobe")`
    when `FFMpegPath` has a directory, otherwise `"ffprobe"`.
  - `type Plan struct { CopyVideo, CopyAudio, HasAudio, ToneMap, Deinterlace bool; ScaleShort int; FPS int }`
  - `func Decide(p Probe, maxGap time.Duration, forceConvert bool) (Plan, error)`.
    It returns `ErrNoVideo` when `!p.HasVideo`. `CopyVideo` requires all of:
    codec `h264`, pix_fmt `yuv420p`/`yuvj420p`, a transfer that is not
    `smpte2084`/`arib-std-b67`, short side ≤ 1080, `Rotation % 360 == 0`,
    field order `progressive` or empty, `MaxKeyframeGap ≤ maxGap`, and
    `!forceConvert`. `ScaleShort` is 1080 when the short side is above
    1080, otherwise 0. `FPS` is `min(round(FrameRate), 60)`, or 30 when the
    rate is unknown.
  - `func IntakeArgs(input string, pl Plan, preset string, crf int, threads int) []string`
    builds an ffmpeg command that writes FLV to `pipe:1`:
    - start: `-hide_banner -loglevel error -i input`;
    - maps: `-map 0:v:0` and, when `HasAudio`, `-map 0:a:0`;
    - video copy: `-c:v copy`;
    - video convert:
      - `-vf` chain: `yadif` (when Deinterlace), then
        `zscale=t=linear:npl=100,format=gbrpf32le,zscale=p=bt709,tonemap=hable:desat=0,zscale=t=bt709:m=bt709:r=tv`
        (when ToneMap), then `scale=…:-2` / `-2:…` pinning the short side
        (when ScaleShort), then `format=yuv420p`;
      - `-c:v libx264 -preset P -crf C -fps_mode cfr -r FPS -force_key_frames expr:gte(t,n_forced*2)`.

      Auto-rotation is ffmpeg's default.
    - audio: `-c:a copy` when `CopyAudio`; otherwise
      `-c:a aac -b:a 128k`, plus `-ac 2` when there are more than 2
      channels;
    - `-threads N` when threads > 0;
    - end: `-avoid_negative_ts make_zero -flvflags no_duration_filesize -f flv pipe:1`.
- [ ] **Step 1: Write the failing tests.**
  - `TestDecide` is a table over Probe values with one row per spec rule:
    - copy: H.264 1080p progressive;
    - convert: HEVC, 10-bit `yuv420p10le`, HLG, rotated 90, interlaced
      `tt`, short side 1440 (`ScaleShort` 1080), and a keyframe gap of 9 s
      with `maxGap` 8 s;
    - portrait 1080x1920 copies;
    - `forceConvert` converts;
    - no video returns `ErrNoVideo`;
    - Opus audio has `CopyAudio` false;
    - no audio has `HasAudio` false.
  - `TestIntakeArgs` checks the exact arg slices for the copy plan, the
    full-convert plan (tonemap + deinterlace + scale), and the 6-channel
    audio case.
  - `TestParseProbe` parses JSON fixtures saved from real ffprobe output
    (generate them in Step 3 with the Task 7 generator commands, and store
    them in `upload/testdata/*.json`) into the expected `Probe`, including
    rotation from the display matrix.
- [ ] **Step 2:** Run `go test ./upload -v`. Expected: FAIL.
- [ ] **Step 3: Implement** as in Interfaces.
  - JSON parsing is `parseStreams([]byte) (Probe, error)` and
    `parseKeyframes([]byte) (time.Duration, error)`.
  - `RunProbe` calls them.
- [ ] **Step 4:** Run `go test -race ./upload`. Expected: PASS.
- [ ] **Step 5: Commit** `feat(transcode): probe uploads and decide copy or convert`.

### Task 7: `upload.Run`, file to FLV pipe to file-mode session

**Files:**
- Create: `upload/run.go`, `upload/run_test.go`, `upload/gen_test.go`
  (test-file generator helpers)

**Interfaces:**
- Consumes: `live.NewSession`, `live.FileMode`, `live.SessionDeps`,
  `Session.Duration`, `Session.ThumbnailURL`, `Probe`/`Decide`/`IntakeArgs`.
- Produces:
  - `type Result struct { Duration time.Duration; ThumbnailURL string }`
  - `var ErrInput = errors.New("upload input")`. Input-caused failures wrap
    it: `ErrNoVideo`, intake ffmpeg failing on a copy plan, `HandlePacket`
    errors, and a `Close` error once intake had started. The worker uses it
    for retry-with-conversion.
  - `func Run(ctx context.Context, tc config.Transcode, st storage.Storage, vodID, input string, forceConvert bool) (Result, error)`.
    1. `RunProbe`, `Decide(maxGap = 2 × RecordSegmentDurationMs)` and
       `IntakeArgs`.
    2. Start ffmpeg with stdout piped and niceness applied (reuse a small
       exported `live.SetNiceness(pid, n)`, factored out of Task 3).
    3. Read packets from `flv.NewDemuxer(stdout)` into
       `sess.HandlePacket` until EOF or error.
    4. Run `cmd.Wait()`, then `sess.Close(ctx)`.
    5. Any error comes back with the ffmpeg stderr tail (last 4 KiB).

    It never returns success when ffmpeg exited non-zero or `Close`
    failed.
- [ ] **Step 1: Write the failing tests.** All of them `requireFFmpeg`.
  Inputs are made in `t.TempDir()` with `ffmpeg -f lavfi` (`testsrc2` +
  `sine`, 12 s) and run against `recordingStorage`-like fake storage plus
  a temp `PublicHLSPath`:
  - `h264_aac_mp4` (720p): copied (assert from a debug log line
    `copy video`). Ladder `[720p, 360p]`, `CODECS` present,
    duration 12 s ± 1.
  - `hevc` (`libx265` if available, else skip): converted; the output
    rendition 0 playlist's first segment probes as h264.
  - `vp9_opus_webm`: converted, and the audio track exists in segment 0
    (ffprobe the segment).
  - `rotated_portrait`: 1280x720 with `-display_rotation 90`. Rendition 0
    segment probes as 720x1280.
  - `big` (2560x1440): rendition 0 is short side 1080.
  - `no_audio`: VOD written and segments have no audio stream.
  - `no_video` (audio only): returns an error wrapping `ErrInput` and
    `ErrNoVideo`.
  - `hdr` (`-color_trc arib-std-b67 -pix_fmt yuv420p10le` via `libx265`):
    skips unless `ffmpeg -filters` lists `zscale`. Rendition 0 is 8-bit
    `yuv420p` with a `bt709` transfer.
  - Every success case asserts `Result.ThumbnailURL != ""` and that the
    master is at `{PublicHLSPath}/vods/{vodID}/index.m3u8`.
- [ ] **Step 2:** Run `go test ./upload -run Run -v`. Expected: FAIL.
- [ ] **Step 3:** Implement `Run`.
- [ ] **Step 4:** Run `go test -race ./upload -run Run -v`. Expected: PASS
  (`hdr` may SKIP locally; record which).
- [ ] **Step 5: Commit** `feat(transcode): run uploads through the live engine`.

### Task 8: Worker switch, retry with conversion, delete the old transcoder

**Files:**
- Modify: `worker/worker.go`, `cmd/*` (if it wires config), the transcode
  config YAMLs under `configs/` (add an `upload:` block with defaults
  commented)
- Create: `worker/worker_test.go`
- Delete: `transcoder/file_transcoder.go` (and the package dir)
- Docs: append a "Results" section to the spec, and record the change in
  `docs/ISSUES.md` if any entry is affected

**Interfaces:**
- Consumes: `upload.Run`, `upload.ErrInput`.
- Produces:
  - The claim query also selects `tj.error_message`.
  - `func forceConvert(prevErr *string) bool`: true when the previous error
    starts with `inputErrPrefix = "input: "`.
  - `markJobFailed` messages for `ErrInput` errors are prefixed with
    `inputErrPrefix`.
  - `func prepareAttempt(publicHLSPath, vodID string) error` removes
    `{publicHLSPath}/vods/{vodID}`.
  - `doTranscode` flow:
    1. download (unchanged);
    2. `prepareAttempt`;
    3. `upload.Run`;
    4. `UpdateVODStatus(ready, fmt.Sprintf("%s/%s/index.m3u8", VODPlaybackUrlPrefix, vodId), res.ThumbnailURL, &ceilSeconds)`;
    5. mark completed, delete the raw object (unchanged).
  - Deleted: `uploadHLSToStorage`, `getHLSDurationSeconds`.
- [ ] **Step 1: Benchmark before deleting** (spec "Testing"). Write a
  scratch program in the session scratchpad, not the repo. It runs the old
  `transcoder.TranscodeFile` and the new `upload.Run` on the same 60 s
  1080p H.264/AAC file and on a 60 s HEVC file. For each it records the
  children's CPU seconds (`syscall.Getrusage(RUSAGE_CHILDREN)` deltas),
  wall time and output bytes. Put the table in the spec under "Results
  (2026-11-10)".
- [ ] **Step 2: Write the failing tests.**
  - `TestForceConvert`: nil → false; `"input: …"` → true;
    `"failed to download…"` → false.
  - `TestPrepareAttemptRemovesOldVODDir`: the dir with a stale master is
    gone; a missing dir is not an error.
  - `TestInputErrorMessagePrefixed`: the message helper given a wrapped
    `upload.ErrInput` starts with `input: `.
- [ ] **Step 3:** Run `go test ./worker -v`. Expected: FAIL.
- [ ] **Step 4: Implement** as in Interfaces. Delete the old code, then run
  `grep -rn "TranscodeFile\|transcoder\"" --include='*.go' .`. Expected: no
  match.
- [ ] **Step 5:** Run `gofmt -l .`, `go vet ./...` and
  `go test -race ./...`, and build the image's binary with
  `go build ./cmd/...`. Expected: clean, PASS, builds.
- [ ] **Step 6: End-to-end check in the running stack** (if docker compose
  is up locally):
  1. Upload a phone-style portrait MP4 through the web upload.
  2. Watch the VOD go `processing` → `ready`.
  3. `curl` its `/transcode/vods/{id}/index.m3u8` and get a 200 with
     `CODECS`.
  4. Open it in the web player and confirm it plays upright.

  If the stack is not up, say so in the report; do not claim it.
- [ ] **Step 7: Commit** `feat(transcode): uploads use the live engine; drop the old file transcoder`.
