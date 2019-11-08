# Shimmering Bee: UNPI

[![license](https://img.shields.io/github/license/shimmeringbee/unpi.svg)](https://github.com/shimmeringbee/unpi/blob/master/LICENSE)
[![standard-readme compliant](https://img.shields.io/badge/standard--readme-OK-green.svg)](https://github.com/RichardLitt/standard-readme)
[![Actions Status](https://github.com/shimmeringbee/unpi/workflows/test/badge.svg)](https://github.com/shimmeringbee/unpi/actions)

> Implementation of Texas Instruments Unified Network Processor Interface frame protocol in Go.

## Table of Contents

- [Background](#background)
- [Install](#install)
- [Usage](#usage)
- [Maintainers](#maintainers)
- [Contributing](#contributing)
- [License](#license)

## Background

UNPI is used by Texas Instruments for communicating with many of their Network Processor microcontrollers, specifically
this library was written to support the CC253X series of Zigbee sniffers flashed with the 
[zigbee2mqtt](https://www.zigbee2mqtt.io/getting_started/flashing_the_cc2531.html) Z-Stack coordinator firmware.

More information about UNPI is available from [Texas Instruments](http://processors.wiki.ti.com/index.php/Unified_Network_Processor_Interface) directly.

[Another implementation](https://github.com/dyrkin/unp-go/) of UNPI exists for Golang, however it holds [no licence and 
communication attempts](https://github.com/dyrkin/zigbee-steward/issues/1) with the author have failed. This is a
complete reimplementation of the library, however it is likely there will be strong coincidences due to Golang standards.

## Install

Add an import and most IDEs will `go get` automatically, if it doesn't `go build` will fetch.

```go
import "github.com/shimmeringbee/unpi"
```

## Usage

### Writing

```go
serialPort := // UART port providing a Writer

// Construct a SysReset
frame := &Frame{
    MessageType: AREQ,
    Subsystem:   SYS,
    CommandID:   0x00,
    Payload:     []byte { 0x00 },
}

// Send Frame to CC253X, blocking operation
err := unpi.Write(serialPort, frame)
```

### Reading

```go
serialPort := // UART port providing a Reader

// Read from CC253X, blocking operation
frame, err := unpi.Read(serialPort)

if err != nil {
    // Handle Error
}

// Use frame
fmt.Printf("%+v\n", frame)
```

## Maintainers

[@pwood](https://github.com/pwood)

## Contributing

Feel free to dive in! [Open an issue](https://github.com/shimmeringbee/unpi/issues/new) or submit PRs.

All Shimmering Bee projects follow the [Contributor Covenant](http://contributor-covenant.org/version/1/3/0/) Code of Conduct.

## License

   Copyright 2019 Peter Wood

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.