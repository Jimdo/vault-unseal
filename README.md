# Jimdo / vault-unseal

[![License: Apache v2.0](https://badge.luzifer.io/v1/badge?color=5d79b5&title=license&text=Apache+v2.0)](http://www.apache.org/licenses/LICENSE-2.0)

This small utility is a helper to automatically unlock a [Vault](https://www.vaultproject.io/) instance by having an amount of servers having access to one or multiple tokens.

## Features

- Provide one or multiple tokens for the unseal command
- `vault-unseal` does check whether the vault instance is locked and tries to unlock if it is locked

## Usage

```bash
# ./vault-unseal --help
Usage of ./vault-unseal:
      --instance="http://127.0.0.1:8200": Vault instance to unlock
  -1, --oneshot[=true]: Only try once and exit after
      --sleep=30: How long to wait between sealed-state checks
      --tokens="": Tokens to try for unsealing the vault instance
```
