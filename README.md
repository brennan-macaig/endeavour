# Endeavour
<img src="https://upload.wikimedia.org/wikipedia/commons/7/7b/Space_Shuttle_Endeavour_launches_on_STS-99.jpg" width="200">

_The last NASA-produced manned vehicle to bring things to space._

A lightweight Go program to upload files to Nexus in the raw repository format.

## Usage
Endeavour is a command line tool only, written for unix systems (although it should be compatible with Windows).
Generally, the command is invoked as:
```bash
$ endeavour [options] file-or-dir0 ... file-or-dirN
```

Where `file-or-dir0 ... file-or-dirN` is any number of files or directories to upload.

Nexus credentials (username and password) must be specified in environment variables, which default to `REPO_USERNAME`,
`REPO_PASSWORD` respectively. Typically, for CI/CD this is done in Group settings.

If you are setting environment variables manually via the terminal, please do so in a secure manner, eg:

```bash
$ read -s REPO_USERNAME
<enter username>
$ export REPO_USERNAME
$ read -s REPO_PASSWORD
<enter password>
$ export REPO_PASSWORD
```

### Command Line Flags

| endeavour flag | raw-publisher flag | action |
|:---------------|:-------------------|:-------|
|`-h`|Display the help message
|`-r`|Set the repository on Nexus to upload to|
|`-U`|Set the URL to upload to if different than the default from the build (rare for from-source builds)|
|`-P`|Set the path inside the repository to upload to, typically `appname/${VERSION}`|
|`-user-var`|Set the environment variable that the Nexus username is stored in, if not `REPO_USERNAME`|
|`-pass-var`|Set the environment variable that the Nexus password is stored in, if not `REPO_PASSWORD`|
|`-v`|Enable verbose logging output|
|`-no-art`|Supress art from being printed to neaten logs|
|`-version`|Print the version and exit.

### Typical Use
Typically, when uploading a file, the following will be the command structure:

```bash
$ endeavour -r "top-level-nexus-repository" -P "projectname/version#" file-or-dir0 file-or-dir1 ... file-or-dirN
```

### Packaging
Endeavour is automatically packaged as a `.deb` and `.rpm`.

## FAQ

### What is endeavour?
Endeavour is a lightweight raw-repository writer for Nexus written in Go, which serves to upload
any content to the Nexus repository manager in a raw repository format. Endeavour accomplishes this
through HTTP PUT with basic auth.

### I found a bug with endeavour, what do I do?
Feel free to make a merge request!
Otherwise, submit an issue on the github issue tracker.

### I dislike endeavour, or feel it needs to be improved
Feel free to make a merge request!
Otherwise, submit an issue on the github issue tracker.
Feel free to message me on github/twitter/wherever otherwise.


### What gets printed in the console when endeavour is done?
It's art, see [ART.md](ART.md) for more information.

#### I don't like it, or it clutters my logs
Add `-no-art` to your endeavour call to make your CI logs boring and artless.
