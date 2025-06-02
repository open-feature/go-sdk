# Changelog

## [2.0.0](https://github.com/open-feature/go-sdk/compare/v1.14.1...v2.0.0) (2025-06-02)


### ⚠ BREAKING CHANGES

* simplify and improve logging hook ([#372](https://github.com/open-feature/go-sdk/issues/372))

### 🐛 Bug Fixes

* **deps:** update module github.com/go-logr/logr to v1.4.3 ([#375](https://github.com/open-feature/go-sdk/issues/375)) ([f1b6836](https://github.com/open-feature/go-sdk/commit/f1b683682405885ee08ffda4269449666b7d7c1f))
* **deps:** update module go.uber.org/mock to v0.5.2 ([#355](https://github.com/open-feature/go-sdk/issues/355)) ([1d6069a](https://github.com/open-feature/go-sdk/commit/1d6069a9c56ca0d89a093f7f74168103295ecee8))
* **deps:** update module golang.org/x/text to v0.22.0 ([#315](https://github.com/open-feature/go-sdk/issues/315)) ([06e9304](https://github.com/open-feature/go-sdk/commit/06e93041fc916d607b42f295498915a6ce5993e1))
* **deps:** update module golang.org/x/text to v0.25.0 ([#330](https://github.com/open-feature/go-sdk/issues/330)) ([42e33ba](https://github.com/open-feature/go-sdk/commit/42e33ba0542a59677389d5542b9ee3cc7f8fdb3d))
* remove `ITracking` interface in favour of `Tracker` ([#365](https://github.com/open-feature/go-sdk/issues/365)) ([3b4c49b](https://github.com/open-feature/go-sdk/commit/3b4c49b111b515a5fd84f04a6de68b1e64168092))


### ✨ New Features

* add evaluation details to finally hook ([#328](https://github.com/open-feature/go-sdk/issues/328)) ([63de64a](https://github.com/open-feature/go-sdk/commit/63de64af7fb1b82aa224e5991f4a17f111df83c3))
* add OTel event creation util func ([#325](https://github.com/open-feature/go-sdk/issues/325)) ([3c70dc2](https://github.com/open-feature/go-sdk/commit/3c70dc2b50d5f088a92926ee09bceaebfc8595a1))
* loosen test instance interface to support various test frameworks ([#317](https://github.com/open-feature/go-sdk/issues/317)) ([ce4920a](https://github.com/open-feature/go-sdk/commit/ce4920ad0dde1023da06082d81cb3b682b133c7c))
* loosen test instance interface to support various test frameworks [#316](https://github.com/open-feature/go-sdk/issues/316) ([#319](https://github.com/open-feature/go-sdk/issues/319)) ([344cd39](https://github.com/open-feature/go-sdk/commit/344cd39e3e834423846da4bf2e0e92273e27a9a4))
* Make Mocks Exportable ([#341](https://github.com/open-feature/go-sdk/issues/341)) ([f675d01](https://github.com/open-feature/go-sdk/commit/f675d018df7196043bc075a9992d7445bd9359c2))
* simplify and improve logging hook ([#372](https://github.com/open-feature/go-sdk/issues/372)) ([5043b24](https://github.com/open-feature/go-sdk/commit/5043b248933c8556282f56ab323f920412448dd4))
* update telemetry package to conform to latest spec ([#371](https://github.com/open-feature/go-sdk/issues/371)) ([5c69ff1](https://github.com/open-feature/go-sdk/commit/5c69ff122e650357701db1870eb501f4e87808df))


### 🧹 Chore

* add `misspell` and `dupword` linters ([#374](https://github.com/open-feature/go-sdk/issues/374)) ([1b1af45](https://github.com/open-feature/go-sdk/commit/1b1af4519c590426e3b975b2b3066d6f5b68daef))
* add global maintainers to codeowners ([#338](https://github.com/open-feature/go-sdk/issues/338)) ([6d1004c](https://github.com/open-feature/go-sdk/commit/6d1004c20ac94d994c67287f01049e401ca5ba68))
* add golangci-lint config file ([#366](https://github.com/open-feature/go-sdk/issues/366)) ([8844be4](https://github.com/open-feature/go-sdk/commit/8844be4153cffe4b9e730712748d1d90759425d1))
* add Makefile `docs` target ([#352](https://github.com/open-feature/go-sdk/issues/352)) ([0bdbe12](https://github.com/open-feature/go-sdk/commit/0bdbe1266cf9e77c9f9bcfd98edb55ede3cf90c1))
* **deps:** pin dependencies ([#335](https://github.com/open-feature/go-sdk/issues/335)) ([68d9cf3](https://github.com/open-feature/go-sdk/commit/68d9cf381ee2e3ce5be37d6a5d623018b633b413))
* **deps:** update codecov/codecov-action action to v5 ([#303](https://github.com/open-feature/go-sdk/issues/303)) ([44a74c3](https://github.com/open-feature/go-sdk/commit/44a74c3d603aa51d2cc36adad091c056bd2225d5))
* **deps:** update codecov/codecov-action action to v5.4.0 ([#329](https://github.com/open-feature/go-sdk/issues/329)) ([a55418d](https://github.com/open-feature/go-sdk/commit/a55418d9ab2bae4c4d8a922c6928fb441e79f5fa))
* **deps:** update codecov/codecov-action action to v5.4.2 ([#339](https://github.com/open-feature/go-sdk/issues/339)) ([05a6d97](https://github.com/open-feature/go-sdk/commit/05a6d97ad4f7c4d2391a239621e538cbbd81cb04))
* **deps:** update codecov/codecov-action action to v5.4.3 ([#356](https://github.com/open-feature/go-sdk/issues/356)) ([9ff5e20](https://github.com/open-feature/go-sdk/commit/9ff5e204c726f3a60c8c04c89f8841dc16a6ac53))
* require go 1.23.0 ([#351](https://github.com/open-feature/go-sdk/issues/351)) ([3c52fa7](https://github.com/open-feature/go-sdk/commit/3c52fa7fe1dc3f256954a4035c0d317bc076d5e7))
* update renovate config to utilize general openfeature one ([#334](https://github.com/open-feature/go-sdk/issues/334)) ([58a2a17](https://github.com/open-feature/go-sdk/commit/58a2a1743c8a461d3e073e1f8f6aa8c59648c1a9))
* update test-harness ([#376](https://github.com/open-feature/go-sdk/issues/376)) ([36241dc](https://github.com/open-feature/go-sdk/commit/36241dc6b60f165ad37c60654236998d50aa990d))


### 📚 Documentation

* add PR title and DCO rules to CONTRIBUTING.md ([#321](https://github.com/open-feature/go-sdk/issues/321)) ([c2eb5af](https://github.com/open-feature/go-sdk/commit/c2eb5af651dfec4dc099892bc53a5d8a84b895e1))
* deprecate identifiers correctly ([#357](https://github.com/open-feature/go-sdk/issues/357)) ([965e766](https://github.com/open-feature/go-sdk/commit/965e7665c7d53e2cf46c97f4de779e5002c7e3e8))
* format godoc lists correctly ([#353](https://github.com/open-feature/go-sdk/issues/353)) ([a905ce3](https://github.com/open-feature/go-sdk/commit/a905ce3adcf47ec5fd66fe870d0c0661929340f3))
* Use `SetProviderAndWait` in README ([#332](https://github.com/open-feature/go-sdk/issues/332)) ([a7dbd93](https://github.com/open-feature/go-sdk/commit/a7dbd931903eb9734ac1119fdd1226f233e3944d))
* Use `SetProviderAndWait` in README example ([#333](https://github.com/open-feature/go-sdk/issues/333)) ([74eaac3](https://github.com/open-feature/go-sdk/commit/74eaac3804c3041a0ae798bff14a30baa099608c))
* use shorthand eval syntax ([#331](https://github.com/open-feature/go-sdk/issues/331)) ([69d31fe](https://github.com/open-feature/go-sdk/commit/69d31fe089e96693375075a6ae600f3b88ae6422))


### 🔄 Refactoring

* remove dependency on golang.org/x/exp ([#320](https://github.com/open-feature/go-sdk/issues/320)) ([38e2300](https://github.com/open-feature/go-sdk/commit/38e2300c67a8316bfb0e319dddfaa95236bc9edc))
* simplify and improve logging hook ([#361](https://github.com/open-feature/go-sdk/issues/361)) ([5d41ca0](https://github.com/open-feature/go-sdk/commit/5d41ca06cdc9c8b66394ccd9ee19d9dbd4a5ad4d))

## [1.14.1](https://github.com/open-feature/go-sdk/compare/v1.14.0...v1.14.1) (2025-01-21)


### 🐛 Bug Fixes

* internal provider comparison causing race conditions in tests ([440072f](https://github.com/open-feature/go-sdk/commit/440072f917b4d767e2d4b7ad0a9c73b358d2818b))
* internal provider comparison causing race conditions in tests ([#312](https://github.com/open-feature/go-sdk/issues/312)) ([440072f](https://github.com/open-feature/go-sdk/commit/440072f917b4d767e2d4b7ad0a9c73b358d2818b))


### 📚 Documentation

* add missing entry for "tracking" in feature table ([#313](https://github.com/open-feature/go-sdk/issues/313)) ([890bfd0](https://github.com/open-feature/go-sdk/commit/890bfd0d00ee267890c67f7e3b1c88abd24ebb87))

## [1.14.0](https://github.com/open-feature/go-sdk/compare/v1.13.1...v1.14.0) (2024-12-04)


### 🐛 Bug Fixes

* consider default client when loading states ([d906d6f](https://github.com/open-feature/go-sdk/commit/d906d6fe3e228d18e93c2d07be53ef07fb787083))
* consider default client when loading states ([#309](https://github.com/open-feature/go-sdk/issues/309)) ([d906d6f](https://github.com/open-feature/go-sdk/commit/d906d6fe3e228d18e93c2d07be53ef07fb787083))
* **deps:** update module github.com/cucumber/godog to v0.15.0 ([#300](https://github.com/open-feature/go-sdk/issues/300)) ([dc66ed3](https://github.com/open-feature/go-sdk/commit/dc66ed366b781e09a56ddcfee7eec270ce13868c))
* **deps:** update module golang.org/x/text to v0.20.0 ([#301](https://github.com/open-feature/go-sdk/issues/301)) ([efab565](https://github.com/open-feature/go-sdk/commit/efab565e5679717bcb35cec2b9e5b106b7baed77))
* **deps:** update module golang.org/x/text to v0.21.0 ([#311](https://github.com/open-feature/go-sdk/issues/311)) ([3d1615a](https://github.com/open-feature/go-sdk/commit/3d1615a8f4e873a128f4060467a970b15bf1c15e))


### ✨ New Features

* add logging hook, rm logging from evaluation ([#289](https://github.com/open-feature/go-sdk/issues/289)) ([7850eec](https://github.com/open-feature/go-sdk/commit/7850eec993f7570ab2603c52edfccea7a36f8a09))
* Implement Tracking in Go ([#297](https://github.com/open-feature/go-sdk/issues/297)) ([dee5ec7](https://github.com/open-feature/go-sdk/commit/dee5ec72a9ec3da272804032a8fc8f469f36345c))
* make provider interface "stateless" as per spec 0.8.0 ([#299](https://github.com/open-feature/go-sdk/issues/299)) ([510b2a6](https://github.com/open-feature/go-sdk/commit/510b2a6fdb54567e4bce93a2cd3eeeab5ec0a168))
* TestProvider for easy, parallel-safe testing ([#295](https://github.com/open-feature/go-sdk/issues/295)) ([3e3d0b1](https://github.com/open-feature/go-sdk/commit/3e3d0b17b56938d09514ca30059403ab916d2291))


### 🧹 Chore

* fix coverage uploads to Codecov ([#306](https://github.com/open-feature/go-sdk/issues/306)) ([96d86ba](https://github.com/open-feature/go-sdk/commit/96d86bac249d8eb26cfbe445bd2739d12a5e5686))
* require go 1.21 ([#294](https://github.com/open-feature/go-sdk/issues/294)) ([ddfffdd](https://github.com/open-feature/go-sdk/commit/ddfffddf52c1cbd48e0c40a828f33d3342824706))
* run e2e tests in ci, add sm pull to makefile ([#298](https://github.com/open-feature/go-sdk/issues/298)) ([8675832](https://github.com/open-feature/go-sdk/commit/8675832174d876eccd73378ff86eb21ded2016b8))
* use correct term "domain" instead of name/clientName in code ([#305](https://github.com/open-feature/go-sdk/issues/305)) ([73c0e85](https://github.com/open-feature/go-sdk/commit/73c0e85795c8c0a0ccb70c7d3babee0f3c8e2fd5))


### 📚 Documentation

* update readme to be included in the docs ([13444e7](https://github.com/open-feature/go-sdk/commit/13444e73dec4c5a14136e6b88d776c68527d7963))

## [1.13.1](https://github.com/open-feature/go-sdk/compare/v1.13.0...v1.13.1) (2024-10-18)


### 🐛 Bug Fixes

* **deps:** update module golang.org/x/text to v0.19.0 ([#290](https://github.com/open-feature/go-sdk/issues/290)) ([cd5b89a](https://github.com/open-feature/go-sdk/commit/cd5b89a5c404b374612295c22f8ec31a409bc033))
* Prevent panic when setting non-comparable named providers ([#286](https://github.com/open-feature/go-sdk/issues/286)) ([01953bb](https://github.com/open-feature/go-sdk/commit/01953bb18a1f80e9ce3d90f2ed146332cf611d64))


### 🧹 Chore

* **deps:** update codecov/codecov-action action to v4.6.0 ([#287](https://github.com/open-feature/go-sdk/issues/287)) ([368336f](https://github.com/open-feature/go-sdk/commit/368336f303b80be2298441fcb7440c29205c9b0e))


### 📚 Documentation

* update Go package manager link ([#292](https://github.com/open-feature/go-sdk/issues/292)) ([3551f3c](https://github.com/open-feature/go-sdk/commit/3551f3c08610b5acd1a40fb4c30d88100bfbd258))

## [1.13.0](https://github.com/open-feature/go-sdk/compare/v1.12.0...v1.13.0) (2024-09-09)


### 🐛 Bug Fixes

* **deps:** update module golang.org/x/text to v0.16.0 ([#277](https://github.com/open-feature/go-sdk/issues/277)) ([80c0235](https://github.com/open-feature/go-sdk/commit/80c0235d26250a739781729855cb0264e19a4409))
* **deps:** update module golang.org/x/text to v0.18.0 ([#281](https://github.com/open-feature/go-sdk/issues/281)) ([ad9db29](https://github.com/open-feature/go-sdk/commit/ad9db29564e84b9c8c9bc51c9a3b0a8f93b8f2e6))


### ✨ New Features

* add TransactionContext ([#283](https://github.com/open-feature/go-sdk/issues/283)) ([788151d](https://github.com/open-feature/go-sdk/commit/788151d2fdc69429ef1841cba91ad5fc2103e0ca))


### 🧹 Chore

* **deps:** update codecov/codecov-action action to v4.5.0 ([#279](https://github.com/open-feature/go-sdk/issues/279)) ([881e019](https://github.com/open-feature/go-sdk/commit/881e019c67873f4b5477355ad90d26cd4530d017))

## [1.12.0](https://github.com/open-feature/go-sdk/compare/v1.11.0...v1.12.0) (2024-05-29)


### 🐛 Bug Fixes

* **deps:** update module github.com/cucumber/godog to v0.14.1 ([#267](https://github.com/open-feature/go-sdk/issues/267)) ([2cf5717](https://github.com/open-feature/go-sdk/commit/2cf5717eadbeff55923ae7a77428c03c3c9f6a70))
* **deps:** update module github.com/go-logr/logr to v1.4.2 ([#275](https://github.com/open-feature/go-sdk/issues/275)) ([aeb4f6c](https://github.com/open-feature/go-sdk/commit/aeb4f6c96535f379252afe797a2af5b0e88cec36))
* **deps:** update module golang.org/x/exp to v0.0.0-20240416160154-fe59bbe5cc7f ([#269](https://github.com/open-feature/go-sdk/issues/269)) ([45596a5](https://github.com/open-feature/go-sdk/commit/45596a5e063351e86ca178bac226d8950b51bbcd))
* **deps:** update module golang.org/x/exp to v0.0.0-20240506185415-9bf2ced13842 ([#272](https://github.com/open-feature/go-sdk/issues/272)) ([1c07c5b](https://github.com/open-feature/go-sdk/commit/1c07c5bdfce2d8fc99cf008b2e57b9262ffa4032))
* **deps:** update module golang.org/x/text to v0.15.0 ([#271](https://github.com/open-feature/go-sdk/issues/271)) ([dc28442](https://github.com/open-feature/go-sdk/commit/dc28442da78a8adadab7cad984549439ad864722))


### ✨ New Features

* Implement domain scoping ([#261](https://github.com/open-feature/go-sdk/issues/261)) ([a9e19dd](https://github.com/open-feature/go-sdk/commit/a9e19dd5ec52bf8d84a8b2e2f6d263d841059d61))
* isolate interfaces from SDK to improve testability  ([#268](https://github.com/open-feature/go-sdk/issues/268)) ([5e06c45](https://github.com/open-feature/go-sdk/commit/5e06c45fc9b2df1535a4a5d035b7b0c0412e6e6a))


### 🧹 Chore

* bump Go to version 1.20 ([#255](https://github.com/open-feature/go-sdk/issues/255)) ([fbec799](https://github.com/open-feature/go-sdk/commit/fbec7995ba57ba407cc43dcba0c4c558ddb0324f))
* **deps:** update codecov/codecov-action action to v4 ([#250](https://github.com/open-feature/go-sdk/issues/250)) ([a488697](https://github.com/open-feature/go-sdk/commit/a48869773e683bdc95dc61722419b95dba271b71))
* **deps:** update codecov/codecov-action action to v4.3.1 ([#270](https://github.com/open-feature/go-sdk/issues/270)) ([080a87b](https://github.com/open-feature/go-sdk/commit/080a87bfcac887c65b81fc5141d57000d407eec5))
* **deps:** update codecov/codecov-action action to v4.4.0 ([#273](https://github.com/open-feature/go-sdk/issues/273)) ([266cfc0](https://github.com/open-feature/go-sdk/commit/266cfc069ca674b3de0915c58106e772fd0a5867))
* **deps:** update codecov/codecov-action action to v4.4.1 ([#274](https://github.com/open-feature/go-sdk/issues/274)) ([c4ca1a8](https://github.com/open-feature/go-sdk/commit/c4ca1a8e6f8fb2f968e15cc961de2a6741a0a431))
* **deps:** update goreleaser/goreleaser-action action to v5 ([#219](https://github.com/open-feature/go-sdk/issues/219)) ([71854d4](https://github.com/open-feature/go-sdk/commit/71854d424c6c18ca89179ee0a802936935754a1b))

## [1.11.0](https://github.com/open-feature/go-sdk/compare/v1.10.0...v1.11.0) (2024-04-09)


### ✨ New Features

* Adding simplified evaluation methods ([#263](https://github.com/open-feature/go-sdk/issues/263)) ([7b610c7](https://github.com/open-feature/go-sdk/commit/7b610c7c5019535ffde98c0ad725a0e73e7703a1))


### 🧹 Chore

* **deps:** update actions/checkout action to v4 ([#212](https://github.com/open-feature/go-sdk/issues/212)) ([2944608](https://github.com/open-feature/go-sdk/commit/2944608562d04521fcf92000948a22012e630471))
* **deps:** update actions/setup-go action to v5 ([#237](https://github.com/open-feature/go-sdk/issues/237)) ([53d9e7e](https://github.com/open-feature/go-sdk/commit/53d9e7e4fd0127b4819de5d4d8c82735ee9d0c83))
* **deps:** update cyclonedx/gh-gomod-generate-sbom action to v2 ([#179](https://github.com/open-feature/go-sdk/issues/179)) ([b624a43](https://github.com/open-feature/go-sdk/commit/b624a43fc2fa132399543de79b9ecf5008312aaf))

## [1.10.0](https://github.com/open-feature/go-sdk/compare/v1.9.0...v1.10.0) (2024-02-07)


### 🐛 Bug Fixes

* **deps:** update module github.com/cucumber/godog to v0.14.0 ([#249](https://github.com/open-feature/go-sdk/issues/249)) ([bed4eaa](https://github.com/open-feature/go-sdk/commit/bed4eaa519cfea072041a130036f09b191190f09))
* **deps:** update module github.com/go-logr/logr to v1.4.0 ([#241](https://github.com/open-feature/go-sdk/issues/241)) ([72e4317](https://github.com/open-feature/go-sdk/commit/72e4317adc1987a932d0493463c673b09b0c4fe0))
* **deps:** update module github.com/go-logr/logr to v1.4.1 ([#243](https://github.com/open-feature/go-sdk/issues/243)) ([95f592a](https://github.com/open-feature/go-sdk/commit/95f592a50ff669b32a071625029b0566c750c534))


### ✨ New Features

* blocking provider mutator ([#251](https://github.com/open-feature/go-sdk/issues/251)) ([6f71fe4](https://github.com/open-feature/go-sdk/commit/6f71fe40505f07fc241953ee9da8ea4edc6e4d35))
* update to go 1.19 ([#252](https://github.com/open-feature/go-sdk/issues/252)) ([47f8a46](https://github.com/open-feature/go-sdk/commit/47f8a46b7992ce8943170dd7135eb73a9de6d9e2))


### 🧹 Chore

* **deps:** update actions/cache action to v4 ([#246](https://github.com/open-feature/go-sdk/issues/246)) ([eaefcc8](https://github.com/open-feature/go-sdk/commit/eaefcc8875693dd08ee634f47d7f053b7b7fc9cf))
* improve eventing ([#248](https://github.com/open-feature/go-sdk/issues/248)) ([d2c1636](https://github.com/open-feature/go-sdk/commit/d2c1636cd5ea46a4f15083e5d4f90ce54b5fb493))

## [1.9.0](https://github.com/open-feature/go-sdk/compare/v1.8.0...v1.9.0) (2023-11-21)


### 🐛 Bug Fixes

* change typo in readme ([#228](https://github.com/open-feature/go-sdk/issues/228)) ([6795fe1](https://github.com/open-feature/go-sdk/commit/6795fe16c24c695d58474b82284cf8b697a04a3a))
* **deps:** update module github.com/go-logr/logr to v1.3.0 ([#230](https://github.com/open-feature/go-sdk/issues/230)) ([6ab7984](https://github.com/open-feature/go-sdk/commit/6ab79842758518ed63dc712c11d824ba11110dc2))
* **deps:** update module golang.org/x/text to v0.14.0 ([#231](https://github.com/open-feature/go-sdk/issues/231)) ([34fb9d9](https://github.com/open-feature/go-sdk/commit/34fb9d968e6d3b34bd2adca5c1f9aaa833e9e437))


### ✨ New Features

* Repackage SDK from `pkg/openfeature` to `openfeature`. ([#232](https://github.com/open-feature/go-sdk/issues/232)) ([991726c](https://github.com/open-feature/go-sdk/commit/991726ced66913de916ec47ed5dd1837b6daf203))


### 🧹 Chore

* update package usage ([#235](https://github.com/open-feature/go-sdk/issues/235)) ([97204a4](https://github.com/open-feature/go-sdk/commit/97204a47766913e9156cae1976ed13ec291cdea0))
* update spec release link ([b8cb413](https://github.com/open-feature/go-sdk/commit/b8cb4132e5f03aacea2e9e519420e0848748be1c))

## [1.8.0](https://github.com/open-feature/go-sdk/compare/v1.7.0...v1.8.0) (2023-09-26)


### 🐛 Bug Fixes

* **deps:** update module github.com/cucumber/godog to v0.13.0 ([#210](https://github.com/open-feature/go-sdk/issues/210)) ([33c5f2f](https://github.com/open-feature/go-sdk/commit/33c5f2f5de478ee7123ccf0ffe594fcaf4d2555b))
* **deps:** update module golang.org/x/text to v0.13.0 ([#211](https://github.com/open-feature/go-sdk/issues/211)) ([d850ebc](https://github.com/open-feature/go-sdk/commit/d850ebc5c831ccee5edd490e3c2f019e2188b4ad))


### ✨ New Features

* run event handlers immediately, add STALE (0.7.0 compliance) ([#221](https://github.com/open-feature/go-sdk/issues/221)) ([9c0012f](https://github.com/open-feature/go-sdk/commit/9c0012f6762926489d2763b60e486170fdce9c09))


### 🧹 Chore

* bump spec badge in readme to v0.7.0 ([#223](https://github.com/open-feature/go-sdk/issues/223)) ([403275e](https://github.com/open-feature/go-sdk/commit/403275e925d5715e8f90d87296b7e3f626f8fb14))
* **deps:** update codecov/codecov-action action to v4 ([#222](https://github.com/open-feature/go-sdk/issues/222)) ([1ac250b](https://github.com/open-feature/go-sdk/commit/1ac250bc21996d247c38639eb96099e64b10541c))
* fix golangci-lint version ([#216](https://github.com/open-feature/go-sdk/issues/216)) ([e79382a](https://github.com/open-feature/go-sdk/commit/e79382a748fb914e0fa94a61d292e509094cdc46))
* fix logo rendering outside of github ([#226](https://github.com/open-feature/go-sdk/issues/226)) ([e2b3586](https://github.com/open-feature/go-sdk/commit/e2b35865b9ca48a5fe772307dd206b46553da51a))
* revert to CodeCov Action to v3 ([#225](https://github.com/open-feature/go-sdk/issues/225)) ([152416d](https://github.com/open-feature/go-sdk/commit/152416df30e01fe22112105358481eeb2d03a160))
* sort imports of go files ([#214](https://github.com/open-feature/go-sdk/issues/214)) ([a98950d](https://github.com/open-feature/go-sdk/commit/a98950d3f51a1e3ff4660a8ffeb5dcee50876304))
* update comments for named provider related function ([#213](https://github.com/open-feature/go-sdk/issues/213)) ([2e670b2](https://github.com/open-feature/go-sdk/commit/2e670b27391e6a1d5b77ad78a18db23e41c43a50))


### 📚 Documentation

* Update README.md ([#218](https://github.com/open-feature/go-sdk/issues/218)) ([a2ea804](https://github.com/open-feature/go-sdk/commit/a2ea804bcf53307ee1e3bf92b52de812d1358a4f))


### 🔄 Refactoring

* write [T]Value in terms of [T]ValueDetails ([#224](https://github.com/open-feature/go-sdk/issues/224)) ([f554876](https://github.com/open-feature/go-sdk/commit/f554876e5ed32cdb45aaf396ae2214bad28c3c26))

## [1.7.0](https://github.com/open-feature/go-sdk/compare/v1.6.0...v1.7.0) (2023-08-11)


### 🐛 Bug Fixes

* **deps:** update golang.org/x/exp digest to 89c5cff ([#195](https://github.com/open-feature/go-sdk/issues/195)) ([61680ed](https://github.com/open-feature/go-sdk/commit/61680ed2b4dfcff45758e714c348703da82beea3))
* **deps:** update module github.com/open-feature/go-sdk-contrib/providers/flagd to v0.1.15 ([#206](https://github.com/open-feature/go-sdk/issues/206)) ([6916ff9](https://github.com/open-feature/go-sdk/commit/6916ff9e33332b03d2aeca332c1c65ff032fe786))
* **deps:** update module github.com/open-feature/go-sdk-contrib/tests/flagd to v1.2.4 ([#201](https://github.com/open-feature/go-sdk/issues/201)) ([ddcc2d4](https://github.com/open-feature/go-sdk/commit/ddcc2d423478256ac62e6ba9ccc75bea86d4110c))
* **deps:** update module golang.org/x/text to v0.12.0 ([#207](https://github.com/open-feature/go-sdk/issues/207)) ([fc2bc30](https://github.com/open-feature/go-sdk/commit/fc2bc3088dd9dde099a74cdcc68ffb3a12bf5b4f))


### ✨ New Features

* in-memory provider for sdk testing ([#192](https://github.com/open-feature/go-sdk/issues/192)) ([366674f](https://github.com/open-feature/go-sdk/commit/366674fe62b9dd0921c0295a7cc47d548618b510))


### 🧹 Chore

* fix deps ([#208](https://github.com/open-feature/go-sdk/issues/208)) ([9d0f271](https://github.com/open-feature/go-sdk/commit/9d0f271eaa716a41edf02acfb3adb6d1a667e0aa))
* fixed link in readme ([dc65937](https://github.com/open-feature/go-sdk/commit/dc659372608ff6828317a648dd5fd73fd58c391e))
* fixed link to CII best pratices ([439bb02](https://github.com/open-feature/go-sdk/commit/439bb0286e89e4c493c322856215a12e0740d8ec))
* update readme for inclusion in the docs ([#193](https://github.com/open-feature/go-sdk/issues/193)) ([1152826](https://github.com/open-feature/go-sdk/commit/115282638daa0cecbb408bc04c0c99583ec3ee73))

## [1.6.0](https://github.com/open-feature/go-sdk/compare/v1.5.1...v1.6.0) (2023-07-28)


### Features

* EvaluationContext add constructor without TargetingKey ([#204](https://github.com/open-feature/go-sdk/issues/204)) ([07f4974](https://github.com/open-feature/go-sdk/commit/07f49743ec6ade051a2f5a0aea1662146048b240))


### Bug Fixes

* **deps:** update module github.com/open-feature/go-sdk-contrib/providers/flagd to v0.1.14 ([#196](https://github.com/open-feature/go-sdk/issues/196)) ([a2987b8](https://github.com/open-feature/go-sdk/commit/a2987b80569c78441b2068f5683d53feef014a1a))

## [1.5.1](https://github.com/open-feature/go-sdk/compare/v1.5.0...v1.5.1) (2023-07-18)


### Bug Fixes

* broken release process ([#199](https://github.com/open-feature/go-sdk/issues/199)) ([3990314](https://github.com/open-feature/go-sdk/commit/3990314592a672f04b466ef49e4f56e0e57cbadd))

## [1.5.0](https://github.com/open-feature/go-sdk/compare/v1.4.0...v1.5.0) (2023-07-17)


### Features

* eventing implementation ([#188](https://github.com/open-feature/go-sdk/issues/188)) ([220dc33](https://github.com/open-feature/go-sdk/commit/220dc33fbb015a6493f8d285c681761968523fa8))
* Initialize and shutdown behavior  ([#185](https://github.com/open-feature/go-sdk/issues/185)) ([609af46](https://github.com/open-feature/go-sdk/commit/609af46881a371dd1b18ce065164b9f116acdb7f))
* named client support ([#180](https://github.com/open-feature/go-sdk/issues/180)) ([c6720f9](https://github.com/open-feature/go-sdk/commit/c6720f9dbf75160438447f730a2a44ef7fd5dedf))
* provider client 1:n binding support ([#190](https://github.com/open-feature/go-sdk/issues/190)) ([940cb8b](https://github.com/open-feature/go-sdk/commit/940cb8b1a4e01304698f572831c4c26a5de32eae))


### Bug Fixes

* **deps:** update module github.com/open-feature/go-sdk-contrib/providers/flagd to v0.1.13 ([#194](https://github.com/open-feature/go-sdk/issues/194)) ([501c34b](https://github.com/open-feature/go-sdk/commit/501c34bbc9afd3f910d58e20a5992678e7fb2fe1))
* **deps:** update module golang.org/x/text to v0.10.0 ([#181](https://github.com/open-feature/go-sdk/issues/181)) ([d93f58b](https://github.com/open-feature/go-sdk/commit/d93f58bfb9ea92fbcbb0faa7cd7880bb629a6ebe))
* **deps:** update module golang.org/x/text to v0.11.0 ([#191](https://github.com/open-feature/go-sdk/issues/191)) ([713a102](https://github.com/open-feature/go-sdk/commit/713a1021a8534208046cebcebf067575af465c17))

## [1.4.0](https://github.com/open-feature/go-sdk/compare/v1.3.0...v1.4.0) (2023-05-24)


### Features

* add flag metadata field ([#178](https://github.com/open-feature/go-sdk/issues/178)) ([e3b299d](https://github.com/open-feature/go-sdk/commit/e3b299db80e036c569abd04d0e173b6e843e914d))


### Bug Fixes

* **deps:** update module github.com/go-logr/logr to v1.2.4 ([#171](https://github.com/open-feature/go-sdk/issues/171)) ([6ff22f1](https://github.com/open-feature/go-sdk/commit/6ff22f1084f5604099574c9443b0006f885c71e2))
* **deps:** update module golang.org/x/text to v0.8.0 ([#167](https://github.com/open-feature/go-sdk/issues/167)) ([33334fa](https://github.com/open-feature/go-sdk/commit/33334fa7939f5d95bec25a0228800fb43c827d80))
* **deps:** update module golang.org/x/text to v0.9.0 ([#172](https://github.com/open-feature/go-sdk/issues/172)) ([8bc9d7e](https://github.com/open-feature/go-sdk/commit/8bc9d7ef6f800b6593b0223840416a7debbc54fa))

## [1.3.0](https://github.com/open-feature/go-sdk/compare/v1.2.0...v1.3.0) (2023-03-02)


### Features

* go context in hook calls ([#163](https://github.com/open-feature/go-sdk/issues/163)) ([fc569ec](https://github.com/open-feature/go-sdk/commit/fc569ec81c0fe64779cf04ed6b8b7fd14d21b395))


### Bug Fixes

* **deps:** update module github.com/open-feature/go-sdk-contrib/providers/flagd to v0.1.5 ([#154](https://github.com/open-feature/go-sdk/issues/154)) ([ae3f3da](https://github.com/open-feature/go-sdk/commit/ae3f3da27b8b226f9f3b44a5300f4d2fba3d59df))
* **deps:** update module github.com/open-feature/go-sdk-contrib/providers/flagd to v0.1.6 ([#156](https://github.com/open-feature/go-sdk/issues/156)) ([2432c20](https://github.com/open-feature/go-sdk/commit/2432c200332559816ddc237dbb9ff8fd9b3bbfcc))
* **deps:** update module github.com/open-feature/go-sdk-contrib/providers/flagd to v0.1.7 ([#161](https://github.com/open-feature/go-sdk/issues/161)) ([cfe1d74](https://github.com/open-feature/go-sdk/commit/cfe1d7432ee96c4f18fa2d7eb5e18d502b21f693))
* **deps:** update module golang.org/x/text to v0.7.0 ([#157](https://github.com/open-feature/go-sdk/issues/157)) ([6857bb3](https://github.com/open-feature/go-sdk/commit/6857bb3c2b4fabb451aba0e04ba6fbfbbaf3f920))

## [1.2.0](https://github.com/open-feature/go-sdk/compare/v1.1.0...v1.2.0) (2023-02-02)


### ⚠ NOTE

* upgraded Go version to 1.18 ([#140](https://github.com/open-feature/go-sdk/issues/140))

### Features

* add STATIC, CACHED reasons ([#136](https://github.com/open-feature/go-sdk/issues/136)) ([ffdde63](https://github.com/open-feature/go-sdk/commit/ffdde638926e7f68837fd160c45fdc6cf1b34687))
* upgrade Go to 1.18 ([#140](https://github.com/open-feature/go-sdk/issues/140)) ([c4c3c82](https://github.com/open-feature/go-sdk/commit/c4c3c828e581bdc3a29b4ec7859e1688ad9d9554))


### Bug Fixes

* **deps:** update module github.com/open-feature/flagd to v0.3.1 ([#137](https://github.com/open-feature/go-sdk/issues/137)) ([7f2652f](https://github.com/open-feature/go-sdk/commit/7f2652fcbbf26f962a902fd85945e5093d796f16))
* **deps:** update module github.com/open-feature/flagd to v0.3.2 ([#145](https://github.com/open-feature/go-sdk/issues/145)) ([2f20979](https://github.com/open-feature/go-sdk/commit/2f20979e0c25a54710ac27759688c7824bf22429))
* **deps:** update module github.com/open-feature/flagd to v0.3.4 ([#149](https://github.com/open-feature/go-sdk/issues/149)) ([31bd8b7](https://github.com/open-feature/go-sdk/commit/31bd8b7cc73279a58cb329f4d2f16064c1115e5a))
* **deps:** update module github.com/open-feature/go-sdk-contrib/providers/flagd to v0.1.3 ([#144](https://github.com/open-feature/go-sdk/issues/144)) ([1b9fd94](https://github.com/open-feature/go-sdk/commit/1b9fd94537c95e4ef53b24c24e5dc6e63026f71e))
* **deps:** update module github.com/open-feature/go-sdk-contrib/providers/flagd to v0.1.4 ([#146](https://github.com/open-feature/go-sdk/issues/146)) ([a45f288](https://github.com/open-feature/go-sdk/commit/a45f2888493f86759fbd513d3e06480ec83c30be))
* validate that a flag key is valid UTF-8 & implemented fuzzing tests ([#141](https://github.com/open-feature/go-sdk/issues/141)) ([e3e7f82](https://github.com/open-feature/go-sdk/commit/e3e7f829c978a706297365bb72492785be09f39c))

## [1.1.0](https://github.com/open-feature/go-sdk/compare/v1.0.1...v1.1.0) (2023-01-10)


### Features

* HookContext constructor ([#130](https://github.com/open-feature/go-sdk/issues/130)) ([1701648](https://github.com/open-feature/go-sdk/commit/1701648c5f137a78d850a613db2b159f44a19f86))
* NewClientMetadata constructor  ([#133](https://github.com/open-feature/go-sdk/issues/133)) ([fa8b15b](https://github.com/open-feature/go-sdk/commit/fa8b15b4a66c1f606dcc3e631e427631ad63b8c5))


### Bug Fixes

* **deps:** update module github.com/cucumber/godog to v0.12.6 ([#121](https://github.com/open-feature/go-sdk/issues/121)) ([780d5a4](https://github.com/open-feature/go-sdk/commit/780d5a419ffbef2701d018531bbe30676d3bef4d))
* **deps:** update module golang.org/x/text to v0.6.0 ([#115](https://github.com/open-feature/go-sdk/issues/115)) ([728cd4b](https://github.com/open-feature/go-sdk/commit/728cd4bbe4e71eaf03f93edfcd1d1255a616c675))

## [1.0.1](https://github.com/open-feature/go-sdk/compare/v1.0.0...v1.0.1) (2022-12-09)


### Bug Fixes

* allow nil value for object evaluation ([f45dba0](https://github.com/open-feature/go-sdk/commit/f45dba0678eac07eda8923842bae9b15cb8b99af))
* allow nil value for object evaluation ([#118](https://github.com/open-feature/go-sdk/issues/118)) ([f45dba0](https://github.com/open-feature/go-sdk/commit/f45dba0678eac07eda8923842bae9b15cb8b99af))

## [1.0.0](https://github.com/open-feature/go-sdk/compare/v0.6.0...v1.0.0) (2022-10-19)


### Miscellaneous Chores

* release 1.0.0 ([#101](https://github.com/open-feature/go-sdk/issues/101)) ([665d670](https://github.com/open-feature/go-sdk/commit/665d6703fc39b33f0f11d3c427b479855c322c1b))

## [0.6.0](https://github.com/open-feature/go-sdk/compare/v0.5.1...v0.6.0) (2022-10-11)


### ⚠ BREAKING CHANGES

* made EvaluationContext fields unexported with a constructor and setters to enforce immutability (#91)

### Features

* made EvaluationContext fields unexported with a constructor and setters to enforce immutability ([#91](https://github.com/open-feature/go-sdk/issues/91)) ([691a1e3](https://github.com/open-feature/go-sdk/commit/691a1e360e1966280d1b03579ea5e9f03afadf94))


### Bug Fixes

* locks on singleton and client state to ensure thread safety ([#93](https://github.com/open-feature/go-sdk/issues/93)) ([9dbd6b0](https://github.com/open-feature/go-sdk/commit/9dbd6b0f13bf9b22b2dace6445051f55f8031367))
* resolution error only includes the code ([#96](https://github.com/open-feature/go-sdk/issues/96)) ([524b054](https://github.com/open-feature/go-sdk/commit/524b05478a08f17bf7892905352c1a5cf47a69a9))

## [0.5.1](https://github.com/open-feature/go-sdk/compare/v0.5.0...v0.5.1) (2022-10-03)


### Bug Fixes

* Client uses value returned by provider ([#85](https://github.com/open-feature/go-sdk/issues/85)) ([436a712](https://github.com/open-feature/go-sdk/commit/436a7129668b558eb54b80121a75ef9e4b44deba))

## [0.5.0](https://github.com/open-feature/go-sdk/compare/v0.4.0...v0.5.0) (2022-09-30)


### ⚠ BREAKING CHANGES

* changed client details signatures to return new type (#84)
* spec v0.5.0 compliance (#82)
* defined type for provider interface evaluation context (#74)
* replaced EvaluationOptions with variadic option setter in client functions (#77)
* introduced context.Context to client and provider api (#75)

### Features

* changed client details signatures to return new type ([#84](https://github.com/open-feature/go-sdk/issues/84)) ([25ecdac](https://github.com/open-feature/go-sdk/commit/25ecdacb8303f95ec88656a7f47c8bd2ef0c019a))
* introduced context.Context to client and provider api ([#75](https://github.com/open-feature/go-sdk/issues/75)) ([d850c88](https://github.com/open-feature/go-sdk/commit/d850c8873d617aec7d1013aa1c751aa5bf0dce92))
* replaced EvaluationOptions with variadic option setter in client functions ([#77](https://github.com/open-feature/go-sdk/issues/77)) ([fc4b871](https://github.com/open-feature/go-sdk/commit/fc4b8716f6d3c904b464d34176d0c6ed67f741fc))
* spec v0.5.0 compliance ([#82](https://github.com/open-feature/go-sdk/issues/82)) ([69b8f8e](https://github.com/open-feature/go-sdk/commit/69b8f8e534ad0b99bf3de67cca531720f4bfc2de))


### Bug Fixes

* add reason indicating pseudorandom split ([#76](https://github.com/open-feature/go-sdk/issues/76)) ([e843f5d](https://github.com/open-feature/go-sdk/commit/e843f5d101041e6e3ba785168b8526fcf7f50c8e))


### Code Refactoring

* defined type for provider interface evaluation context ([#74](https://github.com/open-feature/go-sdk/issues/74)) ([69988c0](https://github.com/open-feature/go-sdk/commit/69988c097f16f3aaca9bdae07ea33fbce148872d))

## [0.4.0](https://github.com/open-feature/go-sdk/compare/v0.3.0...v0.4.0) (2022-09-20)


### ⚠ BREAKING CHANGES

* rename module to go-sdk (#66)

### Features

* rename module to go-sdk ([#66](https://github.com/open-feature/go-sdk/issues/66)) ([75a901a](https://github.com/open-feature/go-sdk/commit/75a901a330ab7517e4c92def5f7bf854391203d6))


### Bug Fixes

* ensure default client logger is updated when global logger changes ([#61](https://github.com/open-feature/go-sdk/issues/61)) ([f8e2827](https://github.com/open-feature/go-sdk/commit/f8e2827639d7e7f1206de933d4ed043489eadd7d))
* return error code from client given by provider ([#67](https://github.com/open-feature/go-sdk/issues/67)) ([f0822b6](https://github.com/open-feature/go-sdk/commit/f0822b6ce9522cbbb10ed5168cecad2df6c29e40))

## [0.3.0](https://github.com/open-feature/golang-sdk/compare/v0.2.0...v0.3.0) (2022-09-14)


### ⚠ BREAKING CHANGES

* remove duplicate Value field from ResolutionDetail structs (#58)

### Bug Fixes

* remove duplicate Value field from ResolutionDetail structs ([#58](https://github.com/open-feature/golang-sdk/issues/58)) ([945bd96](https://github.com/open-feature/golang-sdk/commit/945bd96c808246614ad5a8ab846b0b530ff313cc))

## [0.2.0](https://github.com/open-feature/golang-sdk/compare/v0.1.0...v0.2.0) (2022-09-02)


### ⚠ BREAKING CHANGES

* flatten evaluationContext object (#51)

### Features

* implemented structured logging ([#54](https://github.com/open-feature/golang-sdk/issues/54)) ([04649c5](https://github.com/open-feature/golang-sdk/commit/04649c5b954531601dc3e8a474bbff66094d3b1c))
* introduce UnimplementedHook to avoid authors having to define empty functions ([#55](https://github.com/open-feature/golang-sdk/issues/55)) ([0c0bd32](https://github.com/open-feature/golang-sdk/commit/0c0bd32894346babe8d180b086362e95fd3670ef))
* remove EvaluationOptions from FeatureProvider func signatures. ([91aaeb5](https://github.com/open-feature/golang-sdk/commit/91aaeb5893a79ae7ebc9949c7c59aa72b7651e09))


### Code Refactoring

* flatten evaluationContext object ([#51](https://github.com/open-feature/golang-sdk/issues/51)) ([b8383e1](https://github.com/open-feature/golang-sdk/commit/b8383e148184c1d8e58ff74217cffc99e713d29f))
