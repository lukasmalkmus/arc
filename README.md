# lukasmalkmus/arc (WIP)

> A complete toolkit for the ARC assembly language. - by **[Lukas Malkmus](https://github.com/lukasmalkmus)**

[![Travis Status][travis_badge]][travis]
[![Coverage Status][coverage_badge]][coverage]
[![Go Report][report_badge]][report]
[![GoDoc][docs_badge]][docs]
[![Latest Release][release_badge]][release]
[![License][license_badge]][license]

---

## DISCLAIMER

This project idled in my GOPATH, so I decided to release it. It is in a WIP
state, some features are not working and some tests not passing. I haven't
worked on it for a while and don't plan to further develop it. Today I would
also do a lot of things differently.

If you are interested what currently works, checkout out the passing tests.

## Table of Contents

1. [Introduction](#introduction)
1. [Features](#features)
1. [Usage](#usage)
1. [Contributing](#contributing)
1. [License](#license)

### Introduction

The *arc* tool is a simple command line application which provides powerful
features that make working with ARC source code a breeze.

### Features

- [x] **Assembling**
- [x] **Checking (Vet)**
- [x] **Formating**
- [x] **Parsing**

#### Todo

- [ ] **Format invalid ARC source code**

### Usage

#### Installation

The easiest way to run the *arc* tool is by grabbing the latest binary from
the [release page][release].

##### Using go get

If go is installed on your system, installing arc can be accomplished by
utilizing `go get`:

```bash
go get -u -d github.com/lukasmalkmus/arc/cmd/...
```

##### Building from source

This project uses [dep](https://github.com/golang/dep) for vendoring.

```bash
git clone https://github.com/lukasmalkmus/arc
cd arc
dep ensure
go install ./... # or make
```

#### How to use

Just run `arc` in your terminal to get some helpful advice.

```bash
arc --help
```

### Contributing

Feel free to submit PRs or to fill Issues. Every kind of help is appreciated.

### License

© Lukas Malkmus, 2018

Distributed under MIT License (`The MIT License`).

See [LICENSE](LICENSE) for more information.

[travis]: https://travis-ci.org/lukasmalkmus/arc
[travis_badge]: https://travis-ci.org/lukasmalkmus/arc.svg
[coverage]: https://coveralls.io/github/lukasmalkmus/arc?branch=master
[coverage_badge]: https://coveralls.io/repos/github/lukasmalkmus/arc/badge.svg?branch=master
[report]: https://goreportcard.com/report/github.com/lukasmalkmus/arc
[report_badge]: https://goreportcard.com/badge/github.com/lukasmalkmus/arc
[docs]: https://godoc.org/github.com/lukasmalkmus/arc
[docs_badge]: https://godoc.org/github.com/lukasmalkmus/arc?status.svg
[release]: https://github.com/lukasmalkmus/arc/releases
[release_badge]: https://img.shields.io/github/release/lukasmalkmus/arc.svg
[license]: https://opensource.org/licenses/MIT
[license_badge]: https://img.shields.io/badge/license-MIT-blue.svg