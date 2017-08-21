# lukasmalkmus/arc (WIP)
> A complete toolkit for the ARC assembly language. - by **[Lukas Malkmus](https://github.com/lukasmalkmus)**

[![Travis Status][travis_badge]][travis]
[![Coverage Status][coverage_badge]][coverage]
[![Go Report][report_badge]][report]
[![GoDoc][docs_badge]][docs]
[![Latest Release][release_badge]][release]
[![License][license_badge]][license]

---

## Table of Contents
1. [Introduction](#introduction)
2. [Features](#features)
3. [Usage](#usage)
4. [Contributing](#contributing)
5. [License](#license)

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

#### Usage
Just run `arc` in your terminal to get some helpful advice.

```bash
arc --help
```

### Contributing
Feel free to submit PRs or to fill Issues. Every kind of help is appreciated.

### License
© Lukas Malkmus, 2017

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