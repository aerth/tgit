# tgit (tiny git)

### abilities:

  * can be static linked
  * easy to compile ([Go](https://golang.org))
  * easy to install ([Binary Releases](https://github.com/aerth/tgit/releases/latest))
  * can be symlinked as `git` for `go get`

### current commands

  * clone (https only)

### for when you cant use git because of:

  * no c compiler
  * various other reasons

### minimal dependencies:

in a nutshell, your `tar` needs to be able to understand the following `tar magic`:

```
tar -v -C [tempdir] -x -z -f [downloaded-tarball]
```

and your `mv` should work

```
mv [directory] [destination]

```
