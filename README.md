## mock-platform

Define platform, plugin running port in `plugins.json`, run gateway.

Make sure platform login function uses the auth URL of this mock gateway.

Download executable gateway from issue #2.

#### Using scripts

```
# Example
cd scripts
./tw-login wizawu
./tw-curl wizawu /comment/thread/900000000000000001
./tw-curl wizawu /comment/comment/1000000000000000021 -X DELETE
```

#### Reference

+ [Google Groups] (https://groups.google.com/forum/#!topic/golang-nuts/KBx9pDlvFOc)
