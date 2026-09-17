# Changelog

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
