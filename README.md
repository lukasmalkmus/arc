# LukasMa/arc (WIP)
> A complete toolkit for the ARC assembly language. - by **[Lukas Malkmus](https://github.com/LukasMa)**

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
  - [ ] **Vet invalid ARC source code**

### Usage
#### Installation
The easiest way to run the *arc* tool is by grabbing the latest binary from
the [release page][release].

##### Building from source
This project uses [glide](http://glide.sh) for vendoring.
```bash
git clone https://github.com/LukasMa/arc
cd arc
glide install
go build cmd/arc/main.go
```

#### Usage
Just run `arc` in your terminal to get some helpful usage advice.

```bash
arc --help
```

### Contributing
Feel free to submit PRs or to fill Issues. Every kind of help is appreciated.

### License
© Lukas Malkmus, 2017

Distributed under Apache License (`Apache License, Version 2.0`).

See [LICENSE](LICENSE) for more information.


[travis]: https://travis-ci.org/LukasMa/arc
[travis_badge]: https://travis-ci.org/LukasMa/arc.svg
[coverage]: https://coveralls.io/github/LukasMa/arc?branch=master
[coverage_badge]: https://coveralls.io/repos/github/LukasMa/arc/badge.svg?branch=master
[report]: https://goreportcard.com/report/github.com/LukasMa/arc
[report_badge]: https://goreportcard.com/badge/github.com/LukasMa/arc
[docs]: https://godoc.org/github.com/LukasMa/arc
[docs_badge]: https://godoc.org/github.com/LukasMa/arc?status.svg
[release]: https://github.com/LukasMa/arc/releases
[release_badge]: https://img.shields.io/github/release/LukasMa/arc.svg
[license]: https://opensource.org/licenses/Apache-2.0
[license_badge]: https://img.shields.io/badge/license-Apache-blue.svg
