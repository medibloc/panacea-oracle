FROM ubuntu:focal-20220801 AS base

# Install Intel SGX drivers
# https://github.com/edgelesssys/edgelessdb/blob/c91d353d45370e68218e885e9cbea621fdedf642/Dockerfile#L45
ARG PSW_VERSION=2.17.100.3-focal1
ARG DCAP_VERSION=1.14.100.3-focal1
RUN apt-get update && apt-get install -y --no-install-recommends ca-certificates gnupg libcurl4 wget jq \
  && wget -qO- https://download.01.org/intel-sgx/sgx_repo/ubuntu/intel-sgx-deb.key | apt-key add \
  && echo 'deb [arch=amd64] https://download.01.org/intel-sgx/sgx_repo/ubuntu focal main' >> /etc/apt/sources.list \
  && wget -qO- https://packages.microsoft.com/keys/microsoft.asc | apt-key add \
  && echo 'deb [arch=amd64] https://packages.microsoft.com/ubuntu/20.04/prod focal main' >> /etc/apt/sources.list \
  && apt-get update && apt-get install -y --no-install-recommends \
  libsgx-ae-id-enclave=$DCAP_VERSION \
  libsgx-ae-pce=$PSW_VERSION \
  libsgx-ae-qe3=$DCAP_VERSION \
  libsgx-dcap-ql=$DCAP_VERSION \
  libsgx-enclave-common=$PSW_VERSION \
  libsgx-launch=$PSW_VERSION \
  libsgx-pce-logic=$DCAP_VERSION \
  libsgx-qe3-logic=$DCAP_VERSION \
  libsgx-urts=$PSW_VERSION \
  && apt-get install -y az-dcap-client

# Install EGo
# https://docs.edgeless.systems/ego/#/getting-started/install?id=install-the-deb-package
RUN wget https://github.com/edgelesssys/ego/releases/download/v1.0.1/ego_1.0.1_amd64.deb \
  && apt install ./ego_1.0.1_amd64.deb

####################################################

FROM base AS build

# Install prerequisites
RUN apt-get update && apt-get install -y --no-install-recommends git build-essential

# Build oracled
WORKDIR /src
COPY . .
RUN make build
RUN ego sign ./scripts/enclave-prod.json

####################################################

FROM base

COPY --from=build /src/build/oracled /usr/bin/oracled
RUN chmod +x /usr/bin/oracled

EXPOSE 8080

CMD ["ego", "run", "/usr/bin/oracled"]
