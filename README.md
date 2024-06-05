![cover](./cover.svg)

## Usage
**A simple example:**

```go
package main

import (
    "context"

    "github.com/gaylatea/instrument"
)

func main() {
    ctx := context.Background()
    instrument.Infof(ctx, "Hello!")
}
```

### Logs

### Traces

### Raw events

## Sinks
### Console

## Example
A sample program that uses all available features: [example/main.go](./example/main.go)

![Example instrument output](./example.gif)

## License
This work is available under the terms of the [CC0-1.0 License](./LICENSE).

[![_Made with Love_](https://img.shields.io/badge/Made%20with%20love-32373B.svg?style=for-the-badge&logo=data:image/svg+xml;base64,PD94bWwgdmVyc2lvbj0iMS4wIiBlbmNvZGluZz0idXRmLTgiPz48IS0tIFVwbG9hZGVkIHRvOiBTVkcgUmVwbywgd3d3LnN2Z3JlcG8uY29tLCBHZW5lcmF0b3I6IFNWRyBSZXBvIE1peGVyIFRvb2xzIC0tPgo8c3ZnIHdpZHRoPSI4MDBweCIgaGVpZ2h0PSI4MDBweCIgdmlld0JveD0iMCAwIDY0IDY0IiB4bWxucz0iaHR0cDovL3d3dy53My5vcmcvMjAwMC9zdmciIHhtbG5zOnhsaW5rPSJodHRwOi8vd3d3LnczLm9yZy8xOTk5L3hsaW5rIiBhcmlhLWhpZGRlbj0idHJ1ZSIgcm9sZT0iaW1nIiBjbGFzcz0iaWNvbmlmeSBpY29uaWZ5LS1lbW9qaW9uZSIgcHJlc2VydmVBc3BlY3RSYXRpbz0ieE1pZFlNaWQgbWVldCI+PHBhdGggZD0iTTYxLjEgMTguMmMtNi40LTE3LTI3LjItOS40LTI5LjEtLjljLTIuNi05LTIyLjktMTUuNy0yOS4xLjlDLTQgMzYuNyAyOS42IDUzLjMgMzIgNTZjMi40LTIuMiAzNi0xOS42IDI5LjEtMzcuOCIgZmlsbD0iI2Y1Y2UzZSI+PC9wYXRoPjwvc3ZnPg==)](https://tfw.computer/systems/made-with-love)
