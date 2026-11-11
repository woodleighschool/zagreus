# Changelog

## [0.1.9](https://github.com/woodleighschool/zagreus/compare/v0.1.8...v0.1.9) (2026-11-11)


### Bug Fixes

* **go:** update module github.com/go-playground/validator/v10 (v10.30.5 → v10.30.6) ([#56](https://github.com/woodleighschool/zagreus/issues/56)) ([540fd5d](https://github.com/woodleighschool/zagreus/commit/540fd5d5e6061ea5ce8f199b6da692ee4e22c1a0))
* **go:** update module go.yaml.in/yaml/v4 (v4.0.0-rc.7 → v4.0.0-rc.8) ([#53](https://github.com/woodleighschool/zagreus/issues/53)) ([265d481](https://github.com/woodleighschool/zagreus/commit/265d481558379d2cdbc8e2368b292f693dc99fd6))
* **go:** update module go.yaml.in/yaml/v4 (v4.0.0-rc.8 → v4.0.0-rc.9) ([#57](https://github.com/woodleighschool/zagreus/issues/57)) ([a7dae72](https://github.com/woodleighschool/zagreus/commit/a7dae72b62c87a9dc5ff1746681742ba6053bc4b))
* **go:** update module go.yaml.in/yaml/v4 (v4.0.0-rc.9 → v4.0.0) ([#58](https://github.com/woodleighschool/zagreus/issues/58)) ([2e6c07f](https://github.com/woodleighschool/zagreus/commit/2e6c07fb5b37e54c586409abe694595e7f0364f9))
* **go:** update module golang.org/x/term (v0.47.0 → v0.48.0) ([#62](https://github.com/woodleighschool/zagreus/issues/62)) ([d91b455](https://github.com/woodleighschool/zagreus/commit/d91b455cbe0244d905618a4cb087603ecbae39cd))

## [0.1.8](https://github.com/woodleighschool/zagreus/compare/v0.1.7...v0.1.8) (2026-11-04)


### Bug Fixes

* fixed ambiguous yaml module version ([d468c49](https://github.com/woodleighschool/zagreus/commit/d468c493460226537d5ccc26b5ef439f4eed0568))

## [0.1.7](https://github.com/woodleighschool/zagreus/compare/v0.1.6...v0.1.7) (2026-11-04)


### Bug Fixes

* **go:** update module go.yaml.in/yaml/v3 (v3.0.5 → v3.0.6) ([#49](https://github.com/woodleighschool/zagreus/issues/49)) ([18bada9](https://github.com/woodleighschool/zagreus/commit/18bada9578f2574cdcc89cea5a5ac278a5ad0596))
* **go:** update module go.yaml.in/yaml/v4 (v4.0.0-rc.6 → v4.0.0-rc.7) ([#50](https://github.com/woodleighschool/zagreus/issues/50)) ([eb8c75b](https://github.com/woodleighschool/zagreus/commit/eb8c75b4b22ebbe43e198971a9ccac1d14d0517e))

## [0.1.6](https://github.com/woodleighschool/zagreus/compare/v0.1.5...v0.1.6) (2026-11-03)


### Bug Fixes

* fixed nessus API model and changed label reconciliation logic ([70a3d05](https://github.com/woodleighschool/zagreus/commit/70a3d0516ef1484538be326ef9af714c816838d1))

## [0.1.5](https://github.com/woodleighschool/zagreus/compare/v0.1.4...v0.1.5) (2026-10-28)


### Bug Fixes

* **reopen:** reopening cards now works properly ([1fa23d2](https://github.com/woodleighschool/zagreus/commit/1fa23d220e3ef9cd736917b5e39a410b85c2f5eb))

## [0.1.4](https://github.com/woodleighschool/zagreus/compare/v0.1.3...v0.1.4) (2026-10-19)


### Bug Fixes

* **logging:** tidying up logging to be more clear by default; added verbose ([bc6f02f](https://github.com/woodleighschool/zagreus/commit/bc6f02f28e90df7dacd7b717b6b62e61d70a4e1f))

## [0.1.3](https://github.com/woodleighschool/zagreus/compare/v0.1.2...v0.1.3) (2026-10-19)


### Bug Fixes

* **report:** vulnerability is null if skipping a card, adding safety check to prevent panic ([6b6c83a](https://github.com/woodleighschool/zagreus/commit/6b6c83a957105f4573671bfc33a09f6093812fd7))

## [0.1.2](https://github.com/woodleighschool/zagreus/compare/v0.1.1...v0.1.2) (2026-10-18)


### Bug Fixes

* **sync:** wasn't returning from function if err != nil ([6a54204](https://github.com/woodleighschool/zagreus/commit/6a542047194a1975d34490340294aac4b54c5d57))

## [0.1.1](https://github.com/woodleighschool/zagreus/compare/v0.1.0...v0.1.1) (2026-10-16)


### Bug Fixes

* **command:** misplaced a definition for log-level when rewriting ([141bef4](https://github.com/woodleighschool/zagreus/commit/141bef44334341c13d54dd6d0ca5b3d7798dbefa))

## 0.1.0 (2026-10-16)


### Features

* overhauled cli interface and other small tidy fixes ([63a3cf3](https://github.com/woodleighschool/zagreus/commit/63a3cf3f95aff52a5260a85fc4fc6d0bff02f638))


### Bug Fixes

* **build:** align Go toolchain and container versions with Mise ([d4b6b7b](https://github.com/woodleighschool/zagreus/commit/d4b6b7b8b2ada6345444a3d2462d56e443956537))
* **go:** update module github.com/mattn/go-runewidth (v0.0.13 → v0.0.31) ([#19](https://github.com/woodleighschool/zagreus/issues/19)) ([01f4456](https://github.com/woodleighschool/zagreus/commit/01f4456f33d7818e9d8647e311f03459daff04a3))
* **go:** update module golang.org/x/term (v0.46.0 → v0.47.0) ([#20](https://github.com/woodleighschool/zagreus/issues/20)) ([943a7e3](https://github.com/woodleighschool/zagreus/commit/943a7e36531495cbf3bf1f4f8cedae4a59b8e660))
* lock mise tools for CI installs ([aff3348](https://github.com/woodleighschool/zagreus/commit/aff3348ff1220b1bd25d801e578031ff2c7a4bdd))
* remediated golangci-lint issues, and finished most other features ([2c02285](https://github.com/woodleighschool/zagreus/commit/2c02285a51c2a5777c8f4d31242af445848956ae))
