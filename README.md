## mock-platform

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

+ [Google Groups] (https://groups.google.com/forum/#!topic/golang-nuts/KBx9pDlvFOc)
