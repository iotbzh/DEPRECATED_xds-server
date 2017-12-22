# SDKs management scripts

To support a new SDK family, you must create a new directory under
`scripts/sdk/xxx` where xxx is the new SDK family.

Then you must create the following scripts (or executable) :

- `get-config`: returned SDK configuration structure
- `list`: returned the list of installed SDKs
- `add`: add a new SDK
- `remove`: remove an existing SDK

## `get-config`

Returned SDK configuration as json format:

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

## `list`

Returned the list all SDKs (available and installed)

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

## `add`

add a new SDK

List of parameters to implement:

- `-f|--file <filepath>` :  install a SDK using a file
- `--force`:                force SDK install when a SDK already in the same destination directory
- `-u|--url <url>` :        download SDK using this URL and then install it
- `-no-clean` :             don't cleanup temporary files
- `-h|--help` :             display help

## `remove`

Remove an existing SDK

The first argument is the full path of the directory of the SDK to removed.
