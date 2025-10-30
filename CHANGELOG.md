# Changelog

## [2.2.1](https://github.com/rvolykh/vui/compare/v2.2.0...v2.2.1) (2025-10-30)


### Bug Fixes

* Allow to save secrets without key ([b8a473d](https://github.com/rvolykh/vui/commit/b8a473d99c4559c3aa4870be2e67ab90e6946912))

## [2.2.0](https://github.com/rvolykh/vui/compare/v2.1.0...v2.2.0) (2025-10-30)


### Features

* Add AWS SSM Parameters profile support ([#15](https://github.com/rvolykh/vui/issues/15)) ([d88c6b3](https://github.com/rvolykh/vui/commit/d88c6b322c0b161ead6dc12aa5cff83b4e35b2ec))

## [2.1.0](https://github.com/rvolykh/vui/compare/v2.0.0...v2.1.0) (2025-10-30)


### Features

* Add AWS SecretsManager profile support ([#13](https://github.com/rvolykh/vui/issues/13)) ([fab712b](https://github.com/rvolykh/vui/commit/fab712b3c2de72672a89296e8bcea10af30eb192))

## [2.0.0](https://github.com/rvolykh/vui/compare/v1.0.0...v2.0.0) (2025-10-30)


### ⚠ BREAKING CHANGES

* Replace Vaults with Profiles in config ([#11](https://github.com/rvolykh/vui/issues/11))

### Features

* Replace Vaults with Profiles in config ([#11](https://github.com/rvolykh/vui/issues/11)) ([0f22ee8](https://github.com/rvolykh/vui/commit/0f22ee89a12a27964067466c054813968f89f154))

## 1.0.0 (2025-10-23)


### Features

* Add Cert auth ([bc77379](https://github.com/rvolykh/vui/commit/bc7737942dabd77bd4ad5e37c3e7fab8b5dc21b7))
* Add clipboard dialing when value is copied ([97e47b0](https://github.com/rvolykh/vui/commit/97e47b08514ad4bcc4b77fe79f20546a8664ac86))
* Add JWT auth ([dd4572b](https://github.com/rvolykh/vui/commit/dd4572b5f94ac7f10d15e8c4d673e5d976b74c08))
* Add kubernetes auth along with Sandbox k8s profile ([1c9c220](https://github.com/rvolykh/vui/commit/1c9c220fa90410ad2a225713bf71cc9c1ec73f05))
* Add OIDC auth ([25fbb29](https://github.com/rvolykh/vui/commit/25fbb2941519409a046449dc14d73415b6a54886))
* Add UserPass auth ([8d17c6a](https://github.com/rvolykh/vui/commit/8d17c6a66b0663c3b174b7033c0ef61f03a8c9d4))
* Add version flag ([b03da87](https://github.com/rvolykh/vui/commit/b03da870ed98e98c397c05d5932f85fd4c81773b))
* Allow to navigate back to profiles page ([852b102](https://github.com/rvolykh/vui/commit/852b102a6b53c365c0a9114c647d94366a786b0c))
* Config add logging configuration ([1599a84](https://github.com/rvolykh/vui/commit/1599a84f67ecf4cede772af22e427629d5549b7a))
* Implement refresh on Vault Profile screen ([0215042](https://github.com/rvolykh/vui/commit/02150423f01825fb65e16268c6f7bcb75782bd84))
* Initial implementation ([fe73d2e](https://github.com/rvolykh/vui/commit/fe73d2ec95e217510a4c29f8d01f6ebfec4daebb))


### Bug Fixes

* Cleanup config ([fac47f2](https://github.com/rvolykh/vui/commit/fac47f27260d47d4211149cfd8388b29700433c1))
* Create a new secret ([960875e](https://github.com/rvolykh/vui/commit/960875ea57bbc2bb364e332825632718eb129c78))
* Delete secret action ([035d25a](https://github.com/rvolykh/vui/commit/035d25a360ce72260bdb8725f3e50f74b8867d68))
* Edit secret form ([63ec686](https://github.com/rvolykh/vui/commit/63ec686be1d452c316209a93907eca5091af612e))
* Fix rendering issue ([25d89e9](https://github.com/rvolykh/vui/commit/25d89e9d2e2a3bc4cbff12dd0a8960960da0ba4c))
* Improve delete secret confirmation dialog ([cc7d3d6](https://github.com/rvolykh/vui/commit/cc7d3d69c37753f77bea07425b72e4131cc8b3fa))
* Makefile - clean target ([f40c100](https://github.com/rvolykh/vui/commit/f40c100122e07ebe9de5636aa527f97cb6231daa))
* Preserve keys order on Value view ([c630fed](https://github.com/rvolykh/vui/commit/c630fed8c317bd20e55afc48241a83485e6b9773))
* Remove redundant code ([2f15565](https://github.com/rvolykh/vui/commit/2f15565716d4ed8eba03bab6695c9018983a6258))
* Replace value InputField with TextArea ([b30cd4c](https://github.com/rvolykh/vui/commit/b30cd4c13b9eaa191118db3c0e6b0f50e12d00fa))
* Save action on Edit Secret form ([5a8c9d0](https://github.com/rvolykh/vui/commit/5a8c9d08a757d56814ea9bb0a4065bc1696ee281))
* Verify Vault auth before rendering next page ([9585c98](https://github.com/rvolykh/vui/commit/9585c98efa21e2ebec376a7032773fe0a626ae52))


### Miscellaneous Chores

* first release ([fb2d647](https://github.com/rvolykh/vui/commit/fb2d64725018371d96aae0ae080d9b19202b0bc7))
