# Installation: Build from source

NOTE: This installation process works only on the hardwares specified in the [Hardware Requirements](../README.md#hardware-requirements)


## Prerequisites

```bash
sudo apt update
sudo apt install build-essential libssl-dev

sudo snap install go --classic
sudo snap install ego-dev --classic
sudo ego install az-dcap-client

sudo usermod -a -G sgx_prv $USER
# After this, reopen the shell so that the updated Linux group info can be loaded.
```

Also, please add the following lines to your shell profile file, such as `~/.profile`.
```bash
unset SGX_AESM_ADDR
export AZDCAP_DEBUG_LOG_LEVEL=INFO
```
If the `SGX_AESM_ADDR` is set, you will face the following error when running some SGX-related operations.
```
ERROR: sgxquoteexprovider: failed to load libsgx_quote_ex.so.1: libsgx_quote_ex.so.1: cannot open shared object file: No such file or directory [openenclave-src/host/sgx/linux/sgxquoteexloader.c:oe_sgx_load_quote_ex_library:118]
ERROR: Failed to load SGX quote-ex library (oe_result_t=OE_QUOTE_LIBRARY_LOAD_ERROR) [openenclave-src/host/sgx/sgxquote.c:oe_sgx_qe_get_target_info:688]
ERROR: SGX Plugin _get_report(): failed to get ecdsa report. OE_QUOTE_LIBRARY_LOAD_ERROR (oe_result_t=OE_QUOTE_LIBRARY_LOAD_ERROR) [openenclave-src/enclave/sgx/attester.c:_get_report:320]
```


## Build a `oracled` binary

```bash
# in SGX-enabled environment,
make build

# in SGX-disabled environment,
GO=go make build
```


## Run unit tests

```bash
# in SGX-enabled environment,
make test

# in SGX-disabled environment,
GO=go make test
```


## Sign the `oracled` with Ego

To run the binary in the enclave, the binary must be signed with EGo.

### For development

First of all, prepare a RSA private key of the signer.

```bash
openssl genrsa -out private.pem -3 3072
openssl rsa -in private.pem -pubout -out public.pem
```

After that, prepare a `enclave.json` file that will be applied when signing the binary with EGo.

In the following example, please replace the `<a-directory-you-want>` with a directory in your local.
This will be mounted as `/home_mnt` to the file system presented to the enclave.
And, the `HOME` environment variable will indicate the `/home_mnt`.

```json
{
  "exe": "./build/oracled",
  "key": "private.pem",
  "debug": true,
  "heapSize": 512,
  "executableHeap": false,
  "productID": 1,
  "securityVersion": 1,
  "mounts": [
    {
      "source": "<a-directory-you-want>",
      "target": "/home_mnt",
      "type": "hostfs",
      "readOnly": false
    },
    {
      "target": "/tmp",
      "type": "memfs"
    }
  ],
  "env": [
    {
      "name": "HOME",
      "value": "/home_mnt"
    }
  ],
  "files": null
}
```

Finally, you can sign the binary with EGo.

```bash
ego sign ./enclave.json
```

If the binary is signed successfully, you can move the binary to where you want.

### For production

A configuration for production is already prepared in the [`scripts/enclave-prod.json`](../scripts/enclave-prod.json).

So, you can just put your RSA private key (`private.pem`) into the `scripts/`, and run the following command to sign the binary.

```bash
ego sign ./scripts/enclave-prod.json
```

If the binary is signed successfully, you can move the binary to where you want, or publish the binary to GitHub or so.

Note that a `/oracle` directory must be created and its permissions must be set properly before running the binary,
because the `/oracle` directory will be mounted as a `HOME` directory to the enclave.
For more details, please see the [`scripts/enclave-prod.json`](../scripts/enclave-prod.json).
