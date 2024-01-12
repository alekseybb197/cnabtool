# cnabtool
The tool for manipulating cnab artifacts.

## Build

`make`

`cp bin/cnabtool /usr/local/bin`

## Setup

`mkdir ~/.cnabtool`

`cat <<EOT >~/.cnabtool/config.yaml`

`credentials:`

`  username: "username"`

`  password: "password"`

`timeout: 10000`

`verbosity: 2`

`EOT`

## Use

### Get manifest

`cnabtool content manifest registry.example.com/project/cnab:tag`

### Inspect CNAB project

`cnabtool content inspect registry.example.com/project/cnab:tag`

### Delete CNAB

`cnabtool content delete registry.example.com/project/cnab:tag`
