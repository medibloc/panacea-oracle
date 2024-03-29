# Panacea Oracle

An oracle which validates off-chain data to be transacted in the data exchange protocol of the Panacea chain while preserving privacy.

## Features

- Validating that data meets the requirements of a specific deal
    - with utilizing TEE (Trusted Execution Environment) for preserving privacy
- Providing encrypted data to buyers


## Hardware Requirements

The oracle only works on [SGX](https://www.intel.com/content/www/us/en/developer/tools/software-guard-extensions/overview.html)-[FLC](https://github.com/intel/linux-sgx/blob/master/psw/ae/ref_le/ref_le.md) environment with a [quote provider](https://docs.edgeless.systems/ego/#/reference/attest) installed.
You can check if your hardware supports SGX and it is enabled in the BIOS by following [EGo guide](https://docs.edgeless.systems/ego/#/getting-started/troubleshoot?id=hardware).


## Installation

- [Build from source](./docs/installation-src.md)
- [Use Docker](./docs/installation-docker.md)


## Usages

- [Initialize and run the oracle](./docs/usage-init-run.md)


## Recommended configurations

Recommend appropriate settings based on your application
* If the request body(receive message) is 500 KB or less in size
    * max-connections: 100
* If the request body(receive message) is between 500 KB and 1 MB in size
    * max-connections: 50


The default settings are as follows
* enclave
  * heap-size: 1024 MB
* gRPC
  * max-connections: 50
  * max-rcv-msg-size: 1024 KB
* API
  * max-connections: 50
  * max-request-body-size: 1024 KB
