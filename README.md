# ng-contrib

[![Go Reference](https://pkg.go.dev/badge/github.com/foxie-io/ng-contrib.svg)](https://pkg.go.dev/github.com/foxie-io/ng-contrib)
[![Go Report Card](https://goreportcard.com/badge/github.com/foxie-io/ng-contrib)](https://goreportcard.com/report/github.com/foxie-io/ng-contrib)

`ng-contrib` is a **community-driven collection of utilities and extensions for [foxie-io/ng](https://github.com/foxie-io/ng)**.
It provides ready-to-use components like **rate limiting**, **middleware helpers**, and other tools to enhance `ng` applications.

Think of it as `fiber-contrib` but for `ng`.

---

## Features

- **Rate Limiting ([ratelimit package](https://pkg.go.dev/github.com/foxie-io/ng-contrib/ratelimit))**
  - Token Bucket
  - Fixed Window
  - Sliding Window
- **Flexible storage backends**
  - In-memory
  - Redis
- **JSON-serializable limiters** for persistence or distributed setups.
- Easy integration with any `ng` project.

---

## Installation

```bash
go get github.com/foxie-io/ng-contrib/...
```
