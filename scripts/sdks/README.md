# SDKs management scripts

To support a new SDK family, you must create a new directory under
`scripts/sdk/xxx` where xxx is the new SDK family.

Then you must create the following scripts (or executable) :

- `add`: add a new SDK
- `db-dump`: returned the list of available SDKs (JSON format)
- `db-update`: update SDKs database
- `get-family-config`: returned SDK family configuration structure (JSON format)
- `get-sdk-info`: extract SDK info (JSON format) from a SDK file/tarball
- `remove`: remove an existing SDK

## `add`

add a new SDK

This script returns code 0 when sdk is successfully installed, else returns an
error code.

List of parameters to implement:

- `-f|--file <filepath>` :  install a SDK using a file
- `--force`:                force SDK install when a SDK already in the same destination directory
- `-u|--url <url>` :        download SDK using this URL and then install it
- `-no-clean` :             don't cleanup temporary files
- `-h|--help` :             display help

## `db-dump`

Returned the list all SDKs (available and installed) using JSON format.

```json
[
  {
    "name":         "My SDK name",
    "description":  "A description",
    "profile":      "profile",
    "version":      "version",
    "arch":         "architecture",
    "path":         "path where sdk installed locally",
    "url":          "https://website.url.to.download.sdk",
    "status":       "Not Installed | Installed",
    "date":         "2017-12-25 00:00",
    "size":         "123 MB",
    "md5sum":       "123456789",
    "setupFile":    "path to file to setup SDK environment"
  }, {
    "name":         "My SDK name 2",
    "description":  "A description 2",
    ...
  }
  ...
]
```

## `db-update`

Update sdk database that may be used by `list` command.

## `get-family-config`

Returned SDK configuration as JSON format:

```json
{
    "familyName": "xxx",
    "description": "bla bla",
    "rootDir": "/yyy/zzz",
    "envSetupFilename": "my-envfilename*",
    "scriptsDir": "scripts_path"
}
```

where:

- `familyName` : sdk familyName (usually same name used as xxx directory)
- `rootDir` : root directory where SDK are/will be  installed
- `envSetupFilename` : sdk files (present in each sdk) that will be sourced to
  setup sdk environment

## `get-sdk-info`

Extract SDK info, such as name, version, ... from a SDK tarball file (when
--file option is set) or from a url (when --url option is set).

This script may also be used to check that a SDK tarball file is correct in
order to determine for example that the SDK family.

List of parameters to implement:

- `-f|--file <filepath>` :  SDK tarball file used to get SDK info
- `-u|--url <url>` :        url link used to get SDK info
- `--md5` :                 md5sum value used to validate SDK tarball file
- `-h|--help` :             display help

This script returns an error (value different from 0) and potential print an
error message. Else when info are successfully extracted, this script must
returned the following JSON structure:

```json
{
    "name":         "My SDK name",
    "description":  "A description",
    "profile":      "profile",
    "version":      "version",
    "arch":         "architecture",
    "path":         "",
    "url":          "https://website.url.to.download.sdk",
    "status":       "Not Installed",
    "date":         "2017-12-25 00:00",
    "size":         "123 MB",
    "md5sum":       "123456789",
    "setupFile":    "path to file to setup SDK environment"
}
```

## `remove`

Remove an existing SDK

The first argument is the full path of the directory of the SDK to removed.
