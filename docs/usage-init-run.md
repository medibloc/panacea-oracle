# Usage: Initialize and run the oracle

NOTE: Before running the `oracled`, please read the [installation](../README.md#installation) guides carefully.


## Preparations

Create an empty directory on your host, to be mounted as the `/home_mnt` directory in the enclave.
The `oracled` process in the enclave will recognize the `/home_mnt` directory as its `$HOME`.
```bash
# If you run oracle on your host
sudo mkdir /oracle

# If you run oracle using Docker
mkdir $(pwd)/oracle
```
If you need more details about the home directory, please see the [installation](./installation-src.md#for-production) guide.

Then, if you want to run the `oracled` using Docker as described in the [installation](./installation-docker.md) guide, it is recommended to create an environment variable that you can execute the Docker container easily.
```bash
export DOCKER_CMD="docker run --rm --device /dev/sgx_enclave --device /dev/sgx_provision -v $(pwd)/oracle:/oracle ghcr.io/medibloc/panacea-oracle:latest"
```
Even if you are not going to use Docker, it would be easier to follow the instruction below if you set the environment variable as an empty string.
```bash
export DOCKER_CMD=""
```


## Initialize an app dir for the `oracled`

```bash
$DOCKER_CMD ego run oracled init
```
By default, the app dir is generated as `$HOME/.oracle` in the enclave.
It means that you can also find the generated app dir from your host (e.g. `/oracle/.oracle` or `$(pwd)/oracle/.oracle`).

## Generate an oracle key

NOTE: This step must be executed only by the first (genesis) oracle.

```bash
$DOCKER_CMD ego run oracled gen-oracle-key \
    --trusted-block-height <block-height> \
    --trusted-block-hash <block-hash>
```
Then, two files are generated under the home directory:
- `oracle_priv_key.sealed` : sealed oracle private key
- `oracle_pub_key.json` : oracle public key & its remote report


## Verify the remote report

You can verify that oracle key is generated in SGX using the promised binary.
For that, the public key and its remote report are required.

```json
{
  "public_key_base64" : "<base64-encoded-public-key>",
  "remote_report_base64": "<base64-encoded-remote-report>"
}
```

Then, you can verify the remote report.
```bash
$DOCKER_CMD ego run oracled verify-report <remote-report-path>
```

## Register an oracle to the Panacea

Request to register an oracle.

The trusted block information is required which would be used for light client verification.
The account number and index are optional with the default value of 0.

```bash
$DOCKER_CMD ego run oracled register-oracle \
    --trusted-block-height <block-height> \
    --trusted-block-hash <block-hash>
```

## Get the oracle key registered in the Panacea

If an oracle registered successfully (vote for oracle registration is passed), the oracle can be shared the oracle private key.
The oracle private key is encrypted and shared, and it can only be decrypted using the node private key (which is used when registering oracle) 

```bash
$DOCKER_CMD ego run oracled get-oracle-key
```

The oracle private key is sealed and stored in a file named `oracle_priv_key.sealed` under `$HOME/.oracle/` in the enclave.
