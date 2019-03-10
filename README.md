# goPacked
[![License](http://img.shields.io/:license-agpl3-blue.svg?style=flat-square)](http://www.gnu.org/licenses/agpl-3.0.html)

goPacked is a simple text-based Minecraft modpack manager. It uses a fairly simple JSON pack format.

## Installing
### Precompiled
There are compiled versions available for Linux, macOS and Windows. The latest version is always available from the following links and in the [releases](https://github.com/tulir/gopacked/releases) section.
* Debian package: https://dl.maunium.net/programs/gopacked.deb
* Other Linuxes: https://dl.maunium.net/programs/gopacked_linux.tar.gz
* macOS: https://dl.maunium.net/programs/gopacked_macos.zip
* Windows: https://dl.maunium.net/programs/gopacked_windows.zip

### Self-compiled
0. Install [Go](https://golang.org/) 1.11+
1. Clone the repository
2. Run `make`

If you don't want to use `make`, you can also manually `go install` or `go build` in the `cmd/gopacked` or `cmd/twitchparse` subdirectories.

## Usage
Basic usage: `gopacked [-h] [-p PATH] [-m PATH] <ACTION> <URL/NAME>`

### Flags
`-h, --help` - Show an argument help page.

`-p, --path` - The install path (game directory) to use.

`-m, --minecraft` - The minecraft directory location (defaults to $home/.minecraft on Linux, $home/Library/Application Support/minecraft on Mac OS X and %APPDATA%/.minecraft on Windows)

### Actions
`install` - Install the goPack from the given goPack definition URL.

`update` - Update a goPack. You must either provide the modpack path with `-p`, the goPack definition URL or the pack name. If you only provide the goPack definition URL or the pack name, the pack must be installed in the default location (`.minecraft/gopacked/<simplename>`)

`uninstall` - Uninstall a goPack. Same arguments as `update`.

## Creating a goPack
[The pack I created goPacked for](https://maunium.net/ventornamodpilerna/modpack.json) can be used as an example.

### Format base
The JSON base must contain a name, simple name, update URL, author and version. The base must also contain two file entries. "mcl-version" is saved into .minecraft/versions and "files" is saved into the modpacks game directory.
The base may contain a profile settings block which contains the non-default settings to insert (as-is) into the modpack profile in Minecraft's launcher_profiles.json.

```json
{
  "name": "Example Modpack",
  "simplename": "examplepack",
  "update-url": "http://example.com/examplemodpack",
  "author": "John Doe",
  "version": "1.0.1.0",
  "profile-settings": {
    "javaArgs": "-Xmx2G",
    "anotherMCProfileArg": false
  },
  "mcl-version": {
    // A directory entry. Will be installed to .minecraft/versions/<simplename>
  },
  "files": {
    // A directory entry. Will be installed to the modpack game directory.
  }
}
```

### File entries
A file entry is a JSON object with at least the type of the entry. All file entries are parsed as equal, but some fields may be ignored when processing depending on the type of the file entry. The possible file entry fields are as follows:
 * `type` - Identifies the type of the file entry. Allowed types:
  * `directory`
  * `file`
  * `zip-archive` (upcoming)
 * `filename` - The name to save the file to. Affects all types, will determine the unarchive directory name for archives.
 * `version` - The version of the file. Ignored by directories, used for comparison of other types for updating/downgrading.
 * `url` - The URL to download the file from. Ignored by directories.
 * `children` - A map of file entries. Ignored by everything but directories.

The display name of the file is the name of the JSON object, but the filesystem name can be overriden using the filename field
If a custom filename is set, the filename is either the JSON object name (for directories) or the final part of the URL (for files)

#### Examples
Here's an example of a file entry that doesn't have the filename field set. This file would be saved as "example.jar" by goPacked.
```json
"A file entry": {
  "type": "file",
  "version": "1.2.3.4",
  "url": "https://example.com/files/example.jar"
}
```

Directory entries usually contain children. They can be any kind of file entries. This directory would be named "exampledir", since the filename field is set.
```json
"A directory entry": {
  "type": "directory",
  "filename": "exampledir",
  "children": {
    // File entries that should be inside this directory go here
  }
}
```

### Version format
All version numbers must contain no more or less than four integers separated by dots. This is due to the fact that a lot of mods have different kinds of versioning styles and it's easiest just to have the modpack manager convert them into an universal style. I have found that nearly all mods can be fairly easily fitted into a four-number version style without any data loss.
