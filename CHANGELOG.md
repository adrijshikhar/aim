# Changelog

## [0.18.1](https://github.com/adrijshikhar/aim/compare/v0.18.0...v0.18.1) (2026-10-24)


### Bug Fixes

* **agy:** bridge agstatusline and merge host statusLine into profile settings ([5435ec4](https://github.com/adrijshikhar/aim/commit/5435ec485d61efffcd92d9b8e4b096e9dbc41a1e))

## [0.18.0](https://github.com/adrijshikhar/aim/compare/v0.17.0...v0.18.0) (2026-10-23)


### Features

* **web:** complete design audit, token standardization, and UI invariants ([#101](https://github.com/adrijshikhar/aim/issues/101)) ([caf95fa](https://github.com/adrijshikhar/aim/commit/caf95faa3e050bf5012788e87eb7399c5b9357ae))

## [0.17.0](https://github.com/adrijshikhar/aim/compare/v0.16.0...v0.17.0) (2026-10-22)


### Features

* multi-agent interface segregation and Claude Code support ([#99](https://github.com/adrijshikhar/aim/issues/99)) ([04b2fb5](https://github.com/adrijshikhar/aim/commit/04b2fb535285ea5fc88de03591f0acaad0ef15ea))
* peer-to-peer profile sharing and credential leasing ([#98](https://github.com/adrijshikhar/aim/issues/98)) ([19498fe](https://github.com/adrijshikhar/aim/commit/19498feadbce2ba9e4f8809f519f063205457cf5))

## [0.16.0](https://github.com/adrijshikhar/aim/compare/v0.15.0...v0.16.0) (2026-10-17)


### Features

* **cli:** add aim update command with Homebrew and standalone auto-update support ([#96](https://github.com/adrijshikhar/aim/issues/96)) ([28623d0](https://github.com/adrijshikhar/aim/commit/28623d06d080aa0f657347238af1856965a216ca))
* **config:** add aim config command, support telemetry unmarshaling, and standardize Lipgloss tables ([#97](https://github.com/adrijshikhar/aim/issues/97)) ([c71eed6](https://github.com/adrijshikhar/aim/commit/c71eed6f129c44db702aa3a767bf6033f2e76490))
* **ui:** standardize Lipgloss tables via presenter.NewTable and add strict AGENTS.md guidelines ([#95](https://github.com/adrijshikhar/aim/issues/95)) ([36f9037](https://github.com/adrijshikhar/aim/commit/36f90378ff350c8c65bafed18809b3baa68b02c1))


### Bug Fixes

* **cli:** use native Lipgloss table width and truncation to prevent terminal wrapping ([#93](https://github.com/adrijshikhar/aim/issues/93)) ([a1ba6b8](https://github.com/adrijshikhar/aim/commit/a1ba6b8d84a366088f3874ed01a86145d2469264))

## [0.15.0](https://github.com/adrijshikhar/aim/compare/v0.14.0...v0.15.0) (2026-10-15)


### Features

* **mcp:** expanded MCP server tools, protocol doctor, and client management suite ([#91](https://github.com/adrijshikhar/aim/issues/91)) ([5ac89ab](https://github.com/adrijshikhar/aim/commit/5ac89aba95f3f3c2dc6def44c50d6bf931e52436))

## [0.14.0](https://github.com/adrijshikhar/aim/compare/v0.13.0...v0.14.0) (2026-10-15)


### Features

* **inspection:** strict read-only profile and candidate inspection ([#89](https://github.com/adrijshikhar/aim/issues/89)) ([1c39c97](https://github.com/adrijshikhar/aim/commit/1c39c97b70946e0683e60fec243bf363453a0e13))
* **oauth:** granular OAuth refusal diagnostics and 24h retry backoff ([#88](https://github.com/adrijshikhar/aim/issues/88)) ([9aada2a](https://github.com/adrijshikhar/aim/commit/9aada2ad0055ff9d0067f49f29637dec620a0a22))
* **profile:** per-directory profile binding with .aim-profile and aim bind ([#85](https://github.com/adrijshikhar/aim/issues/85)) ([7dd192c](https://github.com/adrijshikhar/aim/commit/7dd192c89a007cf73e12ae77808b69653476335e))
* **web:** add embedded web dashboard with multi-adapter profiles, sessions, and MCP management ([#80](https://github.com/adrijshikhar/aim/issues/80)) ([9a0dd63](https://github.com/adrijshikhar/aim/commit/9a0dd63ef34f102b24d7b7e529b4d68473746fb4))
* **web:** add faceted CWD directory filter to sessions dashboard ([#90](https://github.com/adrijshikhar/aim/issues/90)) ([965bb93](https://github.com/adrijshikhar/aim/commit/965bb9352d1ccde79ab02102fbab452208252174))
* **web:** daemonize web dashboard on port 18480, refine settings layout and quota discovery ([#87](https://github.com/adrijshikhar/aim/issues/87)) ([8649b09](https://github.com/adrijshikhar/aim/commit/8649b0985bd81654813b5ce5a39a441b10f6954a))


### Bug Fixes

* **profile:** sanitize profile binding name and fix comment parsing ([#86](https://github.com/adrijshikhar/aim/issues/86)) ([d1c02e5](https://github.com/adrijshikhar/aim/commit/d1c02e5d01645193d216ba11e25852e1c3e60f4e))

## [0.13.0](https://github.com/adrijshikhar/aim/compare/v0.12.0...v0.13.0) (2026-10-13)


### Features

* **daemon:** add native OS background service for 15-minute quota cache refresh ([#81](https://github.com/adrijshikhar/aim/issues/81)) ([affbe58](https://github.com/adrijshikhar/aim/commit/affbe583c4885d9287c6b65cdf2beadc324566a6))
* diagnostics, feedback reporting, auto-updater, and telemetry with runtime fixes ([#79](https://github.com/adrijshikhar/aim/issues/79)) ([56baa03](https://github.com/adrijshikhar/aim/commit/56baa0394d5bd27b77f02aa7a889643083289b90))
* **mcp:** add read-only stdio MCP server ('aim mcp serve') ([#83](https://github.com/adrijshikhar/aim/issues/83)) ([48e4bcc](https://github.com/adrijshikhar/aim/commit/48e4bcc9c7ceb06bf28c3dd69ade09b39feb2aae))
* **oauth:** crash-resilient unclaimed token stash store ([#82](https://github.com/adrijshikhar/aim/issues/82)) ([eb852c7](https://github.com/adrijshikhar/aim/commit/eb852c7f5aa29ab32e359308d378b812538093bd))
* **session:** add generic CWD resolution and rich session summaries across adapters ([#75](https://github.com/adrijshikhar/aim/issues/75)) ([51d6865](https://github.com/adrijshikhar/aim/commit/51d6865bb99025954673fd8f2d007d7c60c84cc9))


### Bug Fixes

* **usage:** prevent false offline badge during quota refresh and preserve cached quotas ([#77](https://github.com/adrijshikhar/aim/issues/77)) ([eeb7375](https://github.com/adrijshikhar/aim/commit/eeb737549b629b599b2b7a37a625463166293298))

## [0.12.0](https://github.com/adrijshikhar/aim/compare/v0.11.2...v0.12.0) (2026-10-02)


### Features

* **claude:** implement quota and usage reporting for Claude Code ([#73](https://github.com/adrijshikhar/aim/issues/73)) ([23bd74b](https://github.com/adrijshikhar/aim/commit/23bd74b7d705cb29991c1ea5c180d744b3433f2f))

## [0.11.2](https://github.com/adrijshikhar/aim/compare/v0.11.1...v0.11.2) (2026-10-01)


### Performance Improvements

* **usage:** parallelize agy quota subprocesses and enable stale-while-revalidate ([#71](https://github.com/adrijshikhar/aim/issues/71)) ([ffcca81](https://github.com/adrijshikhar/aim/commit/ffcca81d073bfd3099b65bf7b5a54b55ee2933d5))

## [0.11.1](https://github.com/adrijshikhar/aim/compare/v0.11.0...v0.11.1) (2026-10-01)


### Bug Fixes

* **tui:** isolate Claude tab profiles and extract native Claude authentication ([#69](https://github.com/adrijshikhar/aim/issues/69)) ([d8fbf02](https://github.com/adrijshikhar/aim/commit/d8fbf02801fa071dbbe9e071aaeae508d6a37b0e))

## [0.11.0](https://github.com/adrijshikhar/aim/compare/v0.10.1...v0.11.0) (2026-10-01)


### Features

* **mcp:** adapter-aware MCP list formatting, run interception, and Claude config seeding ([#66](https://github.com/adrijshikhar/aim/issues/66)) ([97df544](https://github.com/adrijshikhar/aim/commit/97df5444c027b20357d12eacf856412a14b108ac))


### Bug Fixes

* **build:** eliminate git describe hash suffixes from version output ([#68](https://github.com/adrijshikhar/aim/issues/68)) ([34d8afa](https://github.com/adrijshikhar/aim/commit/34d8afa23cf51dcdeab08fa4e4cc59ca52a36177))

## [0.10.1](https://github.com/adrijshikhar/aim/compare/v0.10.0...v0.10.1) (2026-10-01)


### Bug Fixes

* **agy:** preserve SSH_CONNECTION when valid refresh_token is present ([#62](https://github.com/adrijshikhar/aim/issues/62)) ([926b529](https://github.com/adrijshikhar/aim/commit/926b529abf9bce2ea5c34b74ce3c7c5d3e551eb5))
* **build:** inject version via ldflags; refresh README and docs ([#64](https://github.com/adrijshikhar/aim/issues/64)) ([8e5ef3a](https://github.com/adrijshikhar/aim/commit/8e5ef3a4bf28e9b9dff88171270f6222044a1a7f))

## [0.10.0](https://github.com/adrijshikhar/aim/compare/v0.9.1...v0.10.0) (2026-09-29)


### Features

* session-scoped merge of global MCP servers (Spec A v1) ([#58](https://github.com/adrijshikhar/aim/issues/58)) ([a6c7f93](https://github.com/adrijshikhar/aim/commit/a6c7f93776c9e856c5bb257951b8a89190b54ca3))
* session-scoped merge of plugin enablement (Spec B) + non-session skip ([#60](https://github.com/adrijshikhar/aim/issues/60)) ([1dd68af](https://github.com/adrijshikhar/aim/commit/1dd68af5c6a9ba3cccbd65a3cead35400747abc7))
* **tui:** declarative keymaps, dynamic help rendering, fish bridging & upstream canary watcher ([#61](https://github.com/adrijshikhar/aim/issues/61)) ([6189fa7](https://github.com/adrijshikhar/aim/commit/6189fa7cacfc41d0ae7acd6e767315147130398b))


### Bug Fixes

* **session:** size-aware cross-profile hydration, flag forwarding, and parameterized e2e suite ([#56](https://github.com/adrijshikhar/aim/issues/56)) ([22de12e](https://github.com/adrijshikhar/aim/commit/22de12e2c280ba09efade3eb227d4ffd0a21d3d5))

## [0.9.1](https://github.com/adrijshikhar/aim/compare/v0.9.0...v0.9.1) (2026-09-25)


### Bug Fixes

* **keychain:** eliminate destructive host keychain purging and isolate test runs ([#54](https://github.com/adrijshikhar/aim/issues/54)) ([c95ac8f](https://github.com/adrijshikhar/aim/commit/c95ac8f539db7a2f6bbb4e2d54bb22c1726e5bb3))

## [0.9.0](https://github.com/adrijshikhar/aim/compare/v0.8.1...v0.9.0) (2026-09-25)


### Features

* add first-class Claude Code adapter and session provider ([#49](https://github.com/adrijshikhar/aim/issues/49)) ([df75233](https://github.com/adrijshikhar/aim/commit/df75233f9413ec54e1df265afd600df01bcfb193))
* implicit cross-profile resume sync, in-line session picker, dynamic autocompletion, terminal tab title, and decoupled homebrew release ([#47](https://github.com/adrijshikhar/aim/issues/47)) ([e51b825](https://github.com/adrijshikhar/aim/commit/e51b8255efadbc5cff1cee8417bc54243fba9215))


### Bug Fixes

* **claude:** persist profile-scoped authentication and prevent host token leakage ([#51](https://github.com/adrijshikhar/aim/issues/51)) ([5c1421c](https://github.com/adrijshikhar/aim/commit/5c1421c88de885fcb39de6cdddd09c929c7b2a50))
* **codex:** session resume sync, SQLite projection deduplication, and ancestor lineage hydration ([#53](https://github.com/adrijshikhar/aim/issues/53)) ([bc2c395](https://github.com/adrijshikhar/aim/commit/bc2c395b22756fab9392fd3565c5684633859a0f))

## [0.8.1](https://github.com/adrijshikhar/aim/compare/v0.8.0...v0.8.1) (2026-09-22)


### Bug Fixes

* **session/codex:** refresh stale ancestor rollout files during hydration ([#45](https://github.com/adrijshikhar/aim/issues/45)) ([fe45409](https://github.com/adrijshikhar/aim/commit/fe45409d1778027215dde3df1417243a3ec139a5))

## [0.8.0](https://github.com/adrijshikhar/aim/compare/v0.7.0...v0.8.0) (2026-09-19)


### Features

* **cli:** prompt before creating missing profiles and warn on dash typos ([#44](https://github.com/adrijshikhar/aim/issues/44)) ([fe80616](https://github.com/adrijshikhar/aim/commit/fe806162d4e49b47a4d5f4a7e937772b4a413a54))
* **codex:** add sidecar daemon auto-start, hook trust rewriting, and agent onboarding skill ([#38](https://github.com/adrijshikhar/aim/issues/38)) ([1a82274](https://github.com/adrijshikhar/aim/commit/1a822741fd98a4716256bdde12e505d8138cc410))


### Bug Fixes

* **codex:** deduplicate TOML tables and auto-repair config in doctor ([#40](https://github.com/adrijshikhar/aim/issues/40)) ([5893cb7](https://github.com/adrijshikhar/aim/commit/5893cb710a5487264f9df7a8e3555e2125d39cec))
* **session/codex:** full-fidelity session hydration and ancestor tree replication ([#42](https://github.com/adrijshikhar/aim/issues/42)) ([fd11d24](https://github.com/adrijshikhar/aim/commit/fd11d24873d6ebc72861d1b0dca729a8d34e7af8))

## [0.7.0](https://github.com/adrijshikhar/aim/compare/v0.6.0...v0.7.0) (2026-09-17)


### Features

* **codex:** bridge cxstatusline configuration into profiles ([#36](https://github.com/adrijshikhar/aim/issues/36)) ([243a0b7](https://github.com/adrijshikhar/aim/commit/243a0b747e1dc2fbeff611923d060ad5ec34a4fa))

## [0.6.0](https://github.com/adrijshikhar/aim/compare/v0.5.0...v0.6.0) (2026-09-16)


### Features

* **cli:** unify lipgloss styling and table formatting across all commands ([#30](https://github.com/adrijshikhar/aim/issues/30)) ([9b90b36](https://github.com/adrijshikhar/aim/commit/9b90b36ad225c46d44f99fd29371e40200799682))
* **profile:** add aim mv command and TUI 'v' key to move agent accounts ([#32](https://github.com/adrijshikhar/aim/issues/32)) ([89d975e](https://github.com/adrijshikhar/aim/commit/89d975ef1c5cce7a7556cd8c2e7086efee5bdb3b))


### Bug Fixes

* **auth:** enable browser auto-open when re-authenticating expired profiles ([#34](https://github.com/adrijshikhar/aim/issues/34)) ([5ea7c12](https://github.com/adrijshikhar/aim/commit/5ea7c1288920ef403d780df9a9450819a2154b90))

## [0.5.0](https://github.com/adrijshikhar/aim/compare/v0.4.0...v0.5.0) (2026-09-14)


### Features

* **session:** cross-profile conversation resumption & catalyst handoff ([#27](https://github.com/adrijshikhar/aim/issues/27)) ([2566fc7](https://github.com/adrijshikhar/aim/commit/2566fc7845c0ea142fa1f85628e24c5aef8365e1))

## [0.4.0](https://github.com/adrijshikhar/aim/compare/v0.3.3...v0.4.0) (2026-09-14)


### Features

* Codex adapter, K9s-inspired TUI refactor, XDG paths, help & filter ([#23](https://github.com/adrijshikhar/aim/issues/23)) ([d3503ef](https://github.com/adrijshikhar/aim/commit/d3503ef9e8367d19eacb37d9ca17f6a9be00190a))

## [0.3.3](https://github.com/adrijshikhar/aim/compare/v0.3.2...v0.3.3) (2026-09-13)


### Bug Fixes

* **agy:** bridge plugins, skills, and config across profiles ([#20](https://github.com/adrijshikhar/aim/issues/20)) ([9ab599d](https://github.com/adrijshikhar/aim/commit/9ab599d0d4e227b83b76bf5cd56899088cf79d7e))

## [0.3.2](https://github.com/adrijshikhar/aim/compare/v0.3.1...v0.3.2) (2026-09-13)


### Bug Fixes

* **tui:** show application version in header tagline ([#18](https://github.com/adrijshikhar/aim/issues/18)) ([c7cbc5c](https://github.com/adrijshikhar/aim/commit/c7cbc5cd8a55968aac58da422b9618eada749618))

## [0.3.1](https://github.com/adrijshikhar/aim/compare/v0.3.0...v0.3.1) (2026-09-12)


### Bug Fixes

* **tui:** rename bottom inspector header to profile details ([#15](https://github.com/adrijshikhar/aim/issues/15)) ([7635a24](https://github.com/adrijshikhar/aim/commit/7635a24f8832e0ddfb2f290dd12d41e8c354d09a))


### Performance Improvements

* optimize profile cloning, symlinking, cache serialization, and CLI latency ([#13](https://github.com/adrijshikhar/aim/issues/13)) ([01f931a](https://github.com/adrijshikhar/aim/commit/01f931a386ed2381e03dad990cbc0a6704e31e66))

## [0.3.0](https://github.com/adrijshikhar/aim/compare/v0.2.1...v0.3.0) (2026-09-12)


### Features

* **auth & tui:** browser auto-open, keychain harvest, account display & ghost profile prevention ([#11](https://github.com/adrijshikhar/aim/issues/11)) ([a5815c5](https://github.com/adrijshikhar/aim/commit/a5815c5644fbb36f56010005b8af6c9c282a0c7e))
* **usage:** add account column to usage table and improve multi-category quota accuracy ([#14](https://github.com/adrijshikhar/aim/issues/14)) ([da035b2](https://github.com/adrijshikhar/aim/commit/da035b21e77c88c24f04a47540c42638b71d79f8))

## [0.2.1](https://github.com/adrijshikhar/aim/compare/v0.2.0...v0.2.1) (2026-09-12)


### Bug Fixes

* **auth:** auto-seed host keychain credentials and harvest tokens on login ([#9](https://github.com/adrijshikhar/aim/issues/9)) ([1c57af6](https://github.com/adrijshikhar/aim/commit/1c57af6b31d1658d60af096790c9c6ea298d869b))

## [0.2.0](https://github.com/adrijshikhar/aim/compare/v0.1.2...v0.2.0) (2026-09-12)


### Features

* add interactive profile rename modal in TUI ([#6](https://github.com/adrijshikhar/aim/issues/6)) ([d26f987](https://github.com/adrijshikhar/aim/commit/d26f987ebcfab1acc3b1986948751d331e92f33b))
* add opt-in debug logging and macOS keychain isolation ignore list ([#8](https://github.com/adrijshikhar/aim/issues/8)) ([e91e0f0](https://github.com/adrijshikhar/aim/commit/e91e0f04cfc923c09118b76bb8a22791d16beee6))

## [0.1.2](https://github.com/adrijshikhar/aim/compare/v0.1.1...v0.1.2) (2026-09-11)


### Bug Fixes

* **agy:** auto-seed host credentials for personal profile and update tagline to AI Multiplexer ([e9fb0d0](https://github.com/adrijshikhar/aim/commit/e9fb0d0b29a281b6ae65f0148af75eb0ad3f6f10))
* **ci:** fix relative path in tap git diff check ([82aaedb](https://github.com/adrijshikhar/aim/commit/82aaedbe2bb69b8f949b6b41b2d25da8cf9e99fd))
* **ci:** provide HOMEBREW_GITHUB_API_TOKEN to prevent audit rate-limiting ([6c52dcc](https://github.com/adrijshikhar/aim/commit/6c52dcc72347ea8966466b7e8470059c9a8ce9b8))

## [0.1.1](https://github.com/adrijshikhar/aim/compare/v0.1.0...v0.1.1) (2026-09-11)


### Features

* **goreleaser:** configure homebrew_casks for automated tap releases ([4cec193](https://github.com/adrijshikhar/aim/commit/4cec193670324285369c26429763df66ffb3d25c))
* **ui:** add Mole-inspired ASCII art logo and header to TUI and README ([b231151](https://github.com/adrijshikhar/aim/commit/b231151a9694048d2a798a5cceb078bcdb52cc99))


### Bug Fixes

* **installer:** add authenticated and gh fallback for private repo downloads ([acfb1d6](https://github.com/adrijshikhar/aim/commit/acfb1d66ed1ca343857aedd4bd754c18ea4c4974))


### Miscellaneous Chores

* target release version 0.1.1 ([6d00efc](https://github.com/adrijshikhar/aim/commit/6d00efc18352f13c0c79ba48fc4c1fd414e070a6))
