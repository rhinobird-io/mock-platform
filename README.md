## mock-platform

Define platform, plugin running host and port in `plugins.json`, run gateway.

Make sure platform login function uses the auth URL of this mock gateway, http://hostname:8080/auth.

Download executable gateway from issue #2.

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
