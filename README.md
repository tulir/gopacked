# goPacked
goPacked is a simple text-based Minecraft modpack manager. It uses a fairly simple JSON pack format.

## Creating a goPack
### Format base
The JSON base must contain a name, simple name, update URL, author and version. The base must also contain two file entries. "mcl-version" is saved into .minecraft/versions and "files" is saved into the modpacks game directory.
The base may contain a profile settings block which contains the non-default settings to insert into the modpack profile in Minecraft's launcher_profiles.json.

```
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
    // A directory entry. Will be installed to the modpack game directory, by default .minecraft/gopacked/<simplename>
  }
}
```

### File entries
A file entry is a JSON object with at least the type of the entry. The required fields depend on the type of the entry.
Currently supported types are file and directory. Archive support is coming soon™.

The display name of the file is the name of the JSON object, but the filesystem name can be overriden using the filename field
If a custom filename is set, the filename is either the JSON object name (for directories) or the final part of the URL (for files)

Here's an example of a file entry that doesn't have the filename field set. This file would be saved as "example.jar" by goPacked.
```
"A file entry": {
  "type": "file",
  "version": "1.2.3.4",
  "url": "https://example.com/files/example.jar"
}
```

Directory entries usually contain children. They can be any kind of file entries. This directory would be named "exampledir", since the filename field is set.
```
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