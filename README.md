## mock-platform

### Gateway only

If you want to use the mock-gateway only, download the [latest release](https://github.com/rhinobird-io/mock-platform/releases) of the binary and run it.

Then put `plugins.json` aside the binary file, following the format of the [sample file](https://github.com/rhinobird-io/mock-platform/blob/master/plugins.json.sample)


### Prerequisite

* Docker
* Define platform, plugin running host and port in `plugins.json`.
* Set environment variables such as GITLAB_PRIVATE_TOKEN in `.env`.

You can get the gitlab private token from gitlab profile account setting.

### Start gateway and platform

```
$ scripts/start.sh
```

This script needs you to input password to execute sudo commands.

### Start bash shell inside running container

```
$ scripts/shell.sh tw-platform
```

### Using scripts

```
# Example
cd scripts
./tw-login wizawu
./tw-curl wizawu /comment/thread/900000000000000001
./tw-curl wizawu /comment/comment/1000000000000000021 -X DELETE
```

### Reference

+ [Google Groups - How to proxy Websocket in Golang] (https://groups.google.com/forum/#!topic/golang-nuts/KBx9pDlvFOc)
