# Changelog

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
